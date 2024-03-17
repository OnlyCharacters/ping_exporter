package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	//_ "github.com/prometheus/client_golang/prometheus"
	//_ "github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	Port       int      `json:"port"`
	MaxLatency int      `json:"max_latency"`
	Tcpings    []Tcping `json:"tcping"`
	Pings      []Ping   `json:"ping"`
}

type Tcping struct {
	Name   string `json:"name"`
	Host   string `json:"host"`
	IsIPv6 bool   `json:"isIPv6"`
	Port   int    `json:"port"`
}

type Ping struct {
	Name string `json:"name"`
	Host string `json:"host"`
}

var max_latency int

func main() {
	config_path := flag.String("c", "./config.json", "path of config.json")
	flag.Parse()

	// parse config file
	var config_json Config
	ParseConfigFile(config_path, &config_json)

	max_latency = config_json.MaxLatency
	port := strconv.Itoa(config_json.Port)

	tcping_gauge_vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tcping_latency",
		Help: "Tcping Latency",
	}, []string{"name", "host", "isv6"})

	reg := prometheus.NewRegistry()
	reg.MustRegister(tcping_gauge_vec)

	InitTcpingMetrics(config_json.Tcpings, *tcping_gauge_vec)

	// update metrics every 3 seconds
	go func() {
		for {
			TcpingMetrics(config_json.Tcpings, *tcping_gauge_vec)
			time.Sleep(3 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	log.Fatalln(http.ListenAndServe(":"+port, nil))
}

func ParseConfigFile(config_path *string, config_json *Config) {
	config_fd, err := os.Open(*config_path)
	if err != nil {
		fmt.Fprint(os.Stderr, "Open config file failed.")
		log.Fatal(err)
	}

	defer config_fd.Close()

	config_bytes, err := io.ReadAll(config_fd)
	if err != nil {
		fmt.Fprint(os.Stderr, "Read config file failed.")
		log.Fatal(err)
	}

	err = json.Unmarshal(config_bytes, config_json)
	if err != nil {
		fmt.Fprint(os.Stderr, "Parse config file failed.")
		log.Fatal(err)
	}
}

func InitTcpingMetrics(tcpings []Tcping, tcping_gauge_vec prometheus.GaugeVec) {
	for _, tcping := range tcpings {
		flag := "false"
		if tcping.IsIPv6 {
			flag = "true"
		} else {
			flag = "false"
		}
		tcping_gauge_vec.WithLabelValues(tcping.Name, tcping.Host, flag)
	}
}

func TcpingMetrics(tcpings []Tcping, tcping_gauge_vec prometheus.GaugeVec) {
	flag := "false"
	timeout := time.Second * 2
	for _, tcping := range tcpings {
		latency := timeout.Milliseconds()
		if tcping.IsIPv6 {
			flag = "true"
			start := time.Now().UnixMilli()
			coon, err := net.DialTimeout("tcp", tcping.Host+":"+strconv.Itoa(tcping.Port), timeout)
			if err != nil {
				fmt.Println("Tcping Error: " + err.Error())
			} else {
				coon.Close()
				end := time.Now().UnixMilli()
				latency = end - start
			}
			//fmt.Println(tcping.Host + " " + strconv.FormatInt(latency, 10) + " ms")
			//tcping_gauge_vec.WithLabelValues(tcping.Name, tcping.Host, flag).Set(float64(latency))
		} else {
			flag = "false"
			start := time.Now().UnixMilli()
			coon, err := net.DialTimeout("tcp4", tcping.Host+":"+strconv.Itoa(tcping.Port), timeout)
			if err != nil {
				fmt.Println("Tcping Error: " + err.Error())
			} else {
				coon.Close()
				end := time.Now().UnixMilli()
				latency = end - start
			}
			//fmt.Println(tcping.Host + " " + strconv.FormatInt(latency, 10) + " ms")
			//tcping_gauge_vec.WithLabelValues(tcping.Name, tcping.Host, flag).Set(float64(latency))
		}
		if latency > int64(max_latency) {
			latency = int64(max_latency)
		}
		fmt.Println(tcping.Host + " " + strconv.FormatInt(latency, 10) + " ms")
		tcping_gauge_vec.WithLabelValues(tcping.Name, tcping.Host, flag).Set(float64(latency))
	}
}
