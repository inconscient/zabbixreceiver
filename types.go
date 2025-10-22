package zabbixreceiver

type ZabbixMessage struct {
	Host      string `json:"host"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}
