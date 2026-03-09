package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// MailDevClient represents a client for interacting with MailDev REST API
type MailDevClient struct {
	baseURL    string
	httpClient *http.Client
}

// Email represents an email from MailDev
type Email struct {
	ID         string    `json:"id"`
	Time       time.Time `json:"time"`
	From       []Address `json:"from"`
	To         []Address `json:"to"`
	Subject    string    `json:"subject"`
	Text       string    `json:"text"`
	HTML       string    `json:"html"`
	Headers    map[string]string `json:"headers"`
	Read       bool      `json:"read"`
	MessageID  string    `json:"messageId"`
	Priority   string    `json:"priority"`
}

// Address represents an email address
type Address struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

// DefaultTokenPattern is the default regex pattern for extracting tokens from emails
// Portal sends short tokens (as short as 6 characters), so we match tokens of any length
const DefaultTokenPattern = `token[=:][\s"\']*([a-zA-Z0-9_-]+)`

// NewMailDevClient creates a new MailDev client
func NewMailDevClient(baseURL string) *MailDevClient {
	if baseURL == "" {
		baseURL = GetMailDevURL()
	}
	return &MailDevClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetAllEmails retrieves all emails from MailDev
func (m *MailDevClient) GetAllEmails() ([]*Email, error) {
	resp, err := m.httpClient.Get(m.baseURL + "/email")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var emails []*Email
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return nil, fmt.Errorf("failed to decode emails: %w", err)
	}

	return emails, nil
}

// GetEmailsByTo retrieves emails sent to a specific address
func (m *MailDevClient) GetEmailsByTo(emailAddress string) ([]*Email, error) {
	emails, err := m.GetAllEmails()
	if err != nil {
		return nil, err
	}

	var filtered []*Email
	for _, email := range emails {
		for _, addr := range email.To {
			if addr.Address == emailAddress {
				filtered = append(filtered, email)
				break
			}
		}
	}

	return filtered, nil
}

// GetEmailsBySubject retrieves emails with a specific subject
func (m *MailDevClient) GetEmailsBySubject(subject string) ([]*Email, error) {
	emails, err := m.GetAllEmails()
	if err != nil {
		return nil, err
	}

	var filtered []*Email
	for _, email := range emails {
		// Trim whitespace and newlines for comparison
		if strings.TrimSpace(email.Subject) == subject {
			filtered = append(filtered, email)
		}
	}

	return filtered, nil
}

// GetEmailBySubjectAndTo retrieves the most recent email matching subject and recipient
func (m *MailDevClient) GetEmailBySubjectAndTo(subject, to string) (*Email, error) {
	emails, err := m.GetEmailsByTo(to)
	if err != nil {
		return nil, err
	}

	// Find the most recent email with matching subject
	var latestEmail *Email
	for _, email := range emails {
		// Trim whitespace and newlines for comparison
		if strings.TrimSpace(email.Subject) == subject {
			if latestEmail == nil || email.Time.After(latestEmail.Time) {
				latestEmail = email
			}
		}
	}

	if latestEmail == nil {
		return nil, fmt.Errorf("no email found with subject '%s' to '%s'", subject, to)
	}

	return latestEmail, nil
}

// ExtractTokenFromEmail extracts a token from email content using regex
// Common token patterns include: token=XXXXXXXX, ?token=XXXXXXXX, /token/XXXXXXXX
func (m *MailDevClient) ExtractTokenFromEmail(email *Email, pattern string) (string, error) {
	// Try text content first
	token := m.extractToken(email.Text, pattern)
	if token != "" {
		return token, nil
	}

	// Try HTML content
	token = m.extractToken(email.HTML, pattern)
	if token != "" {
		return token, nil
	}

	return "", fmt.Errorf("no token found matching pattern '%s'", pattern)
}

// extractToken extracts token using regex pattern
func (m *MailDevClient) extractToken(content, pattern string) string {
	// Try the provided pattern first
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}

	// Try common token patterns if custom pattern fails
	// Note: Portal sends short 6-character tokens, so we match tokens of any length
	patterns := []string{
		DefaultTokenPattern,
		`/token/([a-zA-Z0-9_-]+)`,
		`\?token=([a-zA-Z0-9_-]+)`,
		`reset[/_]token[=:][\s"\']*([a-zA-Z0-9_-]+)`,
		`verify[/_]token[=:][\s"\']*([a-zA-Z0-9_-]+)`,
	}

	for _, p := range patterns {
		re = regexp.MustCompile(p)
		matches = re.FindStringSubmatch(content)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// DeleteAllEmails deletes all emails from MailDev
func (m *MailDevClient) DeleteAllEmails() error {
	req, err := http.NewRequest("DELETE", m.baseURL+"/email/all", nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// WaitForEmail waits for an email to be received within a timeout
func (m *MailDevClient) WaitForEmail(to, subject string, timeout time.Duration) (*Email, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		email, err := m.GetEmailBySubjectAndTo(subject, to)
		if err == nil && email != nil {
			return email, nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil, fmt.Errorf("timeout waiting for email to '%s' with subject '%s'", to, subject)
}

// GetEmailText returns the text content of an email
func (m *MailDevClient) GetEmailText(email *Email) string {
	if email.Text != "" {
		return email.Text
	}
	
	// Strip HTML tags if only HTML is available
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(email.HTML, "")
}
