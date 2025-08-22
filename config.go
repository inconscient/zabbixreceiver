package zabbixreceiver

import "go.opentelemetry.io/collector/component"

type Config struct {
	component.Config
	Endpoint string `mapstructure:"endpoint"`
}
