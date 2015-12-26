package xengraphite

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/nilshell/xmlrpc"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// NewClient stand up a new XapiClient. Version should probably be "1.2" unless you know what you are doing.
func NewXenApiClient(uri, username, password string) *XenApiClient {
	client := new(XenApiClient)
	client.Url = uri
	client.Username = username
	client.Password = password
	client.Rpc, _ = xmlrpc.NewClient(client.Url, nil)
	return client
}

func (c *XenApiClient) RpcCall(result interface{}, method string, params []interface{}) (err error) {
	log.Debugf("RPCCall method=%v params=%v\n", method, params)
	p := new(xmlrpc.Params)
	p.Params = params
	return c.Rpc.Call(method, *p, result)
}

func (client *XenApiClient) Login() (err error) {

	//Do loging call
	result := xmlrpc.Struct{}
	params := make([]interface{}, 2)
	params[0] = client.Username
	params[1] = client.Password
	err = client.RpcCall(&result, "session.login_with_password", params)
	if err == nil {
		// err might not be set properly, so check the reference
		if result["Value"] == nil {
			return errors.New("Invalid credentials supplied")
		}
	}
	client.Session = result["Value"]
	return err
}

func (c *XenApiClient) GetMetricsUpdate(since time.Time) ([]Metric, error) {

	last_update := strconv.FormatInt(since.Unix(), 10)
	req_url := fmt.Sprintf("%s/rrd_updates?session_id=%s&start=%s&host=true", c.Url, c.Session, last_update)

	log.Info("request host metrics ", req_url)
	resp, err := http.Get(req_url)
	if err != nil {
		return nil, errors.New("Unable to get metrics")
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		return nil, errors.New("Unable to read metrics")
	}

	//fmt.Println(string(data))
	metrics, err := ParseRrdMetrics(data)

	if err != nil {
		return nil, errors.New("Unable to parse metrics")
	}

	return metrics, nil
}

func SendMetricsToGraphite(metrics []Metric) error {
	return nil
}

func Main() {

	client := NewXenApiClient(
		"http://172.31.0.46",
		"root",
		"1q2w3e4r",
	)

	err := client.Login()

	if err != nil {
		log.Fatalf("Error logging in", err.Error())
	}

	log.Info("Logged in")
	fmt.Println(client.Session)
	ten_seconds_before := time.Now().Add(-10 * time.Second)
	metrics, _ := client.GetMetricsUpdate(ten_seconds_before)
	for _, m := range metrics {
		fmt.Printf("[%s]%s : %s\n", m.Timestamp, m.Key, m.Value)
	}
	SendMetricsToGraphite(metrics)

}
