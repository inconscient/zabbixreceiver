package zabbixreceiver

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
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

func (zr *zabbixReceiver) handleConnection0(conn net.Conn) {
	defer conn.Close()

	// Read response
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Println("Error reading from connection:", err)
		return
	}

	log.Println("Received Message:", string(buffer[:n]))

	//decoder := json.NewDecoder(conn)

	for {
		var msg Metric
		var data string = string(buffer[:n])
		trimmedData := data[1 : len(data)-1] // Remove first and last characters

		log.Printf("message trimmed: %s", trimmedData)

		//if err := decoder.Decode(&msg); err != nil {
		if err := json.Unmarshal([]byte(trimmedData), &msg); err != nil {
			// Log message error
			//log.Printf("received wrong message : %s", err)
			log.Printf("verbose error info: %#v", err)
			return
		}

		// Log message received to console
		log.Printf("received message from host: %s", msg.Host)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().PutStr("host", msg.Name)

		sm := rm.ScopeMetrics().AppendEmpty()
		m := sm.Metrics().AppendEmpty()
		m.SetName(strconv.Itoa(int(msg.ItemID)))
		dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
		dp.SetIntValue(msg.Value)
		dp.SetTimestamp(pcommon.Timestamp(time.Unix(int64(msg.Clock), 0).UnixNano()))

		_ = zr.consumer.ConsumeMetrics(context.Background(), metrics)
	}
}

func (zr *zabbixReceiver) handleConnection(conn net.Conn) {
	//defer conn.Close()

	// Wrap the connection with a bufio.Reader
	reader := bufio.NewReader(conn)
	// Parse the HTTP request
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Println("Error reading request:", err)
		return
	}

	// Read the body
	body, err := io.ReadAll(req.Body)

	if err != nil {
		log.Println("Error reading body:", err)
		return
	}
	req.Body.Close() // Close the body after reading

	//log.Printf("upper body: %s", body)

	for {
		var msg Metric
		var data string = string(body)
		trimmedData := data[1 : len(data)-1] // Remove first and last characters
		log.Printf("message trimmed: %s", trimmedData)
		//if err := decoder.Decode(&msg); err != nil {
		//if err := json.Unmarshal([]byte(body), &msg); err != nil {
		if err := json.NewDecoder(strings.NewReader(string(body))).Decode(&msg); err != nil {
			log.Printf("verbose error info: %#v", err)
			return
		}

		// Log message received to console
		//log.Printf("received message: %s", trimmedData)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().PutStr("host", msg.Name)

		sm := rm.ScopeMetrics().AppendEmpty()
		m := sm.Metrics().AppendEmpty()
		m.SetName(strconv.Itoa(int(msg.ItemID)))
		dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
		dp.SetIntValue(msg.Value)
		dp.SetTimestamp(pcommon.Timestamp(time.Unix(int64(msg.Clock), 0).UnixNano()))
		log.Printf("Metrics Object: %s", rm.Resource().Attributes())
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
