package zabbixreceiver

type Host struct {
	Host string `json:"host"`
	Name string `json:"name"`
}

type ItemTag struct {
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

type Metric struct {
	Host     Host      `json:"host"`
	Groups   []string  `json:"groups"`
	ItemTags []ItemTag `json:"item_tags"`
	ItemID   int64     `json:"itemid"`
	Name     string    `json:"name"`
	Clock    int64     `json:"clock"`
	Ns       *int64    `json:"ns"`
	Value    int64     `json:"value"`
	Type     *int      `json:"type"`
}
