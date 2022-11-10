/*
Copyright 2020 The Kubernetes Authors.

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

package profile

import (
	"context"
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"dguest-scheduler/pkg/scheduler/apis/config/v1"
	"fmt"
	"strings"
	"testing"

	"dguest-scheduler/pkg/scheduler/framework"
	frameworkruntime "dguest-scheduler/pkg/scheduler/framework/runtime"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/events"
)

var fakeRegistry = frameworkruntime.Registry{
	"QueueSort": newFakePlugin("QueueSort"),
	"Bind1":     newFakePlugin("Bind1"),
	"Bind2":     newFakePlugin("Bind2"),
	"Another":   newFakePlugin("Another"),
}

func TestNewMap(t *testing.T) {
	cases := []struct {
		name    string
		cfgs    []v1.SchedulerProfile
		wantErr string
	}{
		{
			name: "valid",
			cfgs: []v1.SchedulerProfile{
				{
					SchedulerName: "profile-1",
					Plugins: &v1.Plugins{
						QueueSort: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "QueueSort"},
							},
						},
						Bind: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "Bind1"},
							},
						},
					},
				},
				{
					SchedulerName: "profile-2",
					Plugins: &v1.Plugins{
						QueueSort: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "QueueSort"},
							},
						},
						Bind: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "Bind2"},
							},
						},
					},
					PluginConfig: []v1.PluginConfig{
						{
							Name: "Bind2",
							Args: &runtime.Unknown{Raw: []byte("{}")},
						},
					},
				},
			},
		},
		{
			name: "different queue sort",
			cfgs: []v1.SchedulerProfile{
				{
					SchedulerName: "profile-1",
					Plugins: &v1.Plugins{
						QueueSort: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "QueueSort"},
							},
						},
						Bind: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "Bind1"},
							},
						},
					},
				},
				{
					SchedulerName: "profile-2",
					Plugins: &v1.Plugins{
						QueueSort: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "Another"},
							},
						},
						Bind: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "Bind2"},
							},
						},
					},
				},
			},
			wantErr: "different queue sort plugins",
		},
		{
			name: "different queue sort args",
			cfgs: []v1.SchedulerProfile{
				{
					SchedulerName: "profile-1",
					Plugins: &v1.Plugins{
						QueueSort: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "QueueSort"},
							},
						},
						Bind: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "Bind1"},
							},
						},
					},
					PluginConfig: []v1.PluginConfig{
						{
							Name: "QueueSort",
							Args: &runtime.Unknown{Raw: []byte("{}")},
						},
					},
				},
				{
					SchedulerName: "profile-2",
					Plugins: &v1.Plugins{
						QueueSort: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "QueueSort"},
							},
						},
						Bind: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "Bind2"},
							},
						},
					},
				},
			},
			wantErr: "different queue sort plugin args",
		},
		{
			name: "duplicate scheduler name",
			cfgs: []v1.SchedulerProfile{
				{
					SchedulerName: "profile-1",
					Plugins: &v1.Plugins{
						QueueSort: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "QueueSort"},
							},
						},
						Bind: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "Bind1"},
							},
						},
					},
				},
				{
					SchedulerName: "profile-1",
					Plugins: &v1.Plugins{
						QueueSort: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "QueueSort"},
							},
						},
						Bind: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "Bind2"},
							},
						},
					},
				},
			},
			wantErr: "duplicate profile",
		},
		{
			name: "scheduler name is needed",
			cfgs: []v1.SchedulerProfile{
				{
					Plugins: &v1.Plugins{
						QueueSort: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "QueueSort"},
							},
						},
						Bind: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "Bind1"},
							},
						},
					},
				},
			},
			wantErr: "scheduler name is needed",
		},
		{
			name: "plugins required for profile",
			cfgs: []v1.SchedulerProfile{
				{
					SchedulerName: "profile-1",
				},
			},
			wantErr: "plugins required for profile",
		},
		{
			name: "invalid framework configuration",
			cfgs: []v1.SchedulerProfile{
				{
					SchedulerName: "invalid-profile",
					Plugins: &v1.Plugins{
						QueueSort: v1.PluginSet{
							Enabled: []v1.Plugin{
								{Name: "QueueSort"},
							},
						},
					},
				},
			},
			wantErr: "at least one bind plugin is needed",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stopCh := make(chan struct{})
			defer close(stopCh)
			m, err := NewMap(tc.cfgs, fakeRegistry, nilRecorderFactory, stopCh)
			if err := checkErr(err, tc.wantErr); err != nil {
				t.Fatal(err)
			}
			if len(tc.wantErr) != 0 {
				return
			}
			if len(m) != len(tc.cfgs) {
				t.Errorf("got %d profiles, want %d", len(m), len(tc.cfgs))
			}
		})
	}
}

type fakePlugin struct {
	name string
}

func (p *fakePlugin) Name() string {
	return p.name
}

func (p *fakePlugin) Less(*framework.QueuedDguestInfo, *framework.QueuedDguestInfo) bool {
	return false
}

func (p *fakePlugin) Bind(context.Context, *framework.CycleState, *v1alpha1.Dguest, string) *framework.Status {
	return nil
}

func newFakePlugin(name string) func(object runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return func(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
		return &fakePlugin{name: name}, nil
	}
}

func nilRecorderFactory(_ string) events.EventRecorder {
	return nil
}

func checkErr(err error, wantErr string) error {
	if len(wantErr) == 0 {
		return err
	}
	if err == nil || !strings.Contains(err.Error(), wantErr) {
		return fmt.Errorf("got error %q, want %q", err, wantErr)
	}
	return nil
}
