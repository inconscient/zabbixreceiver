package zabbixreceiver

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/otelcol/metricdata"
	"go.opentelemetry.io/collector/receiver"
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
	rCfg := cfg.(*Config)
	return &zabbixReceiver{
		cfg:      rCfg,
		consumer: nextConsumer,
	}, nil
}

func (zr *zabbixReceiver) Start(_ context.Context, _ component.Host) error {
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

func (zr *zabbixReceiver) Shutdown(_ context.Context) error {
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

func (zr *zabbixReceiver) convertToMetrics(msg ZabbixMessage) metricdata.ResourceMetrics {
	// You will need to use the new metricdata API to construct metrics
	// This is a placeholder for actual metric construction
	return metricdata.ResourceMetrics{}
}

func parseInt(val string) int64 {
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}
	return i
}
