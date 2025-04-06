package informers

import (
	"errors"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/transform/kubernetes/cni"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/transform/kubernetes/model"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

var (
	secondaryNetConfig = []api.SecondaryNetwork{
		{
			Name:  "my-network",
			Index: map[string]any{"mac": nil},
		},
		{
			Name:  "ovn-udn",
			Index: map[string]any{"udn": nil},
		},
	}
)

type Mock struct {
	mock.Mock
	InformersInterface
}

func NewInformersMock() *Mock {
	inf := new(Mock)
	inf.On("InitFromConfig", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	return inf
}

func (o *Mock) InitFromConfig(kubeconfig string, infConfig Config, opMetrics *operational.Metrics) error {
	args := o.Called(kubeconfig, infConfig, opMetrics)
	return args.Error(0)
}

type IndexerMock struct {
	mock.Mock
	cache.Indexer
	parentChecker func(*model.ResourceMetaData)
}

type InformerMock struct {
	mock.Mock
	InformerInterface
}

type InformerInterface interface {
	cache.SharedInformer
	AddIndexers(indexers cache.Indexers) error
	GetIndexer() cache.Indexer
}

func (m *IndexerMock) ByIndex(indexName, indexedValue string) ([]interface{}, error) {
	args := m.Called(indexName, indexedValue)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *IndexerMock) GetByKey(key string) (interface{}, bool, error) {
	args := m.Called(key)
	return args.Get(0), args.Bool(1), args.Error(2)
}

func (m *InformerMock) GetIndexer() cache.Indexer {
	args := m.Called()
	return args.Get(0).(cache.Indexer)
}

func (m *IndexerMock) MockPod(ip, mac, intf, name, namespace, nodeIP, ownerName, ownerKind string) {
	res := model.ResourceMetaData{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Kind:      "Pod",
		OwnerName: ownerName,
		OwnerKind: ownerKind,
		HostIP:    nodeIP,
	}
	m.parentChecker(&res)
	if len(mac) > 0 {
		nsi := cni.NetStatItem{
			Interface: intf,
			MAC:       mac,
			IPs:       []string{ip},
		}
		res.SecondaryNetKeys = nsi.Keys(secondaryNetConfig[0])
		m.On("ByIndex", IndexCustom, res.SecondaryNetKeys[0]).Return([]interface{}{&res}, nil)
	}
	if len(ip) > 0 {
		res.IPs = []string{ip}
		m.On("ByIndex", IndexIP, ip).Return([]interface{}{&res}, nil)
	}
}

func (m *IndexerMock) MockNode(ip, name string) {
	m.On("ByIndex", IndexIP, ip).Return([]interface{}{&model.ResourceMetaData{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Kind:      "Node",
		OwnerKind: "Node",
		OwnerName: name,
		IPs:       []string{ip},
	}}, nil)
}

func (m *IndexerMock) MockService(ip, name, namespace string) {
	m.On("ByIndex", IndexIP, ip).Return([]interface{}{&model.ResourceMetaData{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Kind:      "Service",
		OwnerKind: "Service",
		OwnerName: name,
		IPs:       []string{ip},
	}}, nil)
}

func (m *IndexerMock) MockReplicaSet(name, namespace, ownerName, ownerKind string) {
	m.On("GetByKey", namespace+"/"+name).Return(&metav1.ObjectMeta{
		Name: name,
		OwnerReferences: []metav1.OwnerReference{{
			Kind: ownerKind,
			Name: ownerName,
		}},
	}, true, nil)
}

func (m *IndexerMock) FallbackNotFound() {
	m.On("ByIndex", IndexIP, mock.Anything).Return([]interface{}{}, nil)
}

func SetupIndexerMocks(kd *Informers) (pods, nodes, svc, rs *IndexerMock) {
	// pods informer
	pods = &IndexerMock{parentChecker: kd.checkParent}
	pim := InformerMock{}
	pim.On("GetIndexer").Return(pods)
	kd.pods = &pim
	// nodes informer
	nodes = &IndexerMock{}
	him := InformerMock{}
	him.On("GetIndexer").Return(nodes)
	kd.nodes = &him
	// svc informer
	svc = &IndexerMock{}
	sim := InformerMock{}
	sim.On("GetIndexer").Return(svc)
	kd.services = &sim
	// rs informer
	rs = &IndexerMock{}
	rim := InformerMock{}
	rim.On("GetIndexer").Return(rs)
	kd.replicaSets = &rim
	return
}

type FakeInformers struct {
	InformersInterface
	ipInfo         map[string]*model.ResourceMetaData
	customKeysInfo map[string]*model.ResourceMetaData
	nodes          map[string]*model.ResourceMetaData
}

func SetupStubs(ipInfo, customKeysInfo, nodes map[string]*model.ResourceMetaData) (Config, *FakeInformers) {
	cfg := NewConfig(api.NetworkTransformKubeConfig{SecondaryNetworks: secondaryNetConfig})
	return cfg, &FakeInformers{
		ipInfo:         ipInfo,
		customKeysInfo: customKeysInfo,
		nodes:          nodes,
	}
}

func (f *FakeInformers) InitFromConfig(_ string, _ Config, _ *operational.Metrics) error {
	return nil
}

func (f *FakeInformers) IndexLookup(keys []cni.SecondaryNetKey, ip string) *model.ResourceMetaData {
	for _, key := range keys {
		i := f.customKeysInfo[key.Key]
		if i != nil {
			return i
		}
	}

	i := f.ipInfo[ip]
	if i != nil {
		return i
	}
	return nil
}

func (f *FakeInformers) GetNodeByName(n string) (*model.ResourceMetaData, error) {
	i := f.nodes[n]
	if i != nil {
		return i, nil
	}
	return nil, errors.New("notFound")
}
