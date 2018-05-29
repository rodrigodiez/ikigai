package fitbit

type DailyActivitySummaryResponse struct {
	Summary DailyActivitySummarySummary `json:"summary"`
	Goals   DailyActivitySummaryGoals   `json:"goals"`
}

type DailyActivitySummarySummary struct {
	ActivityCalories     int `json:"activityCalories"`
	CaloriesBMR          int `json:"caloriesBMR"`
	CaloriesOut          int `json:"caloriesOut"`
	FairlyActiveMinutes  int `json:"fairlyActiveMinutes"`
	LightlyActiveMinutes int `json:"lightlyActiveMinutes"`
	VeryActiveMinutes    int `json:"veryActiveMinutes"`
	SedentaryMinutes     int `json:"sedentaryMinutes"`
	Steps                int `json:"steps"`
}

type DailyActivitySummaryGoals struct {
	CaloriesOut int `json:"caloriesOut"`
	Steps       int `json:"steps"`
}
