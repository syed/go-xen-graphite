package xengraphite

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/marpaia/graphite-golang"
	"github.com/nilshell/xmlrpc"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
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
	log.Debug("RPCCall method=%v params=%v\n", method, params)
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

func (c *XenApiClient) GetMetricsUpdate(since time.Time) ([]graphite.Metric, error) {

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

func SendMetricsToGraphite(metrics []graphite.Metric, graphite *graphite.Graphite) error {
	for _, m := range metrics {
		fmt.Printf("[%s]%s : %s\n", m.Timestamp, m.Name, m.Value)
	}
	return graphite.SendMetrics(metrics)
}

func StartClient(config HostConfig, poll_interval int, graphite *graphite.Graphite, wg *sync.WaitGroup) {

	defer wg.Done()
	fmt.Println(config)

	client := NewXenApiClient(
		config.Url,
		config.Username,
		config.Password,
	)

	err := client.Login()

	if err != nil {
		log.Warning("Error logging in:", err.Error())
		return
	}

	log.Info("Logged in: ", config.Url)
	//sleep for some random time before polling
	time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)

	for {

		update_time := time.Now()
		time.Sleep(time.Duration(poll_interval) * time.Second)

		metrics, _ := client.GetMetricsUpdate(update_time)
		SendMetricsToGraphite(metrics, graphite)
	}
}

func Main() {

	var wg sync.WaitGroup

	config_file := FindConfigFile()
	if len(config_file) == 0 {
		log.Errorf("Config file not found. Copy the sample config file to /etc/xengraphite.json")
	}

	conf := ParseConfigFile(config_file)

	// try to connect a graphite server
	Graphite, err := graphite.NewGraphite(conf.GraphiteHost, conf.GraphitePort)

	// if you couldn't connect to graphite, use a nop
	if err != nil {
		log.Warning("Unable to connect to graphite using noop")
		Graphite = graphite.NewGraphiteNop(conf.GraphiteHost, conf.GraphitePort)
	}

	log.Info("Loaded Graphite connection: ", Graphite)

	for _, host_conf := range conf.Hosts {
		wg.Add(1)

		go StartClient(host_conf, conf.PollInterval, Graphite, &wg)

	}

	log.Info("Main watiting for threads")
	wg.Wait()
}
