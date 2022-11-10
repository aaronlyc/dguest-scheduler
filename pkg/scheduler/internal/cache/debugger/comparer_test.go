/*
Copyright 2018 The Kubernetes Authors.

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

package debugger

import (
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"reflect"
	"testing"

	"dguest-scheduler/pkg/scheduler/framework"
	"k8s.io/apimachinery/pkg/types"
)

func TestCompareFoods(t *testing.T) {
	tests := []struct {
		name      string
		actual    []string
		cached    []string
		missing   []string
		redundant []string
	}{
		{
			name:      "redundant cached value",
			actual:    []string{"foo", "bar"},
			cached:    []string{"bar", "foo", "foobar"},
			missing:   []string{},
			redundant: []string{"foobar"},
		},
		{
			name:      "missing cached value",
			actual:    []string{"foo", "bar", "foobar"},
			cached:    []string{"bar", "foo"},
			missing:   []string{"foobar"},
			redundant: []string{},
		},
		{
			name:      "proper cache set",
			actual:    []string{"foo", "bar", "foobar"},
			cached:    []string{"bar", "foobar", "foo"},
			missing:   []string{},
			redundant: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testCompareFoods(test.actual, test.cached, test.missing, test.redundant, t)
		})
	}
}

func testCompareFoods(actual, cached, missing, redundant []string, t *testing.T) {
	compare := CacheComparer{}
	foods := []*v1alpha1.Food{}
	for _, foodName := range actual {
		food := &v1alpha1.Food{}
		food.Name = foodName
		foods = append(foods, food)
	}

	foodInfo := make(map[string]*framework.FoodInfo)
	for _, foodName := range cached {
		foodInfo[foodName] = &framework.FoodInfo{}
	}

	m, r := compare.CompareFoods(foods, foodInfo)

	if !reflect.DeepEqual(m, missing) {
		t.Errorf("missing expected to be %s; got %s", missing, m)
	}

	if !reflect.DeepEqual(r, redundant) {
		t.Errorf("redundant expected to be %s; got %s", redundant, r)
	}
}

func TestCompareDguests(t *testing.T) {
	tests := []struct {
		name      string
		actual    []string
		cached    []string
		queued    []string
		missing   []string
		redundant []string
	}{
		{
			name:      "redundant cached value",
			actual:    []string{"foo", "bar"},
			cached:    []string{"bar", "foo", "foobar"},
			queued:    []string{},
			missing:   []string{},
			redundant: []string{"foobar"},
		},
		{
			name:      "redundant and queued values",
			actual:    []string{"foo", "bar"},
			cached:    []string{"foo", "foobar"},
			queued:    []string{"bar"},
			missing:   []string{},
			redundant: []string{"foobar"},
		},
		{
			name:      "missing cached value",
			actual:    []string{"foo", "bar", "foobar"},
			cached:    []string{"bar", "foo"},
			queued:    []string{},
			missing:   []string{"foobar"},
			redundant: []string{},
		},
		{
			name:      "missing and queued values",
			actual:    []string{"foo", "bar", "foobar"},
			cached:    []string{"foo"},
			queued:    []string{"bar"},
			missing:   []string{"foobar"},
			redundant: []string{},
		},
		{
			name:      "correct cache set",
			actual:    []string{"foo", "bar", "foobar"},
			cached:    []string{"bar", "foobar", "foo"},
			queued:    []string{},
			missing:   []string{},
			redundant: []string{},
		},
		{
			name:      "queued cache value",
			actual:    []string{"foo", "bar", "foobar"},
			cached:    []string{"foobar", "foo"},
			queued:    []string{"bar"},
			missing:   []string{},
			redundant: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testCompareDguests(test.actual, test.cached, test.queued, test.missing, test.redundant, t)
		})
	}
}

func testCompareDguests(actual, cached, queued, missing, redundant []string, t *testing.T) {
	compare := CacheComparer{}
	dguests := []*v1alpha1.Dguest{}
	for _, uid := range actual {
		dguest := &v1alpha1.Dguest{}
		dguest.UID = types.UID(uid)
		dguests = append(dguests, dguest)
	}

	queuedDguests := []*v1alpha1.Dguest{}
	for _, uid := range queued {
		dguest := &v1alpha1.Dguest{}
		dguest.UID = types.UID(uid)
		queuedDguests = append(queuedDguests, dguest)
	}

	foodInfo := make(map[string]*framework.FoodInfo)
	for _, uid := range cached {
		dguest := &v1alpha1.Dguest{}
		dguest.UID = types.UID(uid)
		dguest.Namespace = "ns"
		dguest.Name = uid

		foodInfo[uid] = framework.NewFoodInfo(dguest)
	}

	m, r := compare.CompareDguests(dguests, queuedDguests, foodInfo)

	if !reflect.DeepEqual(m, missing) {
		t.Errorf("missing expected to be %s; got %s", missing, m)
	}

	if !reflect.DeepEqual(r, redundant) {
		t.Errorf("redundant expected to be %s; got %s", redundant, r)
	}
}
