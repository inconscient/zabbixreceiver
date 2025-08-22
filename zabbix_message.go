package zabbixreceiver

type ZabbixMessage struct {
	Timestamp int64  `json:"timestamp"`
	Host      string `json:"host"`
	Key       string `json:"key"`
	Value     string `json:"value"`
}
