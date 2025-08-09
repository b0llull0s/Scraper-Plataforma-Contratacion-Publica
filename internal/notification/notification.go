package notification

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"scraper/internal/scraper"
)

// Notifier handles sending notifications
type Notifier struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	toEmails     []string
}

// NewNotifier creates a new notifier instance
func NewNotifier(smtpHost, smtpPort, smtpUsername, smtpPassword, fromEmail string, toEmails []string) *Notifier {
	return &Notifier{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		fromEmail:    fromEmail,
		toEmails:     toEmails,
	}
}

// SendNewContractsNotification sends an email notification about new contracts
func (n *Notifier) SendNewContractsNotification(contracts []scraper.Contract) error {
	if len(contracts) == 0 {
		return nil
	}

	subject := fmt.Sprintf("New LED Screen Contracts Found (%d)", len(contracts))
	body := n.buildEmailBody(contracts)

	return n.sendEmail(subject, body)
}

// sendEmail sends an email using SMTP
func (n *Notifier) sendEmail(subject, body string) error {
	auth := smtp.PlainAuth("", n.smtpUsername, n.smtpPassword, n.smtpHost)

	// Build email headers
	headers := []string{
		fmt.Sprintf("From: %s", n.fromEmail),
		fmt.Sprintf("To: %s", strings.Join(n.toEmails, ", ")),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		body,
	}

	message := strings.Join(headers, "\r\n")

	// Send email
	err := smtp.SendMail(
		n.smtpHost+":"+n.smtpPort,
		auth,
		n.fromEmail,
		n.toEmails,
		[]byte(message),
	)

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email notification sent to %s", strings.Join(n.toEmails, ", "))
	return nil
}

// buildEmailBody creates the HTML email body
func (n *Notifier) buildEmailBody(contracts []scraper.Contract) string {
	var sb strings.Builder

	sb.WriteString(`
	<html>
	<head>
		<style>
			body { font-family: Arial, sans-serif; margin: 20px; }
			.contract { border: 1px solid #ddd; margin: 10px 0; padding: 15px; border-radius: 5px; }
			.contract-id { font-weight: bold; color: #333; }
			.contract-description { margin: 10px 0; }
			.contract-details { color: #666; font-size: 14px; }
			.amount { color: #2c5aa0; font-weight: bold; }
			.status { color: #28a745; font-weight: bold; }
		</style>
	</head>
	<body>
		<h2>New LED Screen Contracts Found</h2>
		<p>We found <strong>`)
	sb.WriteString(fmt.Sprintf("%d", len(contracts)))
	sb.WriteString(`</strong> new contract(s) for LED screens:</p>
	`)

	for _, contract := range contracts {
		sb.WriteString(`
		<div class="contract">
			<div class="contract-id">`)
		sb.WriteString(contract.ID)
		sb.WriteString(`</div>
			<div class="contract-description">`)
		sb.WriteString(contract.Description)
		sb.WriteString(`</div>
			<div class="contract-details">
				<strong>Type:</strong> `)
		sb.WriteString(contract.ContractType)
		sb.WriteString(` | <strong>Status:</strong> <span class="status">`)
		sb.WriteString(contract.Status)
		sb.WriteString(`</span> | <strong>Amount:</strong> <span class="amount">`)
		sb.WriteString(contract.Amount)
		sb.WriteString(`</span><br>
				<strong>Submission Date:</strong> `)
		sb.WriteString(contract.SubmissionDate)
		sb.WriteString(` | <strong>Contracting Body:</strong> `)
		sb.WriteString(contract.ContractingBody)
		sb.WriteString(`
			</div>
		</div>
		`)
	}

	sb.WriteString(`
		<p><small>This notification was sent automatically by the LED Screen Contract Scraper.</small></p>
	</body>
	</html>
	`)

	return sb.String()
}

// TestConnection tests the email configuration
func (n *Notifier) TestConnection() error {
	log.Println("Testing email configuration...")

	// Try to authenticate with SMTP server
	auth := smtp.PlainAuth("", n.smtpUsername, n.smtpPassword, n.smtpHost)

	// Create a test connection
	addr := n.smtpHost + ":" + n.smtpPort
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Authenticate
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate with SMTP server: %w", err)
	}

	log.Println("Email configuration test successful")
	return nil
} 