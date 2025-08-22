package zabbixreceiver

import "go.opentelemetry.io/collector/config"

// Config defines configuration for the Zabbix receiver.
type Config struct {
	config.ReceiverSettings `mapstructure:",squash"`
	Endpoint                string `mapstructure:"endpoint"`
}
