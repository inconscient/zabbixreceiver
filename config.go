package zabbixreceiver

import "go.opentelemetry.io/collector/config"

type Config struct {
	config.ReceiverSettings `mapstructure:",squash"`
	Endpoint                string `mapstructure:"endpoint"`
}
