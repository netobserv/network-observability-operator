package narrowcache

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	cfg  *rest.Config
	info []GVKInfo
}

type GVKInfo struct {
	Obj     client.Object
	Getter  func(ctx context.Context, cl kubernetes.Interface, key client.ObjectKey) (runtime.Object, error)
	Watcher func(ctx context.Context, cl kubernetes.Interface, key client.ObjectKey) (watch.Interface, error)
}

var (
	ConfigMaps = GVKInfo{
		Obj: &corev1.ConfigMap{},
		Getter: func(ctx context.Context, cl kubernetes.Interface, key client.ObjectKey) (runtime.Object, error) {
			return cl.CoreV1().ConfigMaps(key.Namespace).Get(ctx, key.Name, metav1.GetOptions{})
		},
		Watcher: func(ctx context.Context, cl kubernetes.Interface, key client.ObjectKey) (watch.Interface, error) {
			opts := metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector(metav1.ObjectNameField, key.Name).String()}
			return cl.CoreV1().ConfigMaps(key.Namespace).Watch(ctx, opts)
		},
	}
	Secrets = GVKInfo{
		Obj: &corev1.Secret{},
		Getter: func(ctx context.Context, cl kubernetes.Interface, key client.ObjectKey) (runtime.Object, error) {
			return cl.CoreV1().Secrets(key.Namespace).Get(ctx, key.Name, metav1.GetOptions{})
		},
		Watcher: func(ctx context.Context, cl kubernetes.Interface, key client.ObjectKey) (watch.Interface, error) {
			opts := metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector(metav1.ObjectNameField, key.Name).String()}
			return cl.CoreV1().Secrets(key.Namespace).Watch(ctx, opts)
		},
	}
	NewLiveClient func(c *rest.Config) (kubernetes.Interface, error) = func(c *rest.Config) (kubernetes.Interface, error) {
		return kubernetes.NewForConfig(c)
	}
)

func NewConfig(cfg *rest.Config, info ...GVKInfo) *Config {
	return &Config{
		cfg:  cfg,
		info: info,
	}
}

func (c *Config) ControllerRuntimeClientCacheOptions() *client.CacheOptions {
	disabled := []client.Object{}
	for _, info := range c.info {
		disabled = append(disabled, info.Obj)
	}
	return &client.CacheOptions{DisableFor: disabled}
}

// CreateClient creates a new client layer that sits on top of the provided `underlying` client.
// This client implements Get for the provided GVKs, using a specific cache.
//
// Other kinds of requests (ie. non-get and non-managed GVKs) are forwarded to the `underlying`
// client.
//
// Furthermore, cached objects are also available as `source.Source`.
func (c *Config) CreateClient(underlying client.Client) (*Client, error) {
	liveClient, err := NewLiveClient(c.cfg)
	if err != nil {
		return nil, err
	}
	watchedGVKs := make(map[string]GVKInfo, len(c.info))
	for _, inf := range c.info {
		gvk, err := underlying.GroupVersionKindFor(inf.Obj)
		if err != nil {
			return nil, err
		}
		watchedGVKs[gvk.String()] = inf
	}
	return &Client{
		Client:         underlying,
		liveClient:     liveClient,
		watchedGVKs:    watchedGVKs,
		watchedObjects: make(map[string]*watchedObject),
	}, nil
}
