package datasource

import (
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/transform/kubernetes/cni"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/transform/kubernetes/informers"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/transform/kubernetes/model"
)

type Datasource struct {
	Informers informers.InformersInterface
}

func NewInformerDatasource(kubeconfig string, infConfig informers.Config, opMetrics *operational.Metrics) (*Datasource, error) {
	inf := &informers.Informers{}
	if err := inf.InitFromConfig(kubeconfig, infConfig, opMetrics); err != nil {
		return nil, err
	}
	return &Datasource{Informers: inf}, nil
}

func (d *Datasource) IndexLookup(potentialKeys []cni.SecondaryNetKey, ip string) *model.ResourceMetaData {
	return d.Informers.IndexLookup(potentialKeys, ip)
}

func (d *Datasource) GetNodeByName(name string) (*model.ResourceMetaData, error) {
	return d.Informers.GetNodeByName(name)
}
