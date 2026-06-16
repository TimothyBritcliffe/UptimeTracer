package store

import (
	"encoding/json"
	"fmt"
	"os"
	"uptimetracer/model"
)

// LoadData loads all the sites listed inside of domains.json
func LoadData() ([]model.Site, error) {
	data, err := os.ReadFile("json/domains.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read json/domains.json: %w", err)
	}
	var sites []model.Site
	err = json.Unmarshal(data, &sites)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json/domains.json: %w", err)
	}
	if len(sites) == 0 {
		return nil, fmt.Errorf("json/domains.json contains no sites to monitor")
	}
	for i := range sites {
		// Seed Previous from the saved IsUp in the JSON file preventing immediate alerts on start
		sites[i].Previous = sites[i].IsUp
	}
	return sites, nil
}

// SaveData saves data into the .JSON file for data persistence
func SaveData(sites []model.Site) error {
	data, err := json.MarshalIndent(sites, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("json/domains.json", data, 0644)
}
