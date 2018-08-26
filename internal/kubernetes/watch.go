package kubernetes

import (
	"time"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	resourceNode = "nodes"
)

// An NodeStore is a cache of node resources.
type NodeStore interface {
	// Get an node by name. Returns an error if the node does not exist.
	Get(name string) (*core.Node, error)
}

// An NodeWatch is a cache of node resources that notifies registered
// handlers when its contents change.
type NodeWatch struct {
	cache.SharedInformer
}

// NewNodeWatch creates a watch on node resources. Nodes are cached and the
// provided ResourceEventHandlers are called when the cache changes.
func NewNodeWatch(c kubernetes.Interface, rs ...cache.ResourceEventHandler) *NodeWatch {
	lw := &cache.ListWatch{
		ListFunc:  func(o meta.ListOptions) (runtime.Object, error) { return c.CoreV1().Nodes().List(o) },
		WatchFunc: func(o meta.ListOptions) (watch.Interface, error) { return c.CoreV1().Nodes().Watch(o) },
	}
	i := cache.NewSharedInformer(lw, &core.Node{}, 30*time.Minute)
	for _, r := range rs {
		i.AddEventHandler(r)
	}
	return &NodeWatch{i}
}

// Get an node by name. Returns an error if the node does not exist.
func (w *NodeWatch) Get(name string) (*core.Node, error) {
	o, exists, err := w.GetStore().GetByKey(name)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get node %s", name)
	}
	if !exists {
		return nil, errors.Errorf("node %s does not exist", name)
	}
	return o.(*core.Node), nil
}