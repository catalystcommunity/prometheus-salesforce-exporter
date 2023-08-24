package exporter

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/catalystsquad/app-utils-go/logging"
	sfutils "github.com/catalystsquad/salesforce-utils/pkg"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type SalesforceExporterConfig struct {
	HTTPPort     int
	PollInterval time.Duration

	// Salesforce Config
	SalesforceBaseUrl      string
	SalesforceApiVersion   string
	SalesforceClientID     string
	SalesforceClientSecret string
	SalesforceUsername     string
	SalesforcePassword     string
	SalesforceGrantType    string
}

func Run(config *SalesforceExporterConfig) error {
	// create salesforce client
	sf, err := sfutils.NewSalesforceUtils(true, sfutils.Config{
		BaseUrl:      config.SalesforceBaseUrl,
		ApiVersion:   config.SalesforceApiVersion,
		ClientId:     config.SalesforceClientID,
		ClientSecret: config.SalesforceClientSecret,
		Username:     config.SalesforceUsername,
		Password:     config.SalesforcePassword,
		GrantType:    config.SalesforceGrantType,
	})
	if err != nil {
		return err
	}

	// use a custom registry so we don't get the default metrics
	customRegistry := prometheus.NewRegistry()
	registerMetrics(customRegistry)

	// fetch metrics once to initialize and fail fast if we can't connect
	err = fetchSalesforceMetrics(sf)
	if err != nil {
		return err
	}

	// start a goroutine to fetch metrics on the configured schedule
	go func() {
		ticker := time.NewTicker(config.PollInterval)
		for range ticker.C {
			err := fetchSalesforceMetrics(sf)
			// set health status based on whether we were able to fetch metrics
			healthStatusOK = err == nil
			if err != nil {
				logging.Log.WithError(err).Error("failed to fetch salesforce metrics")
			}
		}
	}()

	// register the health check handler
	http.Handle("/health", http.HandlerFunc(healthCheckHandler))

	// register the metrics handler
	http.Handle("/metrics", promhttp.HandlerFor(customRegistry, promhttp.HandlerOpts{}))

	// start the server
	address := fmt.Sprintf(":%d", config.HTTPPort)
	logging.Log.Info("listening on", address)
	err = http.ListenAndServe(address, nil)
	return err
}

// healthStatusOK is used to indicate whether the exporter is healthy
var healthStatusOK = true

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if healthStatusOK {
		w.WriteHeader(http.StatusOK)
		_,_ = w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		_,_ = w.Write([]byte("NOT OK"))
	}
}

var (
	dailyApiRequestsRemaining = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "salesforce_daily_api_requests_remaining",
		Help: "Salesforce Daily API Requests Remaining",
	})
	dailyApiRequestsLimit = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "salesforce_daily_api_requests_limit",
		Help: "Salesforce Daily API Requests Limit",
	})

	dataStorageMBRemaining = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "salesforce_data_storage_mb_remaining",
		Help: "Salesforce Data Storage MB Remaining",
	})
	dataStorageMBLimit = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "salesforce_data_storage_mb_limit",
		Help: "Salesforce Data Storage MB Limit",
	})
)

func registerMetrics(r *prometheus.Registry) {
	r.MustRegister(dailyApiRequestsRemaining)
	r.MustRegister(dailyApiRequestsLimit)
	r.MustRegister(dataStorageMBRemaining)
	r.MustRegister(dataStorageMBLimit)
}

func fetchSalesforceMetrics(sf *sfutils.SalesforceUtils) error {
	logging.Log.Debug("fetching salesforce metrics")
	resp, err := sf.GetLimits()
	if err != nil {
		if strings.Contains(err.Error(), "INVALID_SESSION_ID") {
			logging.Log.Error("salesforce query failed due to session expiration")
			err := sf.Authenticate()
			if err != nil {
				// panic if we failed, so that the service can restart
				logging.Log.WithError(err).Panic("failed to reauthenticate salesforce utils")
			}
			// try again
			resp, err = sf.GetLimits()
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	dailyApiRequestsRemaining.Set(float64(resp.DailyApiRequests.Remaining))
	dailyApiRequestsLimit.Set(float64(resp.DailyApiRequests.Max))

	dataStorageMBRemaining.Set(float64(resp.DataStorageMB.Remaining))
	dataStorageMBLimit.Set(float64(resp.DataStorageMB.Max))

	return nil
}
