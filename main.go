package main

import (
	"uptimetracer/checker"
	"uptimetracer/logger"
	"uptimetracer/model"
	"uptimetracer/notifier"
	"uptimetracer/store"

	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	var data []model.Site
	var err error
	if data, err = store.LoadData(); err != nil {
		log.Fatal(err)
	}
	envCheck := true
	if err := notifier.ValidateEmailConfig(); err != nil {
		log.Println(err)
		envCheck = false
	}
	logging, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer logging.Close()

	// Retrieves the set interval in .env

	interval := os.Getenv("INTERVAL_MINUTES")
	var intInterval int64

	if intInterval, err = strconv.ParseInt(interval, 10, 64); err != nil {
		log.Println("no integer assigned to \"INTERVAL_MINUTES\" in .env:", err)
		intInterval = 5
	}

	if intInterval <= 0 {
		log.Println("INTERVAL must be > 0, defaulting to 5 minutes")
		intInterval = 5
	}

	// Main loop that iterates forever
	for {
		// Loop to iterate through entire set of domains
		wg := sync.WaitGroup{}
		errs := make([]error, len(data))
		for i := range data {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				errs[i] = checker.CheckStatus(&data[i])
			}(i)
		}
		wg.Wait()
		// Loop to iterate through the list of errors
		for _, err := range errs {
			if err != nil {
				log.Println(err)
			}
		}
		// Loop to determine if the status has changed; will only print an alert if it has.
		for _, site := range data {
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			// Logic to log an Alert (Also sends an email)
			if site.IsUp != site.Previous {
				status := "UP"
				if !site.IsUp {
					status = "DOWN"
				}
				msg := fmt.Sprintf("ALERT: %s is %s", site.Url, status)
				fmt.Println(msg)
				if err := logging.Write(site, timestamp, true); err != nil {
					log.Println(err)
				}
				if envCheck {
					if err := notifier.SendEmail(site, timestamp); err != nil {
						log.Println(err)
					}
				}
			} else {
				// Logic to log when there is no changes (no emails)
				fmt.Println(site.String())
				if err := logging.Write(site, timestamp, false); err != nil {
					log.Println(err)
				}
			}
		}
		// Update site.Previous for next cycle in-place (not copying values)
		for i := range data {
			data[i].Previous = data[i].IsUp
		}
		// Saves updated sites to the .JSON
		if err := store.SaveData(data); err != nil {
			log.Println(err)
		}
		logging.Flush()

		// Waits before moving on with the next iteration
		time.Sleep(time.Duration(intInterval) * time.Minute)
	}
}
