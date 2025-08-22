package zabbixreceiver

import (
	"context"
	"encoding/json"
	"net"
	"strconv"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/xpdata"
)

type zabbixReceiver struct {
	cfg      *Config
	consumer consumer.Metrics
}

func createMetricsReceiver(
	_ context.Context,
	_ receiver.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (receiver.Metrics, error) {
	return &zabbixReceiver{
		cfg:      cfg.(*Config),
		consumer: nextConsumer,
	}, nil
}

func (zr *zabbixReceiver) Start(_ context.Context, _ component.Host) error {
	ln, err := net.Listen("tcp", zr.cfg.Endpoint)
	if err != nil {
		return err
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

func (zr *zabbixReceiver) Shutdown(_ context.Context) error {
	return nil
}

func (zr *zabbixReceiver) handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)

	for {
		var msg ZabbixMessage
		if err := decoder.Decode(&msg); err != nil {
			return
		}

		metrics := xpdata.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().PutString("host", msg.Host)

		sm := rm.ScopeMetrics().AppendEmpty()
		m := sm.Metrics().AppendEmpty()
		m.SetName(msg.Key)
		dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
		dp.SetIntValue(parseInt(msg.Value))
		dp.SetTimestamp(xpdata.NewTimestampFromTime(time.Unix(msg.Timestamp, 0)))

		_ = zr.consumer.ConsumeMetrics(context.Background(), metrics)
	}
}

func parseInt(val string) int64 {
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}
	return i
}
