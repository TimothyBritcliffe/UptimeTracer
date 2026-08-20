# UptimeTracer
UptimeTracer is a self-hosted utility tool that monitors, logs, and reports on status changes. It sends an HTTP(S) request to each domain/IP address you want it to monitor, and based on the response received, it determines if the target is up or down (further clarification on how this is determined is explained in the section below). There are already several planned features for the future, please refer to the Roadmap section of this README to see what I have cooking at the moment.

---

## Features
- Can monitor multiple URLs concurrently
- Detects status changes (up→down, down→up) and alerts via email
- Logs every check into a CSV file in `/logs`
- Persists monitoring state across system restarts via .JSON
- A domain is considered active if an HTTP response is received with a status code in the 200-399 (inclusive) range. Any other HTTP status code received, or any request error (such as timeouts, DNS, connection failure) means the domain is considered inactive.

---

## Setup
1. Clone this repository
2. Add URLs to `json/domains.json`
3. Set environment variables
4. Run `go run .`

---

## Environment Variables
This project uses environment variables to store important data for email alerts such as the sending email address, Google account app password, and the list of recipient emails. The following table shows the variable name, example, and description:

| Variable Name | Example                                 | Description |
|------------|-----------------------------------------|----|
| EMAIL_ADDR | example@gmail.com                       | The email address you are wanting the system to send from |
| EMAIL_PASSWORD | abcd efgh ijkl mnop                     | The Google account application password you generated (not your Google account password) |
| EMAIL_RECIPIENTS | example1@gmail.com,example2@outlook.com | The list of email addresses you want to receive alerts (if more than one, separate them with a comma; no spaces. |
| SMTP_HOST | smtp.gmail.com | The email server you are looking to use |
| SMTP_PORT | 587 | The port you would like to use for the email alert feature |
| INTERVAL_MINUTES | 5 | The length of time (in minutes) between checks |

A .env file must be created to store the environment variables. This project uses [godotenv](https://github.com/joho/godotenv) to manage environment variables. This file should be formatted like so:
```
EMAIL_ADDR=you@gmail.com
EMAIL_PASSWORD=abcd efgh ijkl mnop
EMAIL_RECIPIENTS=you@gmail.com,other@outlook.com
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
INTERVAL_MINUTES=5
```

---

## Roadmap

- **v0.1.0** (Completed): Stable initial release of the CLI application with email alerts, and JSON files for keeping track of monitoring domains
- **v0.2.0** (Planned): SQLite migration & HTML/CSS dashboard (breaking change — JSON files deprecated)
- **v0.3.0**: ICMP monitoring support (non-breaking)
