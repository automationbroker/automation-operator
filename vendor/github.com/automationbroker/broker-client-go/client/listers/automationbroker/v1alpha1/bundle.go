/*
Copyright (c) 2018 Red Hat, Inc.

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

// This file was automatically generated by lister-gen

package v1alpha1

import (
	v1alpha1 "github.com/automationbroker/broker-client-go/pkg/apis/automationbroker/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// BundleLister helps list Bundles.
type BundleLister interface {
	// List lists all Bundles in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.Bundle, err error)
	// Bundles returns an object that can list and get Bundles.
	Bundles(namespace string) BundleNamespaceLister
	BundleListerExpansion
}

// bundleLister implements the BundleLister interface.
type bundleLister struct {
	indexer cache.Indexer
}

// NewBundleLister returns a new BundleLister.
func NewBundleLister(indexer cache.Indexer) BundleLister {
	return &bundleLister{indexer: indexer}
}

// List lists all Bundles in the indexer.
func (s *bundleLister) List(selector labels.Selector) (ret []*v1alpha1.Bundle, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Bundle))
	})
	return ret, err
}

// Bundles returns an object that can list and get Bundles.
func (s *bundleLister) Bundles(namespace string) BundleNamespaceLister {
	return bundleNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// BundleNamespaceLister helps list and get Bundles.
type BundleNamespaceLister interface {
	// List lists all Bundles in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.Bundle, err error)
	// Get retrieves the Bundle from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.Bundle, error)
	BundleNamespaceListerExpansion
}

// bundleNamespaceLister implements the BundleNamespaceLister
// interface.
type bundleNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Bundles in the indexer for a given namespace.
func (s bundleNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Bundle, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Bundle))
	})
	return ret, err
}

// Get retrieves the Bundle from the indexer for a given namespace and name.
func (s bundleNamespaceLister) Get(name string) (*v1alpha1.Bundle, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("bundle"), name)
	}
	return obj.(*v1alpha1.Bundle), nil
}
