package main

import (
    "encoding/json"
    "flag"
    "net/http"
    "time"
    "log"
    "fmt"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

type DebugVarsRoot struct {
    Stats StatsRoot
}

type StatsRoot struct {
    Inbound InboundStats
    Outbound OutboundStats
    User map[string]IOEntry
}

type InboundStats struct {
    Api IOEntry
    MetricsIn IOEntry `json:"metrics_in"`
    VlessTls IOEntry `json:"vless_tls"`
}

type OutboundStats struct {
    Block IOEntry
    Direct IOEntry
}

type IOEntry struct {
    Downlink int
    Uplink int
}

var inboundApiDownlinkValue float64
var inboundApiDownlink = promauto.NewCounter(
    prometheus.CounterOpts{
        Name: "pp_inbound_api_downlink_total",
        Help: "",
    },
)

var inboundApiUplinkValue float64
var inboundApiUplink = promauto.NewCounter(
    prometheus.CounterOpts{
        Name: "pp_inbound_api_uplink_total",
        Help: "",
    },
)

var inboundMetricsInDownlinkValue float64
var inboundMetricsInDownlink = promauto.NewCounter(
    prometheus.CounterOpts{
        Name: "pp_inbound_metrics_in_downlink_total",
        Help: "",
    },
)

var inboundMetricsInUplinkValue float64
var inboundMetricsInUplink = promauto.NewCounter(
    prometheus.CounterOpts{
        Name: "pp_inbound_metrics_in_uplink_total",
        Help: "",
    },
)

var inboundVlessTlsDownlinkValue float64
var inboundVlessTlsDownlink = promauto.NewCounter(
    prometheus.CounterOpts{
        Name: "pp_inbound_vless_tls_downlink_total",
        Help: "",
    },
)

var inboundVlessTlsUplinkValue float64
var inboundVlessTlsUplink = promauto.NewCounter(
    prometheus.CounterOpts{
        Name: "pp_inbound_vless_tls_uplink_total",
        Help: "",
    },
)

var outboundBlockDownlinkValue float64
var outboundBlockDownlink = promauto.NewCounter(
    prometheus.CounterOpts{
        Name: "pp_outbound_block_downlink_total",
        Help: "",
    },
)

var outboundBlockUplinkValue float64
var outboundBlockUplink = promauto.NewCounter(
    prometheus.CounterOpts{
        Name: "pp_outbound_block_uplink_total",
        Help: "",
    },
)

var outboundDirectDownlinkValue float64
var outboundDirectDownlink = promauto.NewCounter(
    prometheus.CounterOpts{
        Name: "pp_outbound_direct_downlink_total",
        Help: "",
    },
)

var outboundDirectUplinkValue float64
var outboundDirectUplink = promauto.NewCounter(
    prometheus.CounterOpts{
        Name: "pp_outbound_direct_uplink_total",
        Help: "",
    },
)

var userDownlinkValues map[string]float64 = make(map[string]float64)
var userDownlink = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "pp_user_downlink_total",
        Help: "",
    },
    []string{"email"},
)

var userUplinkValues map[string]float64 = make(map[string]float64)
var userUplink = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "pp_user_uplink_total",
        Help: "",
    },
    []string{"email"},
)

func getValueWithDefault(m map[string]float64, key string, defaultValue float64) float64 {
    if value, exists := m[key]; exists {
        return value
    }
    return defaultValue
}

func updateStats(url string, wait int) {
    go func() {
        for {
              stats, err := loadPPStats(url)

              if err != nil {
                  log.Println("Error polling PP endpoint, trying again in", wait, "sec")
              } else {
                  inboundApiDownlink.Add(float64(stats.Inbound.Api.Downlink) - inboundApiDownlinkValue)
                  inboundApiDownlinkValue = float64(stats.Inbound.Api.Downlink)

                  inboundApiUplink.Add(float64(stats.Inbound.Api.Uplink) - inboundApiUplinkValue)
                  inboundApiUplinkValue = float64(stats.Inbound.Api.Uplink)

                  inboundMetricsInDownlink.Add(float64(stats.Inbound.MetricsIn.Downlink) - inboundMetricsInDownlinkValue)
                  inboundMetricsInDownlinkValue = float64(stats.Inbound.MetricsIn.Downlink)

                  inboundMetricsInUplink.Add(float64(stats.Inbound.MetricsIn.Uplink) - inboundMetricsInUplinkValue)
                  inboundMetricsInUplinkValue = float64(stats.Inbound.MetricsIn.Uplink)

                  inboundVlessTlsDownlink.Add(float64(stats.Inbound.VlessTls.Downlink) - inboundVlessTlsDownlinkValue)
                  inboundVlessTlsDownlinkValue = float64(stats.Inbound.VlessTls.Downlink)

                  inboundVlessTlsUplink.Add(float64(stats.Inbound.VlessTls.Uplink) - inboundVlessTlsUplinkValue)
                  inboundVlessTlsUplinkValue = float64(stats.Inbound.VlessTls.Uplink)

                  outboundBlockDownlink.Add(float64(stats.Outbound.Block.Downlink) - outboundBlockDownlinkValue)
                  outboundBlockDownlinkValue = float64(stats.Outbound.Block.Downlink)

                  outboundBlockUplink.Add(float64(stats.Outbound.Block.Uplink) - outboundBlockUplinkValue)
                  outboundBlockUplinkValue = float64(stats.Outbound.Block.Uplink)

                  outboundDirectDownlink.Add(float64(stats.Outbound.Direct.Downlink) - outboundDirectDownlinkValue)
                  outboundDirectDownlinkValue = float64(stats.Outbound.Direct.Downlink)

                  outboundDirectUplink.Add(float64(stats.Outbound.Direct.Uplink) - outboundDirectUplinkValue)
                  outboundDirectUplinkValue = float64(stats.Outbound.Direct.Uplink)

                  for email, ioStats := range stats.User {
                      currentDownlink := getValueWithDefault(userDownlinkValues, email, 0.0)
                      currentUplink := getValueWithDefault(userUplinkValues, email, 0.0)

                      userDownlink.WithLabelValues(email).Add(float64(ioStats.Downlink) - currentDownlink)
                      userDownlinkValues[email] = float64(ioStats.Downlink)

                      userUplink.WithLabelValues(email).Add(float64(ioStats.Uplink) - currentUplink)
                      userUplinkValues[email] = float64(ioStats.Uplink)
                  }
              }
              log.Println("Stats are successfully updated")
              time.Sleep(time.Duration(wait) * time.Second)
        }
    }()
}

func loadPPStats(url string) (stats *StatsRoot, err error) {
    resp, err := http.Get(url)
    if err != nil {
        log.Println("Loading failed")
        return
    }
    defer resp.Body.Close()
    var debugVars = new(DebugVarsRoot)
    json.NewDecoder(resp.Body).Decode(&debugVars)
    return &debugVars.Stats, nil
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello to PP exporter"))
}

func main() {
        urlPtr := flag.String("url", "", "PP status endpoint (normally: http://host/debug/vars)")
        portPtr := flag.Int("port", 2112, "Port to listen on for metrics")
        endpointPtr := flag.String("endpoint", "/metrics", "Metrics endpoint to listen on")
        waitPtr := flag.Int("interval", 15, "Interval to update statistics from PP")

        flag.Parse()

        if *urlPtr == "" {
          log.Fatalf("Missing required argument -url, see '-help' for information")
        }

        log.Printf("Starting PP Exporter on :%v%s\n", *portPtr, *endpointPtr)

        updateStats(*urlPtr, *waitPtr)

        http.Handle(*endpointPtr, promhttp.Handler())
        http.HandleFunc("/", rootHandler)
        log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *portPtr), nil))
}
