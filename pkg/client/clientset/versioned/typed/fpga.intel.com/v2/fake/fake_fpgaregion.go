// Copyright 2020 Intel Corporation. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v2 "github.com/intel/intel-device-plugins-for-kubernetes/pkg/apis/fpga.intel.com/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeFpgaRegions implements FpgaRegionInterface
type FakeFpgaRegions struct {
	Fake *FakeFpgaV2
	ns   string
}

var fpgaregionsResource = schema.GroupVersionResource{Group: "fpga.intel.com", Version: "v2", Resource: "fpgaregions"}

var fpgaregionsKind = schema.GroupVersionKind{Group: "fpga.intel.com", Version: "v2", Kind: "FpgaRegion"}

// Get takes name of the fpgaRegion, and returns the corresponding fpgaRegion object, and an error if there is any.
func (c *FakeFpgaRegions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v2.FpgaRegion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(fpgaregionsResource, c.ns, name), &v2.FpgaRegion{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v2.FpgaRegion), err
}

// List takes label and field selectors, and returns the list of FpgaRegions that match those selectors.
func (c *FakeFpgaRegions) List(ctx context.Context, opts v1.ListOptions) (result *v2.FpgaRegionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(fpgaregionsResource, fpgaregionsKind, c.ns, opts), &v2.FpgaRegionList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v2.FpgaRegionList{ListMeta: obj.(*v2.FpgaRegionList).ListMeta}
	for _, item := range obj.(*v2.FpgaRegionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested fpgaRegions.
func (c *FakeFpgaRegions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(fpgaregionsResource, c.ns, opts))

}

// Create takes the representation of a fpgaRegion and creates it.  Returns the server's representation of the fpgaRegion, and an error, if there is any.
func (c *FakeFpgaRegions) Create(ctx context.Context, fpgaRegion *v2.FpgaRegion, opts v1.CreateOptions) (result *v2.FpgaRegion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(fpgaregionsResource, c.ns, fpgaRegion), &v2.FpgaRegion{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v2.FpgaRegion), err
}

// Update takes the representation of a fpgaRegion and updates it. Returns the server's representation of the fpgaRegion, and an error, if there is any.
func (c *FakeFpgaRegions) Update(ctx context.Context, fpgaRegion *v2.FpgaRegion, opts v1.UpdateOptions) (result *v2.FpgaRegion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(fpgaregionsResource, c.ns, fpgaRegion), &v2.FpgaRegion{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v2.FpgaRegion), err
}

// Delete takes name of the fpgaRegion and deletes it. Returns an error if one occurs.
func (c *FakeFpgaRegions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(fpgaregionsResource, c.ns, name), &v2.FpgaRegion{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeFpgaRegions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(fpgaregionsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v2.FpgaRegionList{})
	return err
}

// Patch applies the patch and returns the patched fpgaRegion.
func (c *FakeFpgaRegions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v2.FpgaRegion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(fpgaregionsResource, c.ns, name, pt, data, subresources...), &v2.FpgaRegion{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v2.FpgaRegion), err
}