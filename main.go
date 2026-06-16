package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"

	"uptimetracer/model"

	"github.com/joho/godotenv"
)

// Loads all the sites listed inside of domains.json
func loadData() ([]model.Site, error) {
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
	// Ensures the state persists (won't re-alert on a restart)
	for i := range sites {
		sites[i].Previous = sites[i].IsUp
	}
	return sites, nil
}

// Defines client with a set timeout
var client = &http.Client{
	Timeout: 10 * time.Second,
}

// Checker to change site.IsUp depending on the status code received from a given site
func checkStatus(site *model.Site) error {
	//updates the previous variable, gets the status of the webpage and updates the UpDown variable of the given site
	resp, err := client.Get(site.Url)
	if err != nil {
		site.IsUp = false // mark as down, not just unknown
		return fmt.Errorf("could not get site %s: %w", site.Url, err)
	}
	defer resp.Body.Close()
	site.IsUp = resp.StatusCode >= 200 && resp.StatusCode < 400

	return nil
}

// Saves data into the .JSON file for data persistence
func saveData(sites []model.Site) error {
	data, err := json.MarshalIndent(sites, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("json/domains.json", data, 0644)
}

// CSV Format

// timestamp, url, alert, isUp

// timestamp is the time of the check
// url is the specific url
// alert is a boolean, if there is a change it is true, else false
// isUp is the most up-to-date value of model.Site.UpDown
// Creates the CSV
func createLog() (*csv.Writer, *os.File, error) {
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, nil, fmt.Errorf("could not create log directory: %w", err)
	}

	filename := "logs/data-" + time.Now().Format("2006-01-02_15-04-05") + ".csv"
	file, err := os.Create(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create log file: %w", err)
	}
	writer := csv.NewWriter(file)

	err = writer.Write([]string{"timestamp", "url", "alert", "isUp"})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write to file: %w", err)
	}

	return writer, file, nil
}

// Writes the data to a .csv file
func writeToLog(site model.Site, timestamp string, alert bool, writer *csv.Writer) error {
	record := []string{
		timestamp,
		site.Url,
		fmt.Sprintf("%t", alert),
		fmt.Sprintf("%t", site.IsUp),
	}
	err := writer.Write(record)
	if err != nil {
		return fmt.Errorf("failed to write to log: %w", err)
	}

	return nil
}

// Uses SMTP to send an email from a designated GMAIL account (must have an app password)
func sendEmail(site model.Site, timestamp string) error {
	//Define variables cleanly
	emailAdd := os.Getenv("EMAIL_ADDR")
	password := os.Getenv("EMAIL_PASSWORD") // Specifically an app password
	emailServer := os.Getenv("SMTP_HOST")
	recipients := strings.Split(os.Getenv("EMAIL_RECIPIENTS"), ",")

	a := smtp.PlainAuth("", emailAdd, password, emailServer)
	//Determines if the site went down or up
	var keyword string
	if site.IsUp {
		keyword = "up"
	} else {
		keyword = "down"
	}

	//Define From, To, Subject, MIME, Content Type (HTML), and the HTML-based message
	msg := []byte(
		"From: " + emailAdd + "\r\n" +
			"To: " + strings.Join(recipients, ", ") + "\r\n" +
			"Subject: " + "Alert for " + strings.TrimPrefix(site.Url, "https://") + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" +
			"<h1>Alert!</h1><p>Hey User,<br>As of " + timestamp + ", " + site.Url + " is currently " + keyword + "</p>\r\n",
	)

	//Send the email to all email addresses in "recipients"
	err := smtp.SendMail(emailServer+":587", a, emailAdd, recipients, msg)
	if err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}
	return nil
}

// Helper function to validate whether the environment variables have been set
func validateEmailConfig() error {
	var missing []string
	for _, v := range []string{"EMAIL_ADDR", "EMAIL_PASSWORD", "SMTP_HOST", "EMAIL_RECIPIENTS"} {
		if os.Getenv(v) == "" {
			missing = append(missing, v)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("email alerts disabled, missing: %s", strings.Join(missing, ", "))
	}
	return nil
}

func main() {
	_ = godotenv.Load()
	var data []model.Site
	var err error
	if data, err = loadData(); err != nil {
		log.Fatal(err)
	}
	envCheck := true
	if err := validateEmailConfig(); err != nil {
		log.Println(err)
		envCheck = false
	}
	var logWriter *csv.Writer
	var logFile *os.File
	if logWriter, logFile, err = createLog(); err != nil {
		log.Fatal("Not logging:", err)
	}
	defer logFile.Close()
	defer logWriter.Flush()
	// Main loop that iterates forever
	for {
		// Loop to iterate through entire set of domains
		wg := sync.WaitGroup{}
		errs := make([]error, len(data))
		for i := range data {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				errs[i] = checkStatus(&data[i])
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
				if err := writeToLog(site, timestamp, true, logWriter); err != nil {
					log.Println(err)
				}
				if envCheck {
					if err := sendEmail(site, timestamp); err != nil {
						log.Println(err)
					}
				}
			} else {
				// Logic to log when there is no changes (no emails)
				fmt.Println(site.String())
				if err := writeToLog(site, timestamp, false, logWriter); err != nil {
					log.Println(err)
				}
			}
		}
		// Update site.Previous for next cycle
		for i := range data {
			data[i].Previous = data[i].IsUp
		}
		// Saves updated sites to the .JSON
		if err := saveData(data); err != nil {
			log.Println(err)
		}
		logWriter.Flush()
		time.Sleep(5 * time.Minute) // Modify this depending on the frequency you want to check the domains
	}
}
