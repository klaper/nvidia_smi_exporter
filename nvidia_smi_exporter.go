package main

import (
    "bytes"
    "encoding/csv"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/exec"
    "strings"

    "github.com/kardianos/service"
)

func (p *program) metrics(response http.ResponseWriter, request *http.Request) {
    out, err := exec.Command(
        "C:\\Program Files\\NVIDIA Corporation\\NVSMI\\nvidia-smi",
        "--query-gpu=name,index,temperature.gpu,utilization.gpu,utilization.memory,memory.total,memory.free,memory.used,encoder.stats.averageFps,clocks_throttle_reasons.gpu_idle,clocks_throttle_reasons.applications_clocks_setting,clocks_throttle_reasons.sw_power_cap,clocks_throttle_reasons.hw_slowdown,clocks_throttle_reasons.hw_thermal_slowdown,clocks_throttle_reasons.hw_power_brake_slowdown,clocks_throttle_reasons.sync_boost,fan.speed,driver_version,power.draw,clocks.current.graphics,clocks.current.sm,clocks.current.memory,clocks.current.video",
        "--format=csv,noheader,nounits").Output()

    if err != nil {
        fmt.Fprintf(response, err.Error())
        return
    }

    csvReader := csv.NewReader(bytes.NewReader(out))
    csvReader.TrimLeadingSpace = true
    records, err := csvReader.ReadAll()

    if err != nil {
        fmt.Fprintf(response, err.Error())
        return
    }

    metricList := []string{
        "temperature.gpu", "utilization.gpu",
        "utilization.memory", "memory.total", "memory.free", "memory.used", "encoder.stats.averageFps", "clocks_throttle_reasons.gpu_idle",
        "clocks_throttle_reasons.applications_clocks_setting", "clocks_throttle_reasons.sw_power_cap", "clocks_throttle_reasons.hw_slowdown", "clocks_throttle_reasons.hw_thermal_slowdown",
        "clocks_throttle_reasons.hw_power_brake_slowdown", "clocks_throttle_reasons.sync_boost", "fan.speed", "driver_version", "power.draw", "clocks.current.graphics", "clocks.current.sm", "clocks.current.memory",
        "clocks.current.video"}

    result := ""
    for _, row := range records {
        for idx, value := range row[2:] {
            metric := value
            if metric == "Active" {
                metric = "1"
            } else if metric == "Not Active" {
                metric = "0"
            }
            result = fmt.Sprintf("%s%s%s{gpu=\"%s\", number=\"%s\", vendor=\"nvidia\"} %s\n", result, "node_gpu_", strings.Replace(metricList[idx], ".", "_", -1), row[0], row[1], metric)
        }
    }
    fmt.Fprintf(response, result)
}

var logger service.Logger

type program struct{}

func (p *program) Start(s service.Service) error {
    // Start should not block. Do the actual work async.
    go p.run()
    return nil
}
func (p *program) run() {
    // Do work here
    addr := ":9101"
    if len(os.Args) > 1 {
        addr = os.Args[1]
    }

    http.HandleFunc("/metrics/", p.metrics)
    err := http.ListenAndServe(addr, nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
func (p *program) Stop(s service.Service) error {
    // Stop should not block. Return with a few seconds.
    return nil
}

func main() {
    svcConfig := &service.Config{
        Name:        "prometheus-nvidia-smi-exporter",
        DisplayName: "PrometheusNvidiaSMIExporter",
        Description: "Prometheus Nvidia SMI exporter",
    }

    prg := &program{}
    s, err := service.New(prg, svcConfig)
    if err != nil {
        log.Fatal(err)
    }
    logger, err = s.Logger(nil)
    if err != nil {
        log.Fatal(err)
    }
    err = s.Run()
    if err != nil {
        logger.Error(err)
    }
}
