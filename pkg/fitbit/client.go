package fitbit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/fitbit"
)

type Client interface {
	GetDailyActivitySummary() (DailyActivitySummaryResponse, error)
}

func NewClient(clientId string, clientSecret string, redirectURL string, authorizationCode string) (Client, error) {
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"activity", "heartrate", "location", "nutrition", "profile", "settings", "sleep", "social", "weight"},
		Endpoint:     fitbit.Endpoint,
	}

	tok, err := conf.Exchange(ctx, authorizationCode)

	if err != nil {
		return nil, err
	}

	return &client{httpc: conf.Client(ctx, tok)}, nil
}

type client struct {
	httpc *http.Client
}

func (c *client) GetDailyActivitySummary() (DailyActivitySummaryResponse, error) {
	var response = DailyActivitySummaryResponse{}

	url := fmt.Sprintf(
		"https://api.fitbit.com/1/user/-/activities/date/%s.json",
		time.Now().Format("2006-01-02"))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return response, err
	}

	resp, err := c.httpc.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return response, errors.New(resp.Status)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}
