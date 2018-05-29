package main

import (
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/rodrigodiez/ikigai/pkg/fitbit"
)

type config struct {
	GatewayHost       string        `env:"IKIGAI_PROMETHEUS_PUSHGATEWAY_HOST,required"`
	ClientID          string        `env:"IKIGAI_FITBIT_OAUTH2_CLIENT_ID,required"`
	ClientSecret      string        `env:"IKIGAI_FITBIT_OAUTH2_CLIENT_SECRET,required"`
	AuthorizationCode string        `env:"IKIGAI_FITBIT_OAUTH2_AUTHORIZATION_CODE,required"`
	RedirectURL       string        `env:"IKIGAI_FITBIT_OAUTH2_REDIRECT_URL,required"`
	GatewayPort       int           `env:"IKIGAI_PROMETHEUS_PUSHGATEWAY_PORT" envDefault:"9091"`
	Interval          time.Duration `env:"IKIGAI_INTERVAL_DURATION" envDefault:"60s"`
	Debug             bool          `env:"IKIGAI_DEBUG" envDefault:"false"`
}

func main() {
	cfg := config{}
	err := env.Parse(&cfg)

	if err != nil {
		log.Fatal(err)
	}

	client, err := fitbit.NewClient(cfg.ClientID, cfg.ClientSecret, cfg.RedirectURL, cfg.AuthorizationCode)

	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(cfg.Interval)

	log.Printf("Pushing metrics every %s to http://%s:%d...\n", cfg.Interval, cfg.GatewayHost, cfg.GatewayPort)

	for range ticker.C {
		response, err := client.GetDailyActivitySummary()

		if err != nil {
			log.Printf("GetActivitySummary::ERROR::%s\n", err)
			continue
		}

		if cfg.Debug {
			log.Printf("%+v", response)
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

		totalSteps := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "fitbit_steps_total",
			Help: "Number of steps",
		})
		totalSteps.Set(float64(response.Summary.Steps))

		goalSteps := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "fitbit_steps_goal",
			Help: "Steps goal",
		})
		goalSteps.Set(float64(response.Goals.Steps))

		goalCalories := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "fitbit_calories_goal",
			Help: "Calories goal",
		})
		goalCalories.Set(float64(response.Goals.CaloriesOut))

		registry.MustRegister(totalCalories, activeCalories, bmrCalories, lowActiveMinutes, mediumActiveMinutes, highActiveMinutes, sedentaryMinutes, totalSteps, goalSteps, goalCalories)

		if err := push.FromGatherer("fitbit_api", nil, fmt.Sprintf("http://%s:%d", cfg.GatewayHost, cfg.GatewayPort), registry); err != nil {
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
