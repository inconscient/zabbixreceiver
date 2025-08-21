package zabbixreceiver

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "net"
    "strconv"
    "time"

    "go.opentelemetry.io/collector/component"
    "go.opentelemetry.io/collector/consumer"
    "go.opentelemetry.io/collector/pdata/pmetric"
)

type zabbixReceiver struct {
    cfg      *Config
    consumer consumer.Metrics
}

func createMetricsReceiver(
    ctx context.Context,
    params component.ReceiverCreateSettings,
    cfg component.Config,
    nextConsumer consumer.Metrics,
) (component.MetricsReceiver, error) {
    rCfg := cfg.(*Config)
    return &zabbixReceiver{
        cfg:      rCfg,
        consumer: nextConsumer,
    }, nil
}

func (zr *zabbixReceiver) Start(ctx context.Context, host component.Host) error {
    ln, err := net.Listen("tcp", zr.cfg.Endpoint)
    if err != nil {
        return fmt.Errorf("failed to listen on %s: %w", zr.cfg.Endpoint, err)
    }

    go func() {
        for {
            conn, err := ln.Accept()
            if err != nil {
                continue
            }
            go zr.handleConnection(conn)
        }
    }()

    return nil
}

func (zr *zabbixReceiver) Shutdown(ctx context.Context) error {
    return nil
}

func (zr *zabbixReceiver) handleConnection(conn net.Conn) {
    defer conn.Close()
    scanner := bufio.NewScanner(conn)

    for scanner.Scan() {
        line := scanner.Text()
        var msg ZabbixMessage
        if err := json.Unmarshal([]byte(line), &msg); err != nil {
            continue
        }

        metrics := zr.convertToMetrics(msg)
        _ = zr.consumer.ConsumeMetrics(context.Background(), metrics)
    }
}

func (zr *zabbixReceiver) convertToMetrics(msg ZabbixMessage) pmetric.Metrics {
    md := pmetric.NewMetrics()
    rm := md.ResourceMetrics().AppendEmpty()
    rm.Resource().Attributes().PutString("host", msg.Host)

    sm := rm.ScopeMetrics().AppendEmpty()
    metric := sm.Metrics().AppendEmpty()
    metric.SetName(msg.Key)
    metric.SetEmptyGauge().DataPoints().AppendEmpty().SetIntValue(parseInt(msg.Value))
    metric.Gauge().DataPoints().At(0).SetTimestamp(pmetric.NewTimestampFromTime(time.Unix(msg.Timestamp, 0)))

    return md
}

func parseInt(val string) int64 {
    i, err := strconv.ParseInt(val, 10, 64)
    if err != nil {
        return 0
    }
    return i
}