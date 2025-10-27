package zabbixreceiver

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
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

func (zr *zabbixReceiver) handleConnection(conn net.Conn) {

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

	log.Printf("req body: %s", body)
	saveDebug(string(body))

	for {
		var msg Metric

		if err := json.NewDecoder(strings.NewReader(string(body))).Decode(&msg); err != nil {
			log.Printf("verbose error info: %#v", err)
			return
		}

		// Log message received to console
		//log.Printf("decoded message: %s", msg)

		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().PutStr("host", msg.Host.Name)

		sm := rm.ScopeMetrics().AppendEmpty()
		m := sm.Metrics().AppendEmpty()
		m.SetName(msg.Name)
		dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
		dp.SetDoubleValue(msg.Value)
		dp.SetTimestamp(pcommon.Timestamp(time.Unix(int64(msg.Clock), 0).UnixNano()))
		//log.Printf("Metrics Object: %s", rm.Resource().Attributes())
		_ = zr.consumer.ConsumeMetrics(context.Background(), metrics)
	}
}

func saveDebug(debugcontent string) {
	// Create a new file or open an existing one for writing
	filePath := "debug.txt"

	// Open the file in append mode, create if it doesn't exist, and open for writing only.
	// 0644 are the Unix-style permissions for the file.
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close() // Ensure the file is closed when the function exits

	// Write a string to the file
	_, err = file.WriteString(debugcontent)

	if err != nil {
		fmt.Println("Error writing to file:", err)
	}

	fmt.Println("Data written successfully.")
}

func parseInt(val string) int64 {
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}
	return i
}
