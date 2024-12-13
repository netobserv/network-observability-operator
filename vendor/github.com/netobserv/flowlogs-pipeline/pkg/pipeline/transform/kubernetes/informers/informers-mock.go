package informers

import (
	"errors"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/transform/kubernetes/cni"
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
	}
)

type Mock struct {
	mock.Mock
	InformersInterface
}

func NewInformersMock() *Mock {
	inf := new(Mock)
	inf.On("InitFromConfig", mock.Anything, mock.Anything).Return(nil)
	return inf
}

func (o *Mock) InitFromConfig(cfg api.NetworkTransformKubeConfig, opMetrics *operational.Metrics) error {
	args := o.Called(cfg, opMetrics)
	return args.Error(0)
}

type IndexerMock struct {
	mock.Mock
	cache.Indexer
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

func (m *IndexerMock) MockPod(ip, mac, intf, name, namespace, nodeIP string, owner *Owner) {
	var ownerRef []metav1.OwnerReference
	if owner != nil {
		ownerRef = []metav1.OwnerReference{{
			Kind: owner.Type,
			Name: owner.Name,
		}}
	}
	info := Info{
		Type: "Pod",
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			OwnerReferences: ownerRef,
		},
		HostIP:           nodeIP,
		ips:              []string{},
		secondaryNetKeys: []string{},
	}
	if len(mac) > 0 {
		nsi := cni.NetStatItem{
			Interface: intf,
			MAC:       mac,
			IPs:       []string{ip},
		}
		info.secondaryNetKeys = nsi.Keys(secondaryNetConfig[0])
		m.On("ByIndex", IndexCustom, info.secondaryNetKeys[0]).Return([]interface{}{&info}, nil)
	}
	if len(ip) > 0 {
		info.ips = []string{ip}
		m.On("ByIndex", IndexIP, ip).Return([]interface{}{&info}, nil)
	}
}

func (m *IndexerMock) MockNode(ip, name string) {
	m.On("ByIndex", IndexIP, ip).Return([]interface{}{&Info{
		Type:       "Node",
		ObjectMeta: metav1.ObjectMeta{Name: name},
		ips:        []string{ip},
	}}, nil)
}

func (m *IndexerMock) MockService(ip, name, namespace string) {
	m.On("ByIndex", IndexIP, ip).Return([]interface{}{&Info{
		Type:       "Service",
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		ips:        []string{ip},
	}}, nil)
}

func (m *IndexerMock) MockReplicaSet(name, namespace string, owner Owner) {
	m.On("GetByKey", namespace+"/"+name).Return(&metav1.ObjectMeta{
		Name: name,
		OwnerReferences: []metav1.OwnerReference{{
			Kind: owner.Type,
			Name: owner.Name,
		}},
	}, true, nil)
}

func (m *IndexerMock) FallbackNotFound() {
	m.On("ByIndex", IndexIP, mock.Anything).Return([]interface{}{}, nil)
}

func SetupIndexerMocks(kd *Informers) (pods, nodes, svc, rs *IndexerMock) {
	// pods informer
	pods = &IndexerMock{}
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
	ipInfo         map[string]*Info
	customKeysInfo map[string]*Info
	nodes          map[string]*Info
}

func SetupStubs(ipInfo map[string]*Info, customKeysInfo map[string]*Info, nodes map[string]*Info) *FakeInformers {
	return &FakeInformers{
		ipInfo:         ipInfo,
		customKeysInfo: customKeysInfo,
		nodes:          nodes,
	}
}

func (f *FakeInformers) InitFromConfig(_ api.NetworkTransformKubeConfig, _ *operational.Metrics) error {
	return nil
}

func (f *FakeInformers) GetInfo(keys []cni.SecondaryNetKey, ip string) (*Info, error) {
	if len(keys) > 0 {
		i := f.customKeysInfo[keys[0].Key]
		if i != nil {
			return i, nil
		}
	}

	i := f.ipInfo[ip]
	if i != nil {
		return i, nil
	}
	return nil, errors.New("notFound")
}

func (f *FakeInformers) BuildSecondaryNetworkKeys(flow config.GenericMap, rule *api.K8sRule) []cni.SecondaryNetKey {
	m := cni.MultusHandler{}
	return m.BuildKeys(flow, rule, secondaryNetConfig)
}

func (f *FakeInformers) GetNodeInfo(n string) (*Info, error) {
	i := f.nodes[n]
	if i != nil {
		return i, nil
	}
	return nil, errors.New("notFound")
}
