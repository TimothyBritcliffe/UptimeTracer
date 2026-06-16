package logger

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
	"uptimetracer/model"
)

type Logger struct {
	file   *os.File
	writer *csv.Writer
}

func NewLogger() (*Logger, error) {
	logger := new(Logger)
	var err error
	logger.writer, logger.file, err = createLog()
	if err != nil {
		return nil, err
	}
	return logger, nil
}

// CreateLog creates a new .CSV file to log the current session
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

// Write adds a new entry to the log file
func (logger *Logger) Write(site model.Site, timestamp string, alert bool) error {
	record := []string{
		timestamp,
		site.Url,
		fmt.Sprintf("%t", alert),
		fmt.Sprintf("%t", site.IsUp),
	}
	err := logger.writer.Write(record)
	if err != nil {
		return fmt.Errorf("failed to write to log: %w", err)
	}

	return nil
}

func (logger *Logger) Flush() {
	logger.writer.Flush()
}

func (logger *Logger) Close() error {
	logger.writer.Flush()
	err := logger.file.Close()
	if err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}
	return nil
}
