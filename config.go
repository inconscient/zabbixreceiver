package zabbixreceiver

import "go.opentelemetry.io/collector/config/configmodels"

type Config struct {
    configmodels.ReceiverSettings `mapstructure:",squash"`
    Endpoint                      string `mapstructure:"endpoint"`
}