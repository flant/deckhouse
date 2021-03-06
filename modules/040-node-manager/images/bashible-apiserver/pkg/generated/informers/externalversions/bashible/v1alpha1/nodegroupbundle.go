/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	bashiblev1alpha1 "d8.io/bashible/pkg/apis/bashible/v1alpha1"
	versioned "d8.io/bashible/pkg/generated/clientset/versioned"
	internalinterfaces "d8.io/bashible/pkg/generated/informers/externalversions/internalinterfaces"
	v1alpha1 "d8.io/bashible/pkg/generated/listers/bashible/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// NodeGroupBundleInformer provides access to a shared informer and lister for
// NodeGroupBundles.
type NodeGroupBundleInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.NodeGroupBundleLister
}

type nodeGroupBundleInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewNodeGroupBundleInformer constructs a new informer for NodeGroupBundle type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewNodeGroupBundleInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredNodeGroupBundleInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredNodeGroupBundleInformer constructs a new informer for NodeGroupBundle type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredNodeGroupBundleInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.BashibleV1alpha1().NodeGroupBundles().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.BashibleV1alpha1().NodeGroupBundles().Watch(context.TODO(), options)
			},
		},
		&bashiblev1alpha1.NodeGroupBundle{},
		resyncPeriod,
		indexers,
	)
}

func (f *nodeGroupBundleInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredNodeGroupBundleInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *nodeGroupBundleInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&bashiblev1alpha1.NodeGroupBundle{}, f.defaultInformer)
}

func (f *nodeGroupBundleInformer) Lister() v1alpha1.NodeGroupBundleLister {
	return v1alpha1.NewNodeGroupBundleLister(f.Informer().GetIndexer())
}
