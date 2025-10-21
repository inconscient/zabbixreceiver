package zabbixreceiver

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
)

const typeStr = "zabbixreceiver"

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType(typeStr),
		func() component.Config {
			return &Config{Endpoint: ":54319"}
		},
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelAlpha),
	)
}
