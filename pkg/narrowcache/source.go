package narrowcache

import (
	"context"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type NarrowSource struct {
	source.Source
	onStart func(ctx context.Context, h handler.EventHandler, q workqueue.RateLimitingInterface)
}

func (s *NarrowSource) Start(ctx context.Context, h handler.EventHandler, q workqueue.RateLimitingInterface, _ ...predicate.Predicate) error {
	s.onStart(ctx, h, q)
	return nil
}
