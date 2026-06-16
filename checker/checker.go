package checker

import (
	"fmt"
	"net/http"
	"time"
	"uptimetracer/model"
)

// Defines client with a set timeout
var client = &http.Client{
	Timeout: 10 * time.Second,
}

// CheckStatus changes site.IsUp depending on the status code received from a given site
func CheckStatus(site *model.Site) error {
	resp, err := client.Get(site.Url)
	if err != nil {
		site.IsUp = false // mark as down, not just unknown
		return fmt.Errorf("could not get site %s: %w", site.Url, err)
	}
	defer resp.Body.Close()
	site.IsUp = resp.StatusCode >= 200 && resp.StatusCode < 400

	return nil
}
