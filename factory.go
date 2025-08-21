package zabbixreceiver

import (
    "context"
    "go.opentelemetry.io/collector/component"
    "go.opentelemetry.io/collector/receiver/receiverhelper"
)

const (
    typeStr = "zabbixreceiver"
)

func NewFactory() component.ReceiverFactory {
    return receiverhelper.NewFactory(
        typeStr,
        createDefaultConfig,
        receiverhelper.WithMetrics(createMetricsReceiver),
    )
}

func createDefaultConfig() component.Config {
    return &Config{
        ReceiverSettings: configmodels.ReceiverSettings{
            TypeVal: typeStr,
            NameVal: typeStr,
        },
        Endpoint: ":10051",
    }
}