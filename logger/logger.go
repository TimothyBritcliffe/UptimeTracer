package logger

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
	"uptimetracer/model"
)

// CreateLog creates a new .CSV file to log the current session
func CreateLog() (*csv.Writer, *os.File, error) {
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

// WriteToLog writes the data to a .csv file
func WriteToLog(site model.Site, timestamp string, alert bool, writer *csv.Writer) error {
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
