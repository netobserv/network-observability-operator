/*
 * Copyright (C) 2021 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package informers

import (
	"fmt"
	"net"
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/transform/kubernetes/cni"
	"github.com/netobserv/flowlogs-pipeline/pkg/utils"
	"github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	inf "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
	"k8s.io/client-go/tools/cache"
)

const (
	kubeConfigEnvVariable = "KUBECONFIG"
	syncTime              = 10 * time.Minute
	IndexCustom           = "byCustomKey"
	IndexIP               = "byIP"
	TypeNode              = "Node"
	TypePod               = "Pod"
	TypeService           = "Service"
)

var (
	log        = logrus.WithField("component", "transform.Network.Kubernetes")
	cniPlugins = map[string]cni.Plugin{
		api.OVN: &cni.OVNPlugin{},
	}
	multus = cni.MultusHandler{}
	udn    = cni.UDNHandler{}
)

//nolint:revive
type InformersInterface interface {
	BuildSecondaryNetworkKeys(flow config.GenericMap, rule *api.K8sRule) []cni.SecondaryNetKey
	GetInfo([]cni.SecondaryNetKey, string) (*Info, error)
	GetNodeInfo(string) (*Info, error)
	InitFromConfig(api.NetworkTransformKubeConfig, *operational.Metrics) error
}

type Informers struct {
	InformersInterface
	// pods, nodes and services cache the different object types as *Info pointers
	pods     cache.SharedIndexInformer
	nodes    cache.SharedIndexInformer
	services cache.SharedIndexInformer
	// replicaSets caches the ReplicaSets as partially-filled *ObjectMeta pointers
	replicaSets       cache.SharedIndexInformer
	stopChan          chan struct{}
	mdStopChan        chan struct{}
	managedCNI        []string
	secondaryNetworks []api.SecondaryNetwork
	hasMultus         bool
	hasUDN            bool
	indexerHitMetric  *prometheus.CounterVec
}

type Owner struct {
	Type string
	Name string
}

// Info contains precollected metadata for Pods, Nodes and Services.
// Not all the fields are populated for all the above types. To save
// memory, we just keep in memory the necessary data for each Type.
// For more information about which fields are set for each type, please
// refer to the instantiation function of the respective informers.
type Info struct {
	// Informers need that internal object is an ObjectMeta instance
	metav1.ObjectMeta
	Type             string
	Owner            Owner
	HostName         string
	HostIP           string
	NetworkName      string
	ips              []string
	secondaryNetKeys []string
}

var (
	ipIndexer = func(obj interface{}) ([]string, error) {
		return obj.(*Info).ips, nil
	}
	customKeyIndexer = func(obj interface{}) ([]string, error) {
		return obj.(*Info).secondaryNetKeys, nil
	}
)

func (k *Informers) BuildSecondaryNetworkKeys(flow config.GenericMap, rule *api.K8sRule) []cni.SecondaryNetKey {
	return buildSecondaryNetworkKeys(flow, rule, k.secondaryNetworks, k.hasMultus, k.hasUDN)
}

func buildSecondaryNetworkKeys(flow config.GenericMap, rule *api.K8sRule, secNet []api.SecondaryNetwork, hasMultus, hasUDN bool) []cni.SecondaryNetKey {
	var keys []cni.SecondaryNetKey
	if hasMultus {
		keys = append(keys, multus.BuildKeys(flow, rule, secNet)...)
	}
	if hasUDN {
		keys = append(keys, udn.BuildKeys(flow, rule)...)
	}
	return keys
}

func (k *Informers) GetInfo(potentialKeys []cni.SecondaryNetKey, ip string) (*Info, error) {
	if info, ok := k.fetchInformers(potentialKeys, ip); ok {
		// Owner data might be discovered after the owned, so we fetch it
		// at the last moment
		if info.Owner.Name == "" {
			info.Owner = k.getOwner(info)
		}
		return info, nil
	}

	return nil, fmt.Errorf("informers can't find IP %s", ip)
}

func (k *Informers) fetchInformers(potentialKeys []cni.SecondaryNetKey, ip string) (*Info, bool) {
	if info, ok := k.fetchPodInformer(potentialKeys, ip); ok {
		// it might happen that the Host is discovered after the Pod
		if info.HostName == "" {
			info.HostName = k.getHostName(info.HostIP)
		}
		return info, true
	}
	// Nodes are only indexed by IP
	if info, ok := k.infoForIP(k.nodes.GetIndexer(), "Node", ip); ok {
		return info, true
	}
	// Services are only indexed by IP
	if info, ok := k.infoForIP(k.services.GetIndexer(), "Service", ip); ok {
		return info, true
	}
	return nil, false
}

func (k *Informers) fetchPodInformer(potentialKeys []cni.SecondaryNetKey, ip string) (*Info, bool) {
	// 1. Check if the unique key matches any Pod (secondary networks / multus case)
	if info, ok := k.infoForCustomKeys(k.pods.GetIndexer(), "Pod", potentialKeys); ok {
		return info, ok
	}
	// 2. Check if the IP matches any Pod (primary network)
	return k.infoForIP(k.pods.GetIndexer(), "Pod", ip)
}

func (k *Informers) increaseIndexerHits(kind, namespace, network, warn string) {
	k.indexerHitMetric.WithLabelValues(kind, namespace, network, warn).Inc()
}

func (k *Informers) infoForCustomKeys(idx cache.Indexer, kind string, potentialKeys []cni.SecondaryNetKey) (*Info, bool) {
	for _, key := range potentialKeys {
		objs, err := idx.ByIndex(IndexCustom, key.Key)
		if err != nil {
			k.increaseIndexerHits(kind, "", key.NetworkName, "informer error")
			log.WithError(err).WithField("key", key).Debug("error accessing unique key index, ignoring")
			return nil, false
		}
		if len(objs) > 0 {
			info := objs[0].(*Info)
			info.NetworkName = key.NetworkName
			if len(objs) > 1 {
				k.increaseIndexerHits(kind, info.Namespace, key.NetworkName, "multiple matches")
				log.WithField("key", key).Debugf("found %d objects matching this key, returning first", len(objs))
			} else {
				k.increaseIndexerHits(kind, info.Namespace, key.NetworkName, "")
			}
			log.Tracef("infoForUniqueKey found key %v", info)
			return info, true
		}
	}
	return nil, false
}

func (k *Informers) infoForIP(idx cache.Indexer, kind string, ip string) (*Info, bool) {
	objs, err := idx.ByIndex(IndexIP, ip)
	if err != nil {
		k.increaseIndexerHits(kind, "", "primary", "informer error")
		log.WithError(err).WithField("ip", ip).Debug("error accessing IP index, ignoring")
		return nil, false
	}
	if len(objs) > 0 {
		info := objs[0].(*Info)
		info.NetworkName = "primary"
		if len(objs) > 1 {
			k.increaseIndexerHits(kind, info.Namespace, "primary", "multiple matches")
			log.WithField("ip", ip).Debugf("found %d objects matching this IP, returning first", len(objs))
		} else {
			k.increaseIndexerHits(kind, info.Namespace, "primary", "")
		}
		log.Tracef("infoForIP found ip %v", info)
		return info, true
	}
	return nil, false
}

func (k *Informers) GetNodeInfo(name string) (*Info, error) {
	item, ok, err := k.nodes.GetIndexer().GetByKey(name)
	if err != nil {
		return nil, err
	} else if ok {
		return item.(*Info), nil
	}
	return nil, nil
}

func (k *Informers) getOwner(info *Info) Owner {
	if len(info.OwnerReferences) != 0 {
		ownerReference := info.OwnerReferences[0]
		if ownerReference.Kind != "ReplicaSet" {
			return Owner{
				Name: ownerReference.Name,
				Type: ownerReference.Kind,
			}
		}

		item, ok, err := k.replicaSets.GetIndexer().GetByKey(info.Namespace + "/" + ownerReference.Name)
		if err != nil {
			log.WithError(err).WithField("key", info.Namespace+"/"+ownerReference.Name).
				Debug("can't get ReplicaSet info from informer. Ignoring")
		} else if ok {
			rsInfo := item.(*metav1.ObjectMeta)
			if len(rsInfo.OwnerReferences) > 0 {
				return Owner{
					Name: rsInfo.OwnerReferences[0].Name,
					Type: rsInfo.OwnerReferences[0].Kind,
				}
			}
		}
	}
	// If no owner references found, return itself as owner
	return Owner{
		Name: info.Name,
		Type: info.Type,
	}
}

func (k *Informers) getHostName(hostIP string) string {
	if hostIP != "" {
		if info, ok := k.infoForIP(k.nodes.GetIndexer(), "Node (indirect)", hostIP); ok {
			return info.Name
		}
	}
	return ""
}

func (k *Informers) initNodeInformer(informerFactory inf.SharedInformerFactory) error {
	nodes := informerFactory.Core().V1().Nodes().Informer()
	// Transform any *v1.Node instance into a *Info instance to save space
	// in the informer's cache
	if err := nodes.SetTransform(func(i interface{}) (interface{}, error) {
		node, ok := i.(*v1.Node)
		if !ok {
			return nil, fmt.Errorf("was expecting a Node. Got: %T", i)
		}
		ips := make([]string, 0, len(node.Status.Addresses))
		hostIP := ""
		for _, address := range node.Status.Addresses {
			ip := net.ParseIP(address.Address)
			if ip != nil {
				ips = append(ips, ip.String())
				if hostIP == "" {
					hostIP = ip.String()
				}
			}
		}
		// CNI-dependent logic (must not fail when the CNI is not installed)
		for _, name := range k.managedCNI {
			if plugin := cniPlugins[name]; plugin != nil {
				moreIPs := plugin.GetNodeIPs(node)
				if moreIPs != nil {
					ips = append(ips, moreIPs...)
				}
			}
		}

		return &Info{
			ObjectMeta: metav1.ObjectMeta{
				Name:   node.Name,
				Labels: node.Labels,
			},
			ips:  ips,
			Type: TypeNode,
			// We duplicate HostIP and HostName information to simplify later filtering e.g. by
			// Host IP, where we want to get all the Pod flows by src/dst host, but also the actual
			// host-to-host flows by the same field.
			HostIP:   hostIP,
			HostName: node.Name,
		}, nil
	}); err != nil {
		return fmt.Errorf("can't set nodes transform: %w", err)
	}
	indexers := cache.Indexers{IndexIP: ipIndexer}
	if err := nodes.AddIndexers(indexers); err != nil {
		return fmt.Errorf("can't add %s indexer to Nodes informer: %w", IndexIP, err)
	}
	k.nodes = nodes
	return nil
}

func (k *Informers) initPodInformer(informerFactory inf.SharedInformerFactory) error {
	pods := informerFactory.Core().V1().Pods().Informer()
	// Transform any *v1.Pod instance into a *Info instance to save space
	// in the informer's cache
	if err := pods.SetTransform(func(i interface{}) (interface{}, error) {
		pod, ok := i.(*v1.Pod)
		if !ok {
			return nil, fmt.Errorf("was expecting a Pod. Got: %T", i)
		}
		ips := make([]string, 0, len(pod.Status.PodIPs))
		for _, ip := range pod.Status.PodIPs {
			// ignoring host-networked Pod IPs
			if ip.IP != pod.Status.HostIP {
				ips = append(ips, ip.IP)
			}
		}
		// Index from secondary network info
		var keys []string
		var err error
		if k.hasMultus {
			keys, err = multus.GetPodUniqueKeys(pod, k.secondaryNetworks)
			if err != nil {
				// Log the error as Info, do not block other ips indexing
				log.WithError(err).Infof("Secondary network cannot be identified")
			}
		}
		if k.hasUDN {
			if udnKeys, err := udn.GetPodUniqueKeys(pod); err != nil {
				// Log the error as Info, do not block other ips indexing
				log.WithError(err).Infof("UDNs cannot be identified")
			} else {
				keys = append(keys, udnKeys...)
			}
		}

		return &Info{
			ObjectMeta: metav1.ObjectMeta{
				Name:            pod.Name,
				Namespace:       pod.Namespace,
				Labels:          pod.Labels,
				OwnerReferences: pod.OwnerReferences,
			},
			Type:             TypePod,
			HostIP:           pod.Status.HostIP,
			HostName:         pod.Spec.NodeName,
			secondaryNetKeys: keys,
			ips:              ips,
		}, nil
	}); err != nil {
		return fmt.Errorf("can't set pods transform: %w", err)
	}
	indexers := cache.Indexers{
		IndexIP:     ipIndexer,
		IndexCustom: customKeyIndexer,
	}
	if err := pods.AddIndexers(indexers); err != nil {
		return fmt.Errorf("can't add indexers to Pods informer: %w", err)
	}

	k.pods = pods
	return nil
}

func (k *Informers) initServiceInformer(informerFactory inf.SharedInformerFactory) error {
	services := informerFactory.Core().V1().Services().Informer()
	// Transform any *v1.Service instance into a *Info instance to save space
	// in the informer's cache
	if err := services.SetTransform(func(i interface{}) (interface{}, error) {
		svc, ok := i.(*v1.Service)
		if !ok {
			return nil, fmt.Errorf("was expecting a Service. Got: %T", i)
		}
		ips := make([]string, 0, len(svc.Spec.ClusterIPs))
		for _, ip := range svc.Spec.ClusterIPs {
			// ignoring None IPs
			if isServiceIPSet(ip) {
				ips = append(ips, ip)
			}
		}
		return &Info{
			ObjectMeta: metav1.ObjectMeta{
				Name:      svc.Name,
				Namespace: svc.Namespace,
				Labels:    svc.Labels,
			},
			Type: TypeService,
			ips:  ips,
		}, nil
	}); err != nil {
		return fmt.Errorf("can't set services transform: %w", err)
	}
	indexers := cache.Indexers{IndexIP: ipIndexer}
	if err := services.AddIndexers(indexers); err != nil {
		return fmt.Errorf("can't add %s indexer to Pods informer: %w", IndexIP, err)
	}

	k.services = services
	return nil
}

func (k *Informers) initReplicaSetInformer(informerFactory metadatainformer.SharedInformerFactory) error {
	k.replicaSets = informerFactory.ForResource(
		schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "replicasets",
		}).Informer()
	// To save space, instead of storing a complete *metav1.ObjectMeta instance, the
	// informer's cache will store only the minimal required fields
	if err := k.replicaSets.SetTransform(func(i interface{}) (interface{}, error) {
		rs, ok := i.(*metav1.PartialObjectMetadata)
		if !ok {
			return nil, fmt.Errorf("was expecting a ReplicaSet. Got: %T", i)
		}
		return &metav1.ObjectMeta{
			Name:            rs.Name,
			Namespace:       rs.Namespace,
			OwnerReferences: rs.OwnerReferences,
		}, nil
	}); err != nil {
		return fmt.Errorf("can't set ReplicaSets transform: %w", err)
	}
	return nil
}

func (k *Informers) InitFromConfig(cfg api.NetworkTransformKubeConfig, opMetrics *operational.Metrics) error {
	// Initialization variables
	k.stopChan = make(chan struct{})
	k.mdStopChan = make(chan struct{})

	kconf, err := utils.LoadK8sConfig(cfg.ConfigPath)
	if err != nil {
		return err
	}

	kubeClient, err := kubernetes.NewForConfig(kconf)
	if err != nil {
		return err
	}

	metaKubeClient, err := metadata.NewForConfig(kconf)
	if err != nil {
		return err
	}

	k.managedCNI = cfg.ManagedCNI
	if k.managedCNI == nil {
		k.managedCNI = []string{api.OVN}
	}
	k.secondaryNetworks = cfg.SecondaryNetworks
	for _, netConfig := range cfg.SecondaryNetworks {
		for index := range netConfig.Index {
			if multus.Manages(index) {
				k.hasMultus = true
			}
			if udn.Manages(index) {
				k.hasUDN = true
			}
		}
	}
	k.indexerHitMetric = opMetrics.CreateIndexerHitCounter()
	err = k.initInformers(kubeClient, metaKubeClient)
	if err != nil {
		return err
	}

	return nil
}

func (k *Informers) initInformers(client kubernetes.Interface, metaClient metadata.Interface) error {
	informerFactory := inf.NewSharedInformerFactory(client, syncTime)
	metadataInformerFactory := metadatainformer.NewSharedInformerFactory(metaClient, syncTime)
	err := k.initNodeInformer(informerFactory)
	if err != nil {
		return err
	}
	err = k.initPodInformer(informerFactory)
	if err != nil {
		return err
	}
	err = k.initServiceInformer(informerFactory)
	if err != nil {
		return err
	}
	err = k.initReplicaSetInformer(metadataInformerFactory)
	if err != nil {
		return err
	}

	log.Debugf("starting kubernetes informers, waiting for synchronization")
	informerFactory.Start(k.stopChan)
	informerFactory.WaitForCacheSync(k.stopChan)
	log.Debugf("kubernetes informers started")

	log.Debugf("starting kubernetes metadata informers, waiting for synchronization")
	metadataInformerFactory.Start(k.mdStopChan)
	metadataInformerFactory.WaitForCacheSync(k.mdStopChan)
	log.Debugf("kubernetes metadata informers started")
	return nil
}

func isServiceIPSet(ip string) bool {
	return ip != v1.ClusterIPNone && ip != ""
}
