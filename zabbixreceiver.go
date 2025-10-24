package zabbixreceiver

import (
	"context"
	"log"
	"net"
	"strconv"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
)

type zabbixReceiver struct {
	cfg      *Config
	consumer consumer.Metrics
}

func createMetricsReceiver(
	_ context.Context,
	set receiver.Settings,
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

	// Log successful start to console
	log.Printf("zabbixreceiver started and listening on %s", zr.cfg.Endpoint)

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

		// Log message received to console
		log.Printf("received message from host: %s", msg.Host)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().PutStr("host", msg.Host)

		sm := rm.ScopeMetrics().AppendEmpty()
		m := sm.Metrics().AppendEmpty()
		m.SetName(msg.Key)
		dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
		dp.SetIntValue(parseInt(msg.Value))
		dp.SetTimestamp(pcommon.Timestamp(time.Unix(msg.Timestamp, 0).UnixNano()))

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
