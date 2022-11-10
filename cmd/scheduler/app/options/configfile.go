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

package options

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	utilyaml "k8s.io/apimachinery/pkg/util/yaml"

	"dguest-scheduler/pkg/scheduler/apis/config"
	"dguest-scheduler/pkg/scheduler/apis/config/scheme"
	configv1 "dguest-scheduler/pkg/scheduler/apis/config/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

func loadConfigFromFile(file string) (*config.SchedulerConfiguration, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return loadConfig(data)
}

func loadConfig(data []byte) (*config.SchedulerConfiguration, error) {
	kc, err := decodeConfig(data)
	if err != nil {
		return nil, err
	}

	configv1.SetDefaults_KubeSchedulerConfiguration(kc)
	return kc, nil

	// The UniversalDecoder runs defaulting and returns the internal type by default.
	//decoder := scheme.Codecs.UniversalDecoder()
	//obj, gvk, err := decoder.Decode(data, &config.GVK, nil)
	//if err != nil {
	//	return nil, err
	//}
	//if cfgObj, ok := obj.(*config.SchedulerConfiguration); ok {
	//	// We don't set this field in pkg/scheduler/apis/config/{version}/conversion.go
	//	// because the field will be cleared later by API machinery during
	//	// conversion. See SchedulerConfiguration internal type definition for
	//	// more details.
	//	cfgObj.TypeMeta.APIVersion = gvk.GroupVersion().String()
	//	if cfgObj.TypeMeta.APIVersion == config.SchemeGroupVersion.String() {
	//		klog.InfoS("SchedulerConfiguration v1beta2 is deprecated in v1.25, will be removed in v1.26")
	//	}
	//	return cfgObj, nil
	//}
	//return nil, fmt.Errorf("couldn't decode as SchedulerConfiguration, got %s: ", gvk)
}

func decodeConfig(data []byte) (*config.SchedulerConfiguration, error) {
	jsondata, err := utilyaml.ToJSON(data)
	if err != nil {
		return nil, err
	}

	result := config.SchedulerConfiguration{}
	err = json.Unmarshal(jsondata, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func encodeConfig(cfg *config.SchedulerConfiguration) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	const mediaType = runtime.ContentTypeYAML
	info, ok := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		return buf, fmt.Errorf("unable to locate encoder -- %q is not a supported media type", mediaType)
	}

	var encoder runtime.Encoder
	switch cfg.TypeMeta.APIVersion {
	case configv1.SchemeGroupVersion.String():
		encoder = scheme.Codecs.EncoderForVersion(info.Serializer, configv1.SchemeGroupVersion)
	default:
		encoder = scheme.Codecs.EncoderForVersion(info.Serializer, configv1.SchemeGroupVersion)
	}
	if err := encoder.Encode(cfg, buf); err != nil {
		return buf, err
	}
	return buf, nil
}

// LogOrWriteConfig logs the completed component config and writes it into the given file name as YAML, if either is enabled
func LogOrWriteConfig(fileName string, cfg *config.SchedulerConfiguration, completedProfiles []config.KubeSchedulerProfile) error {
	klogV := klog.V(0)
	if !klogV.Enabled() && len(fileName) == 0 {
		return nil
	}
	cfg.Profiles = completedProfiles

	buf, err := encodeConfig(cfg)
	if err != nil {
		return err
	}

	if klogV.Enabled() {
		klogV.InfoS("Using component config", "config", buf.String())
	}

	if len(fileName) > 0 {
		configFile, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer configFile.Close()
		if _, err := io.Copy(configFile, buf); err != nil {
			return err
		}
		klog.InfoS("Wrote configuration", "file", fileName)
		os.Exit(0)
	}
	return nil
}
