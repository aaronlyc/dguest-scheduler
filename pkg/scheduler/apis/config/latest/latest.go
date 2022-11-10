/*
Copyright 2021 The Kubernetes Authors.

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

package latest

import (
	"dguest-scheduler/pkg/scheduler/apis/config"
	"dguest-scheduler/pkg/scheduler/apis/config/scheme"
	componentbaseconfig "k8s.io/component-base/config"
)

// Default creates a default configuration of the latest versioned type.
// This function needs to be updated whenever we bump the scheduler's component config version.
func Default() (*config.SchedulerConfiguration, error) {
	versionedCfg := config.SchedulerConfiguration{}
	versionedCfg.DebuggingConfiguration = componentbaseconfig.DebuggingConfiguration{
		EnableProfiling: true,
	}

	scheme.Scheme.Default(&versionedCfg)
	cfg := config.SchedulerConfiguration{}
	if err := scheme.Scheme.Convert(&versionedCfg, &cfg, nil); err != nil {
		return nil, err
	}
	// We don't set this field in pkg/scheduler/apis/config/{version}/conversion.go
	// because the field will be cleared later by API machinery during
	// conversion. See SchedulerConfiguration internal type definition for
	// more details.
	cfg.TypeMeta.APIVersion = config.SchemeGroupVersion.String()
	return &cfg, nil
}