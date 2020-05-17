package main

import (
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
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

var webListenAddress = pflag.String("web.listen-address", ":9001", "The address to listen for requests on")

func main() {
	pflag.Parse()

	configFilename := "constants.yml"

	f, err := os.Open(configFilename)

	if err != nil {
		log.Printf("Unable to open config file %v for reading: %v\n", configFilename, err)
		os.Exit(1)
	}

	defer f.Close()

	d := yaml.NewDecoder(f)

	var cfg config

	err = d.Decode(&cfg)

	if err != nil {
		log.Printf("Unable to decode %v as a YAML file: %v\n", configFilename, err)
		os.Exit(1)
	}

	nameToGauge := make(map[string]*prometheus.GaugeVec)

	for _, metric := range cfg.Metrics {
		labelNames := []string{}
		labelValues := []string{}

		for key, value := range metric.Labels {
			labelNames = append(labelNames, key)
			labelValues = append(labelValues, value)
		}

		keyLabelNames := make([]string, len(labelNames))
		copy(keyLabelNames, labelNames)
		sort.Strings(keyLabelNames)

		lookupKey := metric.Name + "\x1f" + strings.Join(keyLabelNames, "\x1f")

		// XXX because of this, you can only provide help in the first metric of its name
		//     the config should either be called "time_series", or be more hierarchical
		gauge := nameToGauge[lookupKey]

		if gauge == nil {
			gauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
				Name: metric.Name,
				Help: metric.Help,
			}, labelNames)

			nameToGauge[lookupKey] = gauge
		}

		gauge.WithLabelValues(labelValues...).Set(metric.Value)
	}

	promHandler := promhttp.Handler()
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling /metrics request")
		promHandler.ServeHTTP(w, r)
	})

	log.Printf("Listening on %s...\n", *webListenAddress)
	http.ListenAndServe(*webListenAddress, nil)
}
