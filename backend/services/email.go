package services

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"civicweave/backend/config"
)

// EmailService handles email operations using Mailgun
type EmailService struct {
	config *config.MailgunConfig
	client *http.Client
}

// NewEmailService creates a new email service
func NewEmailService(cfg *config.MailgunConfig) *EmailService {
	return &EmailService{
		config: cfg,
		client: &http.Client{},
	}
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	Subject string
	HTML    string
	Text    string
}

// SendEmail sends an email using Mailgun
func (s *EmailService) SendEmail(to, subject, html, text string) error {
	if s.config.APIKey == "" || s.config.Domain == "" {
		return fmt.Errorf("mailgun configuration missing")
	}

	// Prepare form data
	data := url.Values{}
	data.Set("from", fmt.Sprintf("CivicWeave <noreply@%s>", s.config.Domain))
	data.Set("to", to)
	data.Set("subject", subject)

	if html != "" {
		data.Set("html", html)
	}

	if text != "" {
		data.Set("text", text)
	}

	// Create request
	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", s.config.Domain), strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("api", s.config.APIKey)

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mailgun returned status %d", resp.StatusCode)
	}

	return nil
}

// SendVerificationEmail sends an email verification email
func (s *EmailService) SendVerificationEmail(to, token string) error {
	verificationURL := fmt.Sprintf("http://localhost:3000/verify-email?token=%s", token)

	subject := "Verify your CivicWeave account"
	html := fmt.Sprintf(`
		<html>
		<body>
			<h2>Welcome to CivicWeave!</h2>
			<p>Thank you for registering. Please click the link below to verify your email address:</p>
			<p><a href="%s">Verify Email Address</a></p>
			<p>If the link doesn't work, copy and paste this URL into your browser:</p>
			<p>%s</p>
			<p>This link will expire in 24 hours.</p>
			<p>Best regards,<br>The CivicWeave Team</p>
		</body>
		</html>
	`, verificationURL, verificationURL)

	text := fmt.Sprintf(`
		Welcome to CivicWeave!
		
		Thank you for registering. Please visit the following link to verify your email address:
		
		%s
		
		This link will expire in 24 hours.
		
		Best regards,
		The CivicWeave Team
	`, verificationURL)

	return s.SendEmail(to, subject, html, text)
}

// SendWelcomeEmail sends a welcome email after verification
func (s *EmailService) SendWelcomeEmail(to, name string) error {
	subject := "Welcome to CivicWeave!"
	html := fmt.Sprintf(`
		<html>
		<body>
			<h2>Welcome to CivicWeave, %s!</h2>
			<p>Your email has been verified and your account is now active.</p>
			<p>You can now:</p>
			<ul>
				<li>Browse available volunteer opportunities</li>
				<li>Apply to initiatives that match your skills</li>
				<li>Track your volunteer applications</li>
			</ul>
			<p><a href="http://localhost:3000/login">Log in to your account</a></p>
			<p>Thank you for joining our community of civic-minded volunteers!</p>
			<p>Best regards,<br>The CivicWeave Team</p>
		</body>
		</html>
	`, name)

	text := fmt.Sprintf(`
		Welcome to CivicWeave, %s!
		
		Your email has been verified and your account is now active.
		
		You can now:
		- Browse available volunteer opportunities
		- Apply to initiatives that match your skills
		- Track your volunteer applications
		
		Log in to your account: http://localhost:3000/login
		
		Thank you for joining our community of civic-minded volunteers!
		
		Best regards,
		The CivicWeave Team
	`, name)

	return s.SendEmail(to, subject, html, text)
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(to, token string) error {
	resetURL := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", token)

	subject := "Reset your CivicWeave password"
	html := fmt.Sprintf(`
		<html>
		<body>
			<h2>Password Reset Request</h2>
			<p>You requested to reset your password. Click the link below to create a new password:</p>
			<p><a href="%s">Reset Password</a></p>
			<p>If the link doesn't work, copy and paste this URL into your browser:</p>
			<p>%s</p>
			<p>This link will expire in 1 hour.</p>
			<p>If you didn't request this reset, please ignore this email.</p>
			<p>Best regards,<br>The CivicWeave Team</p>
		</body>
		</html>
	`, resetURL, resetURL)

	text := fmt.Sprintf(`
		Password Reset Request
		
		You requested to reset your password. Visit the following link to create a new password:
		
		%s
		
		This link will expire in 1 hour.
		
		If you didn't request this reset, please ignore this email.
		
		Best regards,
		The CivicWeave Team
	`, resetURL)

	return s.SendEmail(to, subject, html, text)
}

// SendApplicationConfirmationEmail sends confirmation when volunteer applies
func (s *EmailService) SendApplicationConfirmationEmail(to, volunteerName, initiativeTitle string) error {
	subject := "Application submitted - " + initiativeTitle
	html := fmt.Sprintf(`
		<html>
		<body>
			<h2>Application Submitted Successfully!</h2>
			<p>Hi %s,</p>
			<p>Thank you for your interest in volunteering for <strong>%s</strong>.</p>
			<p>Your application has been submitted and is currently under review. We'll notify you once a decision has been made.</p>
			<p>You can check the status of your application at any time by logging into your account.</p>
			<p><a href="http://localhost:3000/login">View your applications</a></p>
			<p>Thank you for your commitment to civic engagement!</p>
			<p>Best regards,<br>The CivicWeave Team</p>
		</body>
		</html>
	`, volunteerName, initiativeTitle)

	text := fmt.Sprintf(`
		Application Submitted Successfully!
		
		Hi %s,
		
		Thank you for your interest in volunteering for %s.
		
		Your application has been submitted and is currently under review. We'll notify you once a decision has been made.
		
		You can check the status of your application at any time by logging into your account.
		
		View your applications: http://localhost:3000/login
		
		Thank you for your commitment to civic engagement!
		
		Best regards,
		The CivicWeave Team
	`, volunteerName, initiativeTitle)

	return s.SendEmail(to, subject, html, text)
}

// SendApplicationStatusUpdateEmail sends notification when application status changes
func (s *EmailService) SendApplicationStatusUpdateEmail(to, volunteerName, initiativeTitle, status string) error {
	var subject, message string

	switch status {
	case "accepted":
		subject = "Application accepted - " + initiativeTitle
		message = "Great news! Your application for <strong>" + initiativeTitle + "</strong> has been accepted!"
	case "rejected":
		subject = "Application update - " + initiativeTitle
		message = "Thank you for your interest in <strong>" + initiativeTitle + "</strong>. Unfortunately, we're unable to proceed with your application at this time."
	default:
		return fmt.Errorf("unknown status: %s", status)
	}

	html := fmt.Sprintf(`
		<html>
		<body>
			<h2>%s</h2>
			<p>Hi %s,</p>
			<p>%s</p>
			<p><a href="http://localhost:3000/login">View your applications</a></p>
			<p>Best regards,<br>The CivicWeave Team</p>
		</body>
		</html>
	`, subject, volunteerName, message)

	text := fmt.Sprintf(`
		%s
		
		Hi %s,
		
		%s
		
		View your applications: http://localhost:3000/login
		
		Best regards,
		The CivicWeave Team
	`, subject, volunteerName, message)

	return s.SendEmail(to, subject, html, text)
}
