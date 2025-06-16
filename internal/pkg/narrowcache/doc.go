// Package narrowcache provides an additional cache layer for a Kubernetes client, specialized in caching
// explicitly requested objects rather than full GVKs.
//
// A typical use case is to reduce memory consumption when reading commonly used objects, such as
// ConfigMaps or Secrets. Sometimes, the controller-runtime cache limitation settings (by namespace, by
// label) aren't an option, which leads to excessive memory consumption due to caching many unneeded
// resources. This is quite common when namespaces or labels aren't known beforehand, hence cannot
// be used for static cache configuration. This package aims to address these use cases.
//
// # Examples
//
//	 // This creates a narrowcache.Client for ConfigMaps and Secrets.
//	 narrowCache := narrowcache.NewConfig(cfg, narrowcache.ConfigMaps, narrowcache.Secrets)
//	 client, err := narrowCache.CreateClient(mgr.GetClient())
//	 if err != nil {
//		 setupLog.Error(err, "unable to create narrow cache client")
//		 os.Exit(1)
//	 }
//
// This new cache layer is invoked for any call to `client.Get` on the configured GVK. For other
// GVKs, the underlying client is used.
//
// Furthermore, cached objects are available as `source.Source`, so that they can be used as controller
// watches to enqueue reconcile requests:
//
//	 src, _ := client.GetSource(ctx, obj)
//	 controller.Watch(
//		 src,
//		 handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
//			// etc.
//		 }),
//	 )
package narrowcache
