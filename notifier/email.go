package notifier

import (
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"uptimetracer/model"
)

// SendEmail uses SMTP to send an email from a designated GMAIL account (must have an app password)
func SendEmail(site model.Site, timestamp string) error {
	//Define variables cleanly
	emailAdd := os.Getenv("EMAIL_ADDR")
	password := os.Getenv("EMAIL_PASSWORD") // Specifically an app password
	emailServer := os.Getenv("SMTP_HOST")
	emailPort := os.Getenv("SMTP_PORT")
	recipients := strings.Split(os.Getenv("EMAIL_RECIPIENTS"), ",")

	a := smtp.PlainAuth("", emailAdd, password, emailServer)
	//Determines if the site went down or up
	var keyword string
	if site.IsUp {
		keyword = "up"
	} else {
		keyword = "down"
	}

	// Removes the prefix https:// or http:// - only used for formatting the email
	display := strings.TrimPrefix(site.Url, "https://")
	display = strings.TrimPrefix(display, "http://")

	//Define From, To, Subject, MIME, Content Type (HTML), and the HTML-based message
	msg := []byte(
		"From: " + emailAdd + "\r\n" +
			"To: " + strings.Join(recipients, ", ") + "\r\n" +
			"Subject: " + "Alert for " + display + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" +
			"<h1>Alert!</h1><p>Hey User,<br>As of " + timestamp + ", " + display + " is currently " + keyword + "</p>\r\n",
	)

	//Send the email to all email addresses in "recipients"
	err := smtp.SendMail(emailServer+":"+emailPort, a, emailAdd, recipients, msg)
	if err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}
	return nil
}

// ValidateEmailConfig is a helper function to validate whether the environment variables have been set
func ValidateEmailConfig() error {
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

// AllowedSMTPPorts is a map containing the three valid SMTP port numbers
var AllowedSMTPPorts = map[int]bool{
	25:  true, // Standard SMTP (many ISPs block outbound)
	465: true, // SMTPS (implicit TLS)
	587: true, // STARTTLS (recommended modern default)
}

// ValidatePort ensures SMTP_PORT is set, set to a number, and set to one of the specific ports in AllowedSMTPPorts
func ValidatePort() error {
	portStr := os.Getenv("SMTP_PORT")

	if portStr == "" {
		return fmt.Errorf("SMTP_PORT cannot be empty")
	}

	portNum, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid SMTP_PORT '%s': must be numeric", portStr)
	}

	if !AllowedSMTPPorts[portNum] {
		return fmt.Errorf(
			"invalid SMTP_PORT '%d'. Supported ports: 25, 465, 587 "+
				"(refer to your email provider documentation)",
			portNum,
		)
	}

	return nil
}
