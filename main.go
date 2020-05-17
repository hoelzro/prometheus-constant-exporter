package main

import (
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	yaml "gopkg.in/yaml.v2"
)

type metric struct {
	Name   string
	Help   string
	Value  float64
	Labels map[string]string
}

type config struct {
	Metrics []metric
}

func main() {
	configFilename := "constants.yml"

	f, err := os.Open(configFilename)

	if err != nil {
		log.Printf("Unable to open config file %v for reading: %v\n", configFilename, err)
		return
	}

	defer f.Close()

	d := yaml.NewDecoder(f)

	var cfg config

	err = d.Decode(&cfg)

	if err != nil {
		log.Printf("Unable to decode %v as a YAML file: %v\n", configFilename, err)
	}

	for _, metric := range cfg.Metrics {
		labelNames := []string{}
		labelValues := []string{}

		for key, value := range metric.Labels {
			labelNames = append(labelNames, key)
			labelValues = append(labelValues, value)
		}

		gauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: metric.Name,
			Help: metric.Help,
		}, labelNames)

		gauge.WithLabelValues(labelValues...).Set(metric.Value)
	}

	http.ListenAndServe(":9001", promhttp.Handler())
}
