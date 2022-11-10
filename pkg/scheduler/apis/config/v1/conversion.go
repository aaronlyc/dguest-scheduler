package v1

import (
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var (
	// pluginArgConversionScheme is a scheme with internal and v1 registered,
	// used for defaulting/converting typed PluginConfig Args.
	// Access via getPluginArgConversionScheme()
	pluginArgConversionScheme     *runtime.Scheme
	initPluginArgConversionScheme sync.Once
)

func GetPluginArgConversionScheme() *runtime.Scheme {
	initPluginArgConversionScheme.Do(func() {
		// set up the scheme used for plugin arg conversion
		pluginArgConversionScheme = runtime.NewScheme()
		utilruntime.Must(AddToScheme(pluginArgConversionScheme))
	})
	return pluginArgConversionScheme
}

//func Convert_v1_KubeSchedulerConfiguration_To_config_KubeSchedulerConfiguration(in *v1.KubeSchedulerConfiguration, out *config.SchedulerConfiguration, s conversion.Scope) error {
//	if err := autoConvert_v1_KubeSchedulerConfiguration_To_config_KubeSchedulerConfiguration(in, out, s); err != nil {
//		return err
//	}
//	return convertToInternalPluginConfigArgs(out)
//}

// convertToInternalPluginConfigArgs converts PluginConfig#Args into internal
// types using a scheme, after applying defaults.
func convertToInternalPluginConfigArgs(out *SchedulerConfiguration) error {
	scheme := GetPluginArgConversionScheme()
	for i := range out.Profiles {
		prof := &out.Profiles[i]
		for j := range prof.PluginConfig {
			args := prof.PluginConfig[j].Args
			if args == nil {
				continue
			}
			if _, isUnknown := args.(*runtime.Unknown); isUnknown {
				continue
			}
			internalArgs, err := scheme.ConvertToVersion(args, SchemeGroupVersion)
			if err != nil {
				return fmt.Errorf("converting .Profiles[%d].PluginConfig[%d].Args into internal type: %w", i, j, err)
			}
			prof.PluginConfig[j].Args = internalArgs
		}
	}
	return nil
}

//func Convert_config_KubeSchedulerConfiguration_To_v1_KubeSchedulerConfiguration(in *config.SchedulerConfiguration, out *v1.KubeSchedulerConfiguration, s conversion.Scope) error {
//	if err := autoConvert_config_KubeSchedulerConfiguration_To_v1_KubeSchedulerConfiguration(in, out, s); err != nil {
//		return err
//	}
//	return convertToExternalPluginConfigArgs(out)
//}

// convertToExternalPluginConfigArgs converts PluginConfig#Args into
// external (versioned) types using a scheme.
//func convertToExternalPluginConfigArgs(out *v1.KubeSchedulerConfiguration) error {
//	scheme := GetPluginArgConversionScheme()
//	for i := range out.Profiles {
//		for j := range out.Profiles[i].PluginConfig {
//			args := out.Profiles[i].PluginConfig[j].Args
//			if args.Object == nil {
//				continue
//			}
//			if _, isUnknown := args.Object.(*runtime.Unknown); isUnknown {
//				continue
//			}
//			externalArgs, err := scheme.ConvertToVersion(args.Object, SchemeGroupVersion)
//			if err != nil {
//				return err
//			}
//			out.Profiles[i].PluginConfig[j].Args.Object = externalArgs
//		}
//	}
//	return nil
//}
