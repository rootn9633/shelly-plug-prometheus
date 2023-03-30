package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"
)

const (
	shellyAddress = "192.168.2.189" //TODO: read from env
	shellyPath    = "/rpc"
	shellyMethod  = "Switch.GetStatus?id=0"

	serverPort = "9100"
)

const metricTemplate = `
# HELP power Last measured instantaneous active power (in Watts) delivered to the attached load
# TYPE power gauge
power {{ .APower }}

# HELP output true if the output channel is currently on, false otherwise
# TYPE outputi gauge
output {{ .Output }}

# HELP total Total energy consumed in Watt-hours
# TYPE total gauge
total {{ .AEnergy.Total }}
`

type SwitchStatus struct {
	Id          int          `json:"id"`
	Source      string       `json:"source"`
	Output      bool         `json:"output"`
	APower      float32      `json:"apower"`
	Voltage     float32      `json:"voltage"`
	Current     float32      `json:"current"`
	AEnergy     *AEnergy     `json:"aenergy"`
	Temperature *Temperature `json:"temperature"`
}

type AEnergy struct {
	Total    float64   `json:"total"`
	ByMinute []float64 `json:"by_minute"`
	MinuteTS int       `json:"minute_ts"`
}

type Temperature struct {
	TC float32 `json:"tC"`
	TF float32 `json:"tF"`
}

func metrics(w http.ResponseWriter, req *http.Request) {
	requestURL := fmt.Sprintf("http://%s%s/%s", shellyAddress, shellyPath, shellyMethod)
	res, err := http.Get(requestURL)
	if err != nil {
		log.Fatal("error making http request: %w\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	defer res.Body.Close()

	var status SwitchStatus
	if err := json.NewDecoder(res.Body).Decode(&status); err != nil {
		log.Printf("Error unmarshalling body: %v", err)
	}

	tmpl, err := template.New("metrics").Parse(metricTemplate)
	if err != nil {
		log.Fatal("Error executing template: %w\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	tmpl.Execute(w, status)
}

func main() {
	http.HandleFunc("/metrics", metrics)

	http.ListenAndServe(":"+serverPort, nil)
}
