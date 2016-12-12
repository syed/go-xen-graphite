package xengraphite

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	_ "github.com/marpaia/graphite-golang"
	"github.com/nilshell/xmlrpc"
	"gopkg.in/alexcesaro/statsd.v2"
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

func (c *XenApiClient) GetMetricsUpdate(since time.Time) ([]StatsdMetric, error) {

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

	fmt.Println(metrics)

	if len(metrics) == 0 {
		// We might have lost the session,
		// refresh it by relogin
		log.Info("Got no metrics, trying to login again")
		c.Login()
	}

	return metrics, nil
}

func SendMetricsToGraphite(metrics []StatsdMetric, statsd *statsd.Client) {

	for _, m := range metrics {
		f, _ := strconv.ParseFloat(m.Value, 64)
		fmt.Printf("%s : %f\n", m.Name, f)
		statsd.Timing(m.Name, f)
	}
}

func XenLoginWithRetry(xen_client *XenApiClient, retry_interval int) {

	log.Info("Trying to login: ", xen_client.Url)

	for {
		err := xen_client.Login()

		if err == nil {
			log.Info("Logged in: ", xen_client.Url)
			return
		}
		log.Warning("Error logging in:", err.Error())
		time.Sleep(time.Duration(retry_interval) * time.Second)
	}

}

func StartClient(config HostConfig, poll_interval int, retry_interval int, statsd *statsd.Client, wg *sync.WaitGroup) {

	defer wg.Done()
	fmt.Println(config)

	xen_client := NewXenApiClient(
		config.Url,
		config.Username,
		config.Password,
	)

	XenLoginWithRetry(xen_client, retry_interval)

	//sleep for some random time before polling
	time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)

	for {

		update_time := time.Now()
		time.Sleep(time.Duration(poll_interval) * time.Second)

		metrics, err := xen_client.GetMetricsUpdate(update_time)
		if err != nil {
			log.Warning("Lost connection to XenServer, retrying to login", err.Error())
			XenLoginWithRetry(xen_client, retry_interval)
		}

		SendMetricsToGraphite(metrics, statsd)
	}
}

func Main() {

	var wg sync.WaitGroup

	config_file := FindConfigFile()
	if len(config_file) == 0 {
		log.Errorf("Config file not found. Copy the sample config file to /etc/xenstatsd.json")
	}

	conf := ParseConfigFile(config_file)

	addr := fmt.Sprintf("%s:%d", conf.StatsdHost, conf.StatsdPort)
	fmt.Printf("STATSD ADDRESS: %s\n", addr)

	// try to connect a graphite server
	//Graphite, err := graphite.NewGraphite(conf.GraphiteHost, conf.GraphitePort)
	statsd_client, err := statsd.New(statsd.Address(addr))

	// if you couldn't connect to graphite, use a nop
	if err != nil {
		log.Warning("Unable to connect to statsd", err)
	}

	log.Info("Loaded Graphite connection: ", statsd_client)

	for _, host_conf := range conf.Hosts {
		wg.Add(1)

		go StartClient(host_conf, conf.PollInterval, conf.RetryInterval, statsd_client, &wg)

	}

	log.Info("Main watiting for threads")
	wg.Wait()
}
