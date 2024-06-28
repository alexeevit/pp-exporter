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

var inboundApiDownlink = promauto.NewGauge(
        prometheus.GaugeOpts{
                Name: "pp_inbound_api_downlink",
                Help: "",
        },
)

var inboundApiUplink = promauto.NewGauge(
        prometheus.GaugeOpts{
                Name: "pp_inbound_api_uplink",
                Help: "",
        },
)

var inboundMetricsInDownlink = promauto.NewGauge(
        prometheus.GaugeOpts{
                Name: "pp_inbound_metrics_in_downlink",
                Help: "",
        },
)

var inboundMetricsInUplink = promauto.NewGauge(
        prometheus.GaugeOpts{
                Name: "pp_inbound_metrics_in_uplink",
                Help: "",
        },
)

var inboundVlessTlsDownlink = promauto.NewGauge(
        prometheus.GaugeOpts{
                Name: "pp_inbound_vless_tls_downlink",
                Help: "",
        },
)

var inboundVlessTlsUplink = promauto.NewGauge(
        prometheus.GaugeOpts{
                Name: "pp_inbound_vless_tls_uplink",
                Help: "",
        },
)

var outboundBlockDownlink = promauto.NewGauge(
        prometheus.GaugeOpts{
                Name: "pp_outbound_block_downlink",
                Help: "",
        },
)

var outboundBlockUplink = promauto.NewGauge(
        prometheus.GaugeOpts{
                Name: "pp_outbound_block_uplink",
                Help: "",
        },
)

var outboundDirectDownlink = promauto.NewGauge(
        prometheus.GaugeOpts{
                Name: "pp_outbound_direct_downlink",
                Help: "",
        },
)

var outboundDirectUplink = promauto.NewGauge(
        prometheus.GaugeOpts{
                Name: "pp_outbound_direct_uplink",
                Help: "",
        },
)

var userDownlink = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
                Name: "pp_user_downlink",
                Help: "",
        },
        []string{"email"},
)

var userUplink = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
                Name: "pp_user_uplink",
                Help: "",
        },
        []string{"email"},
)

func updateStats(url string, wait int) {
      go func() {
              for {
                        stats, err := loadPPStats(url)

                        if err != nil {
                                log.Println("Error polling PP endpoint, trying again in", wait, "sec")
                        } else {
                                inboundApiDownlink.Set(float64(stats.Inbound.Api.Downlink))
                                inboundApiUplink.Set(float64(stats.Inbound.Api.Uplink))
                                inboundMetricsInDownlink.Set(float64(stats.Inbound.MetricsIn.Downlink))
                                inboundMetricsInUplink.Set(float64(stats.Inbound.MetricsIn.Uplink))
                                inboundVlessTlsDownlink.Set(float64(stats.Inbound.VlessTls.Downlink))
                                inboundVlessTlsUplink.Set(float64(stats.Inbound.VlessTls.Uplink))
                                outboundBlockDownlink.Set(float64(stats.Outbound.Block.Downlink))
                                outboundBlockUplink.Set(float64(stats.Outbound.Block.Uplink))
                                outboundDirectDownlink.Set(float64(stats.Outbound.Direct.Downlink))
                                outboundDirectUplink.Set(float64(stats.Outbound.Direct.Uplink))

                                for email, ioStats := range stats.User {
                                        userDownlink.WithLabelValues(email).Set(float64(ioStats.Downlink))
                                        userUplink.WithLabelValues(email).Set(float64(ioStats.Uplink))
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
