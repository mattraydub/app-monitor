package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"sync"
	"time"
)

// Config structures
type ApplicationConfig struct {
	Name         string `json:"name"`
	URL          string `json:"url"`
	Enabled      bool   `json:"enabled"`
	ExpectedCode int    `json:"expected_code"`
}

type EmailConfig struct {
	SMTPHost  string `json:"smtp_host"`
	SMTPPort  string `json:"smtp_port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FromEmail string `json:"from_email"`
	ToEmail   string `json:"to_email"`
}

type Config struct {
	CheckInterval string              `json:"check_interval"`
	Applications  []ApplicationConfig `json:"applications"`
	Email         EmailConfig         `json:"email"`
}

type Monitor struct {
	config       Config
	httpClient   *http.Client
	alertTracker map[string]int // Changed from bool to int to track failure count
	mu           sync.RWMutex
}

func loadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &config, nil
}

func (m *Monitor) sendAlert(application ApplicationConfig, statusCode int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Increment the failure counter
	m.alertTracker[application.Name]++

	// Only send alert if we've failed twice and haven't already alerted
	if m.alertTracker[application.Name] == 2 {
		subject := fmt.Sprintf("ALERT: %s is DOWN", application.Name)
		var body string

		if err != nil {
			body = fmt.Sprintf(`
Service Alert - %s

Application: %s
URL: %s
Status: Connection Failed
Error: %s
Time: %s
Failed Attempts: 2

Please investigate immediately.
`, application.Name, application.Name, application.URL, err.Error(), time.Now().Format("2006-01-02 15:04:05"))
		} else {
			body = fmt.Sprintf(`
Service Alert - %s

Application: %s
URL: %s
Expected Status: %d
Actual Status: %d
Time: %s
Failed Attempts: 2

Please investigate immediately.
`, application.Name, application.Name, application.URL, application.ExpectedCode, statusCode, time.Now().Format("2006-01-02 15:04:05"))
		}

		if err := m.sendEmail(subject, body); err != nil {
			log.Printf("Failed to send email alert for %s: %v", application.Name, err)
		} else {
			log.Printf("Email alert sent for %s after 2 failures", application.Name)
			m.alertTracker[application.Name] = 3 // Use 3 to indicate alert was sent
		}
	}
}

func (m *Monitor) sendRecoveryNotice(application ApplicationConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Only send recovery notice if we previously sent an alert (count >= 3)
	if m.alertTracker[application.Name] < 3 {
		m.alertTracker[application.Name] = 0 // Reset counter on recovery
		return
	}

	subject := fmt.Sprintf("RECOVERY: %s is back online", application.Name)
	body := fmt.Sprintf(`
Service Recovery - %s

Application: %s
URL: %s
Status: OK
Time: %s

Service has recovered and is responding normally.
`, application.Name, application.Name, application.URL, time.Now().Format("2006-01-02 15:04:05"))

	if err := m.sendEmail(subject, body); err != nil {
		log.Printf("Failed to send recovery email for %s: %v", application.Name, err)
	} else {
		log.Printf("Recovery email sent for %s", application.Name)
		m.alertTracker[application.Name] = 0 // Reset counter on recovery
	}
}

func (m *Monitor) sendEmail(subject, body string) error {
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		m.config.Email.FromEmail,
		m.config.Email.ToEmail,
		subject,
		body)

	auth := smtp.PlainAuth("",
		m.config.Email.Username,
		m.config.Email.Password,
		m.config.Email.SMTPHost)

	addr := fmt.Sprintf("%s:%s", m.config.Email.SMTPHost, m.config.Email.SMTPPort)

	return smtp.SendMail(addr, auth, m.config.Email.FromEmail,
		[]string{m.config.Email.ToEmail}, []byte(msg))
}

func (m *Monitor) checkApplication(application ApplicationConfig) {
	if !application.Enabled {
		return
	}

	resp, err := m.httpClient.Get(application.URL)
	if err != nil {
		log.Printf("ERROR: Failed to connect to %s (%s): %v", application.Name, application.URL, err)
		m.sendAlert(application, 0, err)
		return
	}
	defer resp.Body.Close()

	// Read and discard response body to allow connection reuse
	io.Copy(io.Discard, resp.Body)

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	if resp.StatusCode == application.ExpectedCode {
		fmt.Printf("[%s] OK - %s is healthy (Status: %d)\n", timestamp, application.Name, resp.StatusCode)
		m.sendRecoveryNotice(application)
	} else {
		log.Printf("[%s] WARNING - %s returned unexpected status code: %d (expected: %d)",
			timestamp, application.Name, resp.StatusCode, application.ExpectedCode)
		m.sendAlert(application, resp.StatusCode, nil)
	}
}

func (m *Monitor) runChecks() {
	fmt.Printf("Running health checks for %d applications...\n", len(m.config.Applications))

	var wg sync.WaitGroup
	for _, application := range m.config.Applications {
		wg.Add(1)
		go func(ep ApplicationConfig) {
			defer wg.Done()
			m.checkApplication(ep)
		}(application)
	}
	wg.Wait()
}

func main() {
	configFile := "config.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	checkInterval, err := time.ParseDuration(config.CheckInterval)
	if err != nil {
		log.Fatalf("Invalid check interval: %v", err)
	}

	monitor := &Monitor{
		config: *config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		alertTracker: make(map[string]int), // Changed from bool to int
	}

	// Count enabled applications
	enabledCount := 0
	for _, application := range config.Applications {
		if application.Enabled {
			enabledCount++
		}
	}

	fmt.Printf("Starting App monitor with %d enabled applications\n", enabledCount)
	fmt.Printf("Check interval: %v\n", checkInterval)
	fmt.Printf("Alert email: %s\n", config.Email.ToEmail)
	fmt.Println("Press Ctrl+C to stop")

	// Initial check
	monitor.runChecks()

	// Set up ticker for periodic checks
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		monitor.runChecks()
	}
}
