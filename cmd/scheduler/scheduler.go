package main

import (
	"dguest-scheduler/pkg/scheduler/framework/plugins/ac"
	"dguest-scheduler/pkg/scheduler/framework/plugins/af"
	"os"

	"dguest-scheduler/cmd/scheduler/app"
	"k8s.io/component-base/cli"
	_ "k8s.io/component-base/logs/json/register" // for JSON log format registration
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	_ "k8s.io/component-base/metrics/prometheus/version" // for version metric registration
)

// 配置文件启动参数为：--config=./cmd/scheduler/setup.yaml
func main() {
	command := app.NewSchedulerCommand(app.WithPlugin(ac.Name, ac.New), app.WithPlugin(af.Name, af.New))
	code := cli.Run(command)
	os.Exit(code)
}
