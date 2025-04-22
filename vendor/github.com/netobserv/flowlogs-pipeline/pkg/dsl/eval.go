package dsl

import (
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/utils/filters"
)

type tree struct {
	logicalOp string
	children  []*tree
	predicate filters.Predicate
}

func (t *tree) apply(flow config.GenericMap) bool {
	if t.predicate != nil {
		return t.predicate(flow)
	}
	if t.logicalOp == operatorAnd {
		for _, child := range t.children {
			if !child.apply(flow) {
				return false
			}
		}
		return true
	}
	// t.logicalOp == operatorOr
	for _, child := range t.children {
		if child.apply(flow) {
			return true
		}
	}
	return false
}
