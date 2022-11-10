/*
Copyright 2015 The Kubernetes Authors.

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

package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	// DefaultExtenderTimeout defines the default extender timeout in second.
	DefaultExtenderTimeout = 5 * time.Second
)

// HTTPExtender implements the Extender interface.
type HTTPExtender struct {
	extenderURL      string
	preemptVerb      string
	filterVerb       string
	prioritizeVerb   string
	bindVerb         string
	weight           int64
	client           *http.Client
	foodCacheCapable bool
	managedResources sets.String
	ignorable        bool
}

//func makeTransport(config *schedulerapi.Extender) (http.RoundTripper, error) {
//	var cfg restclient.Config
//	if config.TLSConfig != nil {
//		cfg.TLSClientConfig.Insecure = config.TLSConfig.Insecure
//		cfg.TLSClientConfig.ServerName = config.TLSConfig.ServerName
//		cfg.TLSClientConfig.CertFile = config.TLSConfig.CertFile
//		cfg.TLSClientConfig.KeyFile = config.TLSConfig.KeyFile
//		cfg.TLSClientConfig.CAFile = config.TLSConfig.CAFile
//		cfg.TLSClientConfig.CertData = config.TLSConfig.CertData
//		cfg.TLSClientConfig.KeyData = config.TLSConfig.KeyData
//		cfg.TLSClientConfig.CAData = config.TLSConfig.CAData
//	}
//	if config.EnableHTTPS {
//		hasCA := len(cfg.CAFile) > 0 || len(cfg.CAData) > 0
//		if !hasCA {
//			cfg.Insecure = true
//		}
//	}
//	tlsConfig, err := restclient.TLSConfigFor(&cfg)
//	if err != nil {
//		return nil, err
//	}
//	if tlsConfig != nil {
//		return utilnet.SetTransportDefaults(&http.Transport{
//			TLSClientConfig: tlsConfig,
//		}), nil
//	}
//	return utilnet.SetTransportDefaults(&http.Transport{}), nil
//}
//
//// NewHTTPExtender creates an HTTPExtender object.
//func NewHTTPExtender(config *schedulerapi.Extender) (framework.Extender, error) {
//	if config.HTTPTimeout.Duration.Nanoseconds() == 0 {
//		config.HTTPTimeout.Duration = time.Duration(DefaultExtenderTimeout)
//	}
//
//	transport, err := makeTransport(config)
//	if err != nil {
//		return nil, err
//	}
//	schdulerClient := &http.Client{
//		Transport: transport,
//		Timeout:   config.HTTPTimeout.Duration,
//	}
//	managedResources := sets.NewString()
//	for _, r := range config.ManagedResources {
//		managedResources.Insert(string(r.Name))
//	}
//	return &HTTPExtender{
//		extenderURL:      config.URLPrefix,
//		preemptVerb:      config.PreemptVerb,
//		filterVerb:       config.FilterVerb,
//		prioritizeVerb:   config.PrioritizeVerb,
//		bindVerb:         config.BindVerb,
//		weight:           config.Weight,
//		schdulerClient:           schdulerClient,
//		foodCacheCapable: config.FoodCacheCapable,
//		managedResources: managedResources,
//		ignorable:        config.Ignorable,
//	}, nil
//}

// Name returns extenderURL to identify the extender.
func (h *HTTPExtender) Name() string {
	return h.extenderURL
}

// IsIgnorable returns true indicates scheduling should not fail when this extender
// is unavailable
func (h *HTTPExtender) IsIgnorable() bool {
	return h.ignorable
}

// SupportsPreemption returns true if an extender supports preemption.
// An extender should have preempt verb defined and enabled its own food cache.
func (h *HTTPExtender) SupportsPreemption() bool {
	return len(h.preemptVerb) > 0
}

// ProcessPreemption returns filtered candidate foods and victims after running preemption logic in extender.
//func (h *HTTPExtender) ProcessPreemption(
//	dguest *v1alpha1.Dguest,
//	foodNameToVictims map[string]*extenderv1.Victims,
//	foodInfos framework.FoodInfoLister,
//) (map[string]*extenderv1.Victims, error) {
//	var (
//		result extenderv1.ExtenderPreemptionResult
//		args   *extenderv1.ExtenderPreemptionArgs
//	)
//
//	if !h.SupportsPreemption() {
//		return nil, fmt.Errorf("preempt verb is not defined for extender %v but run into ProcessPreemption", h.extenderURL)
//	}
//
//	if h.foodCacheCapable {
//		// If extender has cached food info, pass FoodNameToMetaVictims in args.
//		foodNameToMetaVictims := convertToMetaVictims(foodNameToVictims)
//		args = &extenderv1.ExtenderPreemptionArgs{
//			Dguest:                dguest,
//			FoodNameToMetaVictims: foodNameToMetaVictims,
//		}
//	} else {
//		args = &extenderv1.ExtenderPreemptionArgs{
//			Dguest:            dguest,
//			FoodNameToVictims: foodNameToVictims,
//		}
//	}
//
//	if err := h.send(h.preemptVerb, args, &result); err != nil {
//		return nil, err
//	}
//
//	// Extender will always return FoodNameToMetaVictims.
//	// So let's convert it to FoodNameToVictims by using <foodInfos>.
//	newFoodNameToVictims, err := h.convertToVictims(result.FoodNameToMetaVictims, foodInfos)
//	if err != nil {
//		return nil, err
//	}
//	// Do not override <foodNameToVictims>.
//	return newFoodNameToVictims, nil
//}

// convertToVictims converts "foodNameToMetaVictims" from object identifiers,
// such as UIDs and names, to object pointers.
//func (h *HTTPExtender) convertToVictims(
//	foodNameToMetaVictims map[string]*extenderv1.MetaVictims,
//	foodInfos framework.FoodInfoLister,
//) (map[string]*extenderv1.Victims, error) {
//	foodNameToVictims := map[string]*extenderv1.Victims{}
//	for foodName, metaVictims := range foodNameToMetaVictims {
//		foodInfo, err := foodInfos.Get(foodName)
//		if err != nil {
//			return nil, err
//		}
//		victims := &extenderv1.Victims{
//			Dguests:          []*v1alpha1.Dguest{},
//			NumPDBViolations: metaVictims.NumPDBViolations,
//		}
//		for _, metaDguest := range metaVictims.Dguests {
//			dguest, err := h.convertDguestUIDToDguest(metaDguest, foodInfo)
//			if err != nil {
//				return nil, err
//			}
//			victims.Dguests = append(victims.Dguests, dguest)
//		}
//		foodNameToVictims[foodName] = victims
//	}
//	return foodNameToVictims, nil
//}

// convertDguestUIDToDguest returns v1alpha1.Dguest object for given MetaDguest and food info.
// The v1alpha1.Dguest object is restored by foodInfo.Dguests().
// It returns an error if there's cache inconsistency between default scheduler
// and extender, i.e. when the dguest is not found in foodInfo.Dguests.
//func (h *HTTPExtender) convertDguestUIDToDguest(
//	metaDguest *extenderv1.MetaDguest,
//	foodInfo *framework.FoodInfo) (*v1alpha1.Dguest, error) {
//	for _, p := range foodInfo.Dguests {
//		if string(p.Dguest.UID) == metaDguest.UID {
//			return p.Dguest, nil
//		}
//	}
//	return nil, fmt.Errorf("extender: %v claims to preempt dguest (UID: %v) on food: %v, but the dguest is not found on that food",
//		h.extenderURL, metaDguest, foodInfo.Food().Name)
//}

// convertToMetaVictims converts from struct type to meta types.
//func convertToMetaVictims(
//	foodNameToVictims map[string]*extenderv1.Victims,
//) map[string]*extenderv1.MetaVictims {
//	foodNameToMetaVictims := map[string]*extenderv1.MetaVictims{}
//	for food, victims := range foodNameToVictims {
//		metaVictims := &extenderv1.MetaVictims{
//			Dguests:          []*extenderv1.MetaDguest{},
//			NumPDBViolations: victims.NumPDBViolations,
//		}
//		for _, dguest := range victims.Dguests {
//			metaDguest := &extenderv1.MetaDguest{
//				UID: string(dguest.UID),
//			}
//			metaVictims.Dguests = append(metaVictims.Dguests, metaDguest)
//		}
//		foodNameToMetaVictims[food] = metaVictims
//	}
//	return foodNameToMetaVictims
//}

// Filter based on extender implemented predicate functions. The filtered list is
// expected to be a subset of the supplied list; otherwise the function returns an error.
// The failedFoods and failedAndUnresolvableFoods optionally contains the list
// of failed foods and failure reasons, except foods in the latter are
// unresolvable.
//func (h *HTTPExtender) Filter(
//	dguest *v1alpha1.Dguest,
//	foods []*v1alpha1.Food,
//) (filteredList []*v1alpha1.Food, failedFoods, failedAndUnresolvableFoods extenderv1.FailedFoodsMap, err error) {
//	var (
//		result     extenderv1.ExtenderFilterResult
//		foodList   *v1alpha1.FoodList
//		foodNames  *[]string
//		foodResult []*v1alpha1.Food
//		args       *extenderv1.ExtenderArgs
//	)
//	fromFoodName := make(map[string]*v1alpha1.Food)
//	for _, n := range foods {
//		fromFoodName[n.Name] = n
//	}
//
//	if h.filterVerb == "" {
//		return foods, extenderv1.FailedFoodsMap{}, extenderv1.FailedFoodsMap{}, nil
//	}
//
//	if h.foodCacheCapable {
//		foodNameSlice := make([]string, 0, len(foods))
//		for _, food := range foods {
//			foodNameSlice = append(foodNameSlice, food.Name)
//		}
//		foodNames = &foodNameSlice
//	} else {
//		foodList = &v1alpha1.FoodList{}
//		for _, food := range foods {
//			foodList.Items = append(foodList.Items, *food)
//		}
//	}
//
//	args = &extenderv1.ExtenderArgs{
//		Dguest:    dguest,
//		Foods:     foodList,
//		FoodNames: foodNames,
//	}
//
//	if err := h.send(h.filterVerb, args, &result); err != nil {
//		return nil, nil, nil, err
//	}
//	if result.Error != "" {
//		return nil, nil, nil, fmt.Errorf(result.Error)
//	}
//
//	if h.foodCacheCapable && result.FoodNames != nil {
//		foodResult = make([]*v1alpha1.Food, len(*result.FoodNames))
//		for i, foodName := range *result.FoodNames {
//			if n, ok := fromFoodName[foodName]; ok {
//				foodResult[i] = n
//			} else {
//				return nil, nil, nil, fmt.Errorf(
//					"extender %q claims a filtered food %q which is not found in the input food list",
//					h.extenderURL, foodName)
//			}
//		}
//	} else if result.Foods != nil {
//		foodResult = make([]*v1alpha1.Food, len(result.Foods.Items))
//		for i := range result.Foods.Items {
//			foodResult[i] = &result.Foods.Items[i]
//		}
//	}
//
//	return foodResult, result.FailedFoods, result.FailedAndUnresolvableFoods, nil
//}

// Prioritize based on extender implemented priority functions. Weight*priority is added
// up for each such priority function. The returned score is added to the score computed
// by Kubernetes scheduler. The total score is used to do the host selection.
//func (h *HTTPExtender) Prioritize(dguest *v1alpha1.Dguest, foods []*v1alpha1.Food) (*extenderv1.HostPriorityList, int64, error) {
//	var (
//		result    extenderv1.HostPriorityList
//		foodList  *v1alpha1.FoodList
//		foodNames *[]string
//		args      *extenderv1.ExtenderArgs
//	)
//
//	if h.prioritizeVerb == "" {
//		result := extenderv1.HostPriorityList{}
//		for _, food := range foods {
//			result = append(result, extenderv1.HostPriority{Host: food.Name, Score: 0})
//		}
//		return &result, 0, nil
//	}
//
//	if h.foodCacheCapable {
//		foodNameSlice := make([]string, 0, len(foods))
//		for _, food := range foods {
//			foodNameSlice = append(foodNameSlice, food.Name)
//		}
//		foodNames = &foodNameSlice
//	} else {
//		foodList = &v1alpha1.FoodList{}
//		for _, food := range foods {
//			foodList.Items = append(foodList.Items, *food)
//		}
//	}
//
//	args = &extenderv1.ExtenderArgs{
//		Dguest:    dguest,
//		Foods:     foodList,
//		FoodNames: foodNames,
//	}
//
//	if err := h.send(h.prioritizeVerb, args, &result); err != nil {
//		return nil, 0, err
//	}
//	return &result, h.weight, nil
//}

// Bind delegates the action of binding a dguest to a food to the extender.
//func (h *HTTPExtender) Bind(binding *v1.Binding) error {
//	var result extenderv1.ExtenderBindingResult
//	if !h.IsBinder() {
//		// This shouldn't happen as this extender wouldn't have become a Binder.
//		return fmt.Errorf("unexpected empty bindVerb in extender")
//	}
//	req := &extenderv1.ExtenderBindingArgs{
//		DguestName:      binding.Name,
//		DguestNamespace: binding.Namespace,
//		DguestUID:       binding.UID,
//		Food:            binding.Target.Name,
//	}
//	if err := h.send(h.bindVerb, req, &result); err != nil {
//		return err
//	}
//	if result.Error != "" {
//		return fmt.Errorf(result.Error)
//	}
//	return nil
//}

// IsBinder returns whether this extender is configured for the Bind method.
func (h *HTTPExtender) IsBinder() bool {
	return h.bindVerb != ""
}

// Helper function to send messages to the extender
func (h *HTTPExtender) send(action string, args interface{}, result interface{}) error {
	out, err := json.Marshal(args)
	if err != nil {
		return err
	}

	url := strings.TrimRight(h.extenderURL, "/") + "/" + action

	req, err := http.NewRequest("POST", url, bytes.NewReader(out))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed %v with extender at URL %v, code %v", action, url, resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// IsInterested returns true if at least one extended resource requested by
// this dguest is managed by this extender.
//func (h *HTTPExtender) IsInterested(dguest *v1alpha1.Dguest) bool {
//	if h.managedResources.Len() == 0 {
//		return true
//	}
//	if h.hasManagedResources(dguest.Spec.Containers) {
//		return true
//	}
//	if h.hasManagedResources(dguest.Spec.InitContainers) {
//		return true
//	}
//	return false
//}
//
//func (h *HTTPExtender) hasManagedResources(containers []v1.Container) bool {
//	for i := range containers {
//		container := &containers[i]
//		for resourceName := range container.Resources.Requests {
//			if h.managedResources.Has(string(resourceName)) {
//				return true
//			}
//		}
//		for resourceName := range container.Resources.Limits {
//			if h.managedResources.Has(string(resourceName)) {
//				return true
//			}
//		}
//	}
//	return false
//}
