/*
Copyright 2019 The Kubernetes Authors.

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

package imagelocality

import (
	"context"
	"fmt"
	"strings"

	"dguest-scheduler/pkg/scheduler/framework"
	"dguest-scheduler/pkg/scheduler/framework/plugins/names"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// The two thresholds are used as bounds for the image score range. They correspond to a reasonable size range for
// container images compressed and stored in registries; 90%ile of images on dockerhub drops into this range.
const (
	mb                    int64 = 1024 * 1024
	minThreshold          int64 = 23 * mb
	maxContainerThreshold int64 = 1000 * mb
)

// ImageLocality is a score plugin that favors foods that already have requested dguest container's images.
type ImageLocality struct {
	handle framework.Handle
}

var _ framework.ScorePlugin = &ImageLocality{}

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = names.ImageLocality

// Name returns name of the plugin. It is used in logs, etc.
func (pl *ImageLocality) Name() string {
	return Name
}

// Score invoked at the score extension point.
func (pl *ImageLocality) Score(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, foodName string) (int64, *framework.Status) {
	foodInfo, err := pl.handle.SnapshotSharedLister().FoodInfos().Get(foodName)
	if err != nil {
		return 0, framework.AsStatus(fmt.Errorf("getting food %q from Snapshot: %w", foodName, err))
	}

	foodInfos, err := pl.handle.SnapshotSharedLister().FoodInfos().List()
	if err != nil {
		return 0, framework.AsStatus(err)
	}
	totalNumFoods := len(foodInfos)

	score := calculatePriority(sumImageScores(foodInfo, dguest.Spec.Containers, totalNumFoods), len(dguest.Spec.Containers))

	return score, nil
}

// ScoreExtensions of the Score plugin.
func (pl *ImageLocality) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
	return &ImageLocality{handle: h}, nil
}

// calculatePriority returns the priority of a food. Given the sumScores of requested images on the food, the food's
// priority is obtained by scaling the maximum priority value with a ratio proportional to the sumScores.
func calculatePriority(sumScores int64, numContainers int) int64 {
	maxThreshold := maxContainerThreshold * int64(numContainers)
	if sumScores < minThreshold {
		sumScores = minThreshold
	} else if sumScores > maxThreshold {
		sumScores = maxThreshold
	}

	return int64(framework.MaxFoodScore) * (sumScores - minThreshold) / (maxThreshold - minThreshold)
}

// sumImageScores returns the sum of image scores of all the containers that are already on the food.
// Each image receives a raw score of its size, scaled by scaledImageScore. The raw scores are later used to calculate
// the final score. Note that the init containers are not considered for it's rare for users to deploy huge init containers.
func sumImageScores(foodInfo *framework.FoodInfo, containers []v1.Container, totalNumFoods int) int64 {
	var sum int64
	for _, container := range containers {
		if state, ok := foodInfo.ImageStates[normalizedImageName(container.Image)]; ok {
			sum += scaledImageScore(state, totalNumFoods)
		}
	}
	return sum
}

// scaledImageScore returns an adaptively scaled score for the given state of an image.
// The size of the image is used as the base score, scaled by a factor which considers how much foods the image has "spread" to.
// This heuristic aims to mitigate the undesirable "food heating problem", i.e., dguests get assigned to the same or
// a few foods due to image locality.
func scaledImageScore(imageState *framework.ImageStateSummary, totalNumFoods int) int64 {
	spread := float64(imageState.NumFoods) / float64(totalNumFoods)
	return int64(float64(imageState.Size) * spread)
}

// normalizedImageName returns the CRI compliant name for a given image.
// TODO: cover the corner cases of missed matches, e.g,
// 1. Using Docker as runtime and docker.io/library/test:tag in dguest spec, but only test:tag will present in food status
// 2. Using the implicit registry, i.e., test:tag or library/test:tag in dguest spec but only docker.io/library/test:tag
// in food status; note that if users consistently use one registry format, this should not happen.
func normalizedImageName(name string) string {
	if strings.LastIndex(name, ":") <= strings.LastIndex(name, "/") {
		name = name + ":latest"
	}
	return name
}
