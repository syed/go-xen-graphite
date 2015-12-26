package xengraphite

import (
	"encoding/xml"
	log "github.com/Sirupsen/logrus"
	"strconv"
	"strings"
)

type metrics struct {
	XMLName xml.Name `xml:"xport"`
	NumRows string   `xml:"meta>rows"`
	NumCols string   `xml:"meta>columns"`
	Entries []string `xml:"meta>legend>entry"`
	Rows    []row    `xml:"data>row"`
}

type row struct {
	Timestamp string   `xml:"t"`
	Values    []string `xml:"v"`
}

func ParseRrdMetrics(xml_data []byte) ([]Metric, error) {

	m := metrics{}
	err := xml.Unmarshal(xml_data, &m)
	if err != nil {
		log.Error("Error parsing data", err.Error())
	}

	norm_metrics := cleanupMetrics(&m)

	return norm_metrics, nil
}

func cleanupMetrics(m *metrics) []Metric {

	nr, _ := strconv.Atoi(m.NumRows)
	nc, _ := strconv.Atoi(m.NumCols)
	num_metrics := nr * nc

	if num_metrics <= 0 {
		return nil
	}
	norm_metrics := make([]Metric, num_metrics)
	count := 0

	for _, row := range m.Rows {
		ts := row.Timestamp
		for i, val := range row.Values {
			norm_metrics[count].Timestamp = ts
			norm_metrics[count].Key = cleanMetricName(m.Entries[i])
			norm_metrics[count].Value = val
			count += 1
		}

	}

	return norm_metrics
}

//AVERAGE:host:391d9021-49f9-4d97-8d79-1fbd82fd4ffc:memory_total_kib
//to
//host.391d9021-49f9-4d97-8d79-1fbd82fd4ffc.memory_total_kib
func cleanMetricName(name string) string {
	replaced_str := strings.Replace(name, ":", ".", -1)
	return replaced_str[len("AVERAGE:"):]
}
