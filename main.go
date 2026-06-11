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

// Loads all the Urls listed inside of domains.json
func loadData() ([]model.Url, error) {
	data, err := os.ReadFile("json/domains.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read json/domains.json: %w", err)
	}
	var urls []model.Url
	err = json.Unmarshal(data, &urls)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json/domains.json: %w", err)
	}
	if len(urls) == 0 {
		return nil, fmt.Errorf("json/domains.json contains no URLs to monitor")
	}
	// Ensures the state persists (won't re-alert on a restart)
	for i := range urls {
		urls[i].Previous = urls[i].IsUp
	}
	return urls, nil
}

// Defines client with a set timeout
var client = &http.Client{
	Timeout: 10 * time.Second,
}

// Checker to change url.IsUp depending on the status code received from a given url
func checkStatus(url *model.Url, wg *sync.WaitGroup) {
	defer wg.Done()
	//updates the previous variable, gets the status of the webpage and updates the UpDown variable of the given url
	resp, err := client.Get(url.Url)
	if err != nil {
		url.IsUp = false // mark as down, not just unknown
		return
	}
	defer resp.Body.Close()
	url.IsUp = resp.StatusCode >= 200 && resp.StatusCode < 400
}

// Saves data into the .JSON file for data persistence
func saveData(urls []model.Url) error {
	data, err := json.MarshalIndent(urls, "", "  ")
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
// isUp is the most up-to-date value of model.Url.UpDown
// Creates the CSV
func createLog() string {
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal("Failed to create logs directory:", err)
	}

	filename := "logs/data-" + time.Now().Format("2006-01-02_15-04-05") + ".csv"
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write([]string{"timestamp", "url", "alert", "isUp"})
	if err != nil {
		return ""
	}

	return filename
}

// Writes the data to a .csv file
func writeToLog(url model.Url, filename string, timestamp string, alert bool) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open log file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{
		timestamp,
		url.Url,
		fmt.Sprintf("%t", alert),
		fmt.Sprintf("%t", url.IsUp),
	}
	if err := writer.Write(record); err != nil {
		log.Println("Failed to write to log:", err)
	}
}

// Uses SMTP to send an email from a designated GMAIL account (must have an app password)
func sendEmail(url model.Url, timestamp string) {
	//Define variables cleanly
	emailAdd := os.Getenv("EMAIL_ADDR")
	password := os.Getenv("EMAIL_PASSWORD") // Specifically an app password
	emailServer := os.Getenv("SMTP_HOST")
	recipients := strings.Split(os.Getenv("EMAIL_RECIPIENTS"), ",")

	a := smtp.PlainAuth("", emailAdd, password, emailServer)
	//Determines if the url went down or up
	var keyword string
	if url.IsUp {
		keyword = "up"
	} else {
		keyword = "down"
	}

	//Define From, To, Subject, MIME, Content Type (HTML), and the HTML-based message
	msg := []byte(
		"From: " + emailAdd + "\r\n" +
			"To: " + strings.Join(recipients, ", ") + "\r\n" +
			"Subject: " + "Alert for " + strings.TrimPrefix(url.Url, "https://") + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" +
			"<h1>Alert!</h1><p>Hey User,<br>As of " + timestamp + ", " + url.Url + " is currently " + keyword + "</p>\r\n",
	)

	//Send the email to all email addresses in "recipients"
	err := smtp.SendMail(emailServer+":587", a, emailAdd, recipients, msg)
	if err != nil {
		log.Println("Error sending email:", err)
	}
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
	data, err := loadData()
	if err != nil {
		log.Fatal(err)
	}
	envCheck := true
	err = validateEmailConfig()
	if err != nil {
		fmt.Println(err)
		envCheck = false
	}
	logName := createLog()
	// Main loop that iterates forever
	for {
		// Loop to iterate through entire set of domains
		wg := sync.WaitGroup{}
		for i := range data {
			wg.Add(1)
			go checkStatus(&data[i], &wg)
		}
		wg.Wait()
		// Loop to determine if the status has changed; will only print an alert if it has.
		for _, myUrl := range data {
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			// Logic to log an Alert (Also sends an email)
			if myUrl.IsUp != myUrl.Previous {
				status := "UP"
				if !myUrl.IsUp {
					status = "DOWN"
				}
				msg := fmt.Sprintf("ALERT: %s is %s", myUrl.Url, status)
				fmt.Println(msg)
				writeToLog(myUrl, logName, timestamp, true)
				if envCheck {
					sendEmail(myUrl, timestamp)
				}
			} else {
				// Logic to log when there is no changes (no emails)
				fmt.Println(myUrl.String())
				writeToLog(myUrl, logName, timestamp, false)
			}
		}
		// Update url.Previous for next cycle
		for i := range data {
			data[i].Previous = data[i].IsUp
		}
		// Saves updated URLS to the .JSON
		err := saveData(data)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(5 * time.Minute) // Modify this depending on the frequency you want to check the domains
	}
}
