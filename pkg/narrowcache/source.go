package narrowcache

import (
	"context"
	"errors"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type NarrowSource struct {
	source.Source
	handler handler.EventHandler
	onStart func(ctx context.Context, q workqueue.TypedRateLimitingInterface[reconcile.Request])
}

func (s *NarrowSource) Start(ctx context.Context, q workqueue.TypedRateLimitingInterface[reconcile.Request]) error {
	if s.handler == nil {
		return errors.New("must specify NarrowSource.handler")
	}
	s.onStart(ctx, q)
	return nil
}
