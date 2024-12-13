package pipeline

import (
	"context"
	"encoding/json"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"

	"github.com/netobserv/flowlogs-pipeline/pkg/utils"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type pipelineConfigWatcher struct {
	clientSet        kubernetes.Clientset
	cmName           string
	cmNamespace      string
	configFile       string
	pipelineEntryMap map[string]*pipelineEntry
}

func newPipelineConfigWatcher(cfg *config.ConfigFileStruct, pipelineEntryMap map[string]*pipelineEntry) (*pipelineConfigWatcher, error) {
	if cfg.DynamicParameters.Name == "" ||
		cfg.DynamicParameters.Namespace == "" ||
		cfg.DynamicParameters.FileName == "" {
		return nil, nil
	}

	config, err := utils.LoadK8sConfig(cfg.DynamicParameters.KubeConfigPath)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	pipelineCW := pipelineConfigWatcher{
		clientSet:        *clientset,
		pipelineEntryMap: pipelineEntryMap,
		cmName:           cfg.DynamicParameters.Name,
		cmNamespace:      cfg.DynamicParameters.Namespace,
		configFile:       cfg.DynamicParameters.FileName,
	}

	return &pipelineCW, nil

}

func (pcw *pipelineConfigWatcher) Run() {
	for {
		watcher, err := pcw.clientSet.CoreV1().ConfigMaps(pcw.cmNamespace).Watch(context.TODO(),
			metav1.SingleObject(metav1.ObjectMeta{Name: pcw.cmName, Namespace: pcw.cmNamespace}))
		if err != nil {
			log.Errorf("Unable to create watcher: %s", err)
			return
		}
		pcw.handleEvent(watcher.ResultChan())
	}
}

func (pcw *pipelineConfigWatcher) handleEvent(eventChannel <-chan watch.Event) {
	for {
		event, open := <-eventChannel
		if open {
			switch event.Type {
			case watch.Added:
				fallthrough
			case watch.Modified:
				// Update our endpoint
				if updatedMap, ok := event.Object.(*corev1.ConfigMap); ok {
					pcw.updateFromConfigmap(updatedMap)
				}
			case watch.Deleted:
				fallthrough
			case watch.Bookmark:
			case watch.Error:
			default:
				// Do nothing
			}
		} else {
			// If eventChannel is closed, it means the server has closed the connection
			return
		}
	}
}

func (pcw *pipelineConfigWatcher) updateFromConfigmap(cm *corev1.ConfigMap) {
	if rawConfig, ok := cm.Data[pcw.configFile]; ok {
		config := config.HotReloadStruct{}
		err := json.Unmarshal([]byte(rawConfig), &config)
		if err != nil {
			log.Errorf("Cannot parse config: %v", err)
			return
		}
		for _, param := range config.Parameters {
			if pentry, ok := pcw.pipelineEntryMap[param.Name]; ok {
				pcw.updateEntry(pentry, param)
			}
		}
	}
}

func (pcw *pipelineConfigWatcher) updateEntry(pEntry *pipelineEntry, param config.StageParam) {
	switch pEntry.stageType {
	case StageEncode:
		pEntry.Encoder.Update(param)
	default:
		log.Warningf("Hot reloading not supported for: %s", pEntry.stageType)
	}
}
