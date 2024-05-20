package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	localHeightGauge   prometheus.Gauge
	networkHeightGauge prometheus.Gauge
)

func main() {
	listenPort := flag.String("listen.port", "8380", "port to listen on")
	endpoint := flag.String("endpoint", "http://localhost:26658", "endpoint to connect to")
	p2pNetwork := flag.String("p2p.network", "blockspacerace", "network to use")
	flag.Parse()

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		client := &http.Client{}
		authToken := getAuthToken(*p2pNetwork)

		_, chainId := getHeight(client, authToken, "header.NetworkHead", *endpoint)

		setHeightHandler(chainId)

		for {
			updateMetrics(client, authToken, *endpoint)
			time.Sleep(5 * time.Second)
		}
	}()

	fmt.Printf("Celestia Bridge Exporter started on port %s\n", *listenPort)
	http.ListenAndServe(":"+*listenPort, nil)
}

func updateMetrics(client *http.Client, authToken, endpoint string) {
	local, network, err := getHeights(client, authToken, endpoint)
	if err != nil {
		fmt.Println("Error getting heights:", err)
		return
	}

	localHeightGauge.Set(float64(local))
	networkHeightGauge.Set(float64(network))

}

func getHeights(client *http.Client, authToken, endpoint string) (int, int, error) {
	local, _ := getHeight(client, authToken, "header.LocalHead", endpoint)
	network, _ := getHeight(client, authToken, "header.NetworkHead", endpoint)

	return local, network, nil
}

func getHeight(client *http.Client, authToken, method, endpoint string) (int, string) {
	reqData := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  []interface{}{},
	}
	reqBytes, _ := json.Marshal(reqData)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	resp, err := client.Do(req)
	if err != nil {
		return 0, ""
	}
	defer resp.Body.Close()

	respBytes, _ := ioutil.ReadAll(resp.Body)
	var respData map[string]interface{}
	json.Unmarshal(respBytes, &respData)

	header := respData["result"].(map[string]interface{})["header"]

	heightStr := header.(map[string]interface{})["height"].(string)
	height, err := strconv.Atoi(heightStr)
	if err != nil {
		fmt.Printf("Error converting height to int: %v\n", err)
		return 0, ""
	}

	chainIdStr := header.(map[string]interface{})["chain_id"].(string)

	return height, chainIdStr
}

func getAuthToken(p2pNetwork string) string {
	out, err := exec.Command("celestia", "bridge", "auth", "admin", "--p2p.network", p2pNetwork).Output()
	if err != nil {
		fmt.Println("Error getting auth token:", err)
		return ""
	}

	return strings.TrimSpace(string(out))
}

func setHeightHandler(chainId string) {
	ConstLabels := map[string]string{
		"chain_id": chainId,
	}

	localHeightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "bridge_local_height",
		Help:        "Local height of the Celestia node",
		ConstLabels: ConstLabels,
	})

	networkHeightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "bridge_network_height",
		Help:        "Network height of the Celestia node",
		ConstLabels: ConstLabels,
	})

	prometheus.MustRegister(localHeightGauge)
	prometheus.MustRegister(networkHeightGauge)
}
