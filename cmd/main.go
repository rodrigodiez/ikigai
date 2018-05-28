package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/rodrigodiez/ikigai/pkg/fitbit"
)

func main() {
	var (
		gatewayHost       string
		gatewayPort       int
		interval          time.Duration
		clientID          string
		clientSecret      string
		authorizationCode string
		redirectURL       string
		err               error
	)

	gatewayHost = os.Getenv("IKIGAI_PROMETHEUS_PUSHGATEWAY_HOST")
	clientID = os.Getenv("IKIGAI_FITBIT_OAUTH2_CLIENT_ID")
	clientSecret = os.Getenv("IKIGAI_FITBIT_OAUTH2_CLIENT_SECRET")
	authorizationCode = os.Getenv("IKIGAI_FITBIT_OAUTH2_AUTHORIZATION_CODE")
	redirectURL = os.Getenv("IKIGAI_FITBIT_OAUTH2_REDIRECT_URL")

	gatewayPort, err = strconv.Atoi(os.Getenv("IKIGAI_PROMETHEUS_PUSHGATEWAY_PORT"))
	if err != nil {
		log.Println("Can't parse IKIGAI_PROMETHEUS_PUSHGATEWAY_PORT to int: ", err)
		printUsage()
		os.Exit(1)
	}

	interval, err = time.ParseDuration(os.Getenv("IKIGAI_INTERVAL_DURATION"))
	if err != nil {
		log.Println("Can't parse IKIGAI_INTERVAL_DURATION to time.Duration: ", err)
		printUsage()
		os.Exit(1)
	}

	if gatewayHost == "" || clientID == "" || clientSecret == "" || authorizationCode == "" || redirectURL == "" {
		printUsage()
		os.Exit(1)
	}

	client, err := fitbit.NewClient(clientID, clientSecret, redirectURL, authorizationCode)

	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(interval)

	log.Printf("Pushing metrics every %s to http://%s:%d...\n", interval, gatewayHost, gatewayPort)

	for range ticker.C {
		response, err := client.GetDailyActivitySummary()

		if err != nil {
			log.Printf("GetActivitySummary::ERROR::%s\n", err)
			continue
		}

		registry := prometheus.NewRegistry()

		totalCalories := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "fitbit_calories_total",
			Help: "Total number of calories",
		})
		totalCalories.Set(float64(response.Summary.CaloriesOut))

		activeCalories := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "fitbit_calories_active_total",
			Help: "Number of calories above BMR",
		})
		activeCalories.Set(float64(response.Summary.ActivityCalories))

		bmrCalories := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "fitbit_calories_bmr_total",
			Help: "Number of BMR calories",
		})
		bmrCalories.Set(float64(response.Summary.CaloriesBMR))

		lowActiveMinutes := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "fitbit_minutes_low_activity_total",
			Help: "Number of minutes expent in low activity",
		})
		lowActiveMinutes.Set(float64(response.Summary.FairlyActiveMinutes))

		mediumActiveMinutes := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "fitbit_minutes_medium_activity_total",
			Help: "Number of minutes expent in medium activity",
		})
		mediumActiveMinutes.Set(float64(response.Summary.LightlyActiveMinutes))

		highActiveMinutes := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "fitbit_minutes_high_activity_total",
			Help: "Number of minutes expent in high activity",
		})
		highActiveMinutes.Set(float64(response.Summary.VeryActiveMinutes))

		sedentaryMinutes := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "fitbit_minutes_sedentary_total",
			Help: "Number of sedentary minutes expent",
		})
		sedentaryMinutes.Set(float64(response.Summary.SedentaryMinutes))

		registry.MustRegister(totalCalories, activeCalories, bmrCalories, lowActiveMinutes, mediumActiveMinutes, highActiveMinutes, sedentaryMinutes)

		if err := push.FromGatherer("fitbit_api", nil, fmt.Sprintf("http://%s:%d", gatewayHost, gatewayPort), registry); err != nil {
			log.Printf("PushMetrics::ERROR::%s\n", err)
			continue
		}
	}
}

func printUsage() {
	fmt.Println("Please set the following environment variables:")
	fmt.Println("IKIGAI_FITBIT_OAUTH2_CLIENT_ID : Client ID of your Fitbit App")
	fmt.Println("IKIGAI_FITBIT_OAUTH2_CLIENT_SECRET : Client secret of your Fitbit App")
	fmt.Println("IKIGAI_FITBIT_OAUTH2_AUTHORIZATION_CODE : Code to exchange for an access and refresh token")
	fmt.Println("IKIGAI_FITBIT_OAUTH2_REDIRECT_URL : URL where Fitbit would send the auhorization code")
	fmt.Println("IKIGAI_PROMETHEUS_PUSHGATEWAY_HOST : Host where Prometheus PushGateway runs on")
	fmt.Println("IKIGAI_PROMETHEUS_PUSHGATEWAY_PORT : Port where Prometheus PushGateway runs on")
	fmt.Println("IKIGAI_INTERVAL_DURATION : How often to pull metrics from Fitbit (60s, 5m, 1h...)")
}
