package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"path/filepath"
	"text/template"
	"time"
)

// EmailService defines the interface for sending emails.
type EmailService interface {
	SendEmail(ctx context.Context, to, subject, templateName string, replacements map[string]interface{}) error
	SendEmailWithAttachment(ctx context.Context, to, subject, templateName string, replacements map[string]interface{}, attachmentFilename string, attachmentContent []byte) error
	SendBulkEmail(ctx context.Context, toEmails []string, subject, templateName string, replacements map[string]interface{}) error
}

// emailService implements EmailService using Go's net/smtp.
type emailService struct {
	smtpHost      string
	smtpPort      int
	smtpUser      string
	smtpPass      string
	fromEmail     string
	templatesPath string
	loginURL      string
}

// NewEmailService creates a new EmailService instance.
func NewEmailService(host string, port int, user, pass, from, templatesPath, loginURL string) EmailService {
	return &emailService{
		smtpHost:      host,
		smtpPort:      port,
		smtpUser:      user,
		smtpPass:      pass,
		fromEmail:     from,
		templatesPath: templatesPath,
		loginURL:      loginURL,
	}
}

func (s *emailService) executeTemplate(templateName string, data map[string]interface{}) (string, error) {
	tmpl, err := template.ParseFiles(filepath.Join(s.templatesPath, templateName))
	if err != nil {
		return "", fmt.Errorf("failed to parse email template %s: %w", templateName, err)
	}

	if data == nil {
		data = make(map[string]interface{})
	}

	data["LoginURL"] = s.loginURL
	data["CurrentYear"] = time.Now().Year()
	var bodyBuffer bytes.Buffer
	if err := tmpl.Execute(&bodyBuffer, data); err != nil {
		log.Printf("Error executing template %s: %v", templateName, err)
		return "", fmt.Errorf("failed to execute email template %s: %w", templateName, err)
	}

	return bodyBuffer.String(), nil
}

func (s *emailService) SendEmail(ctx context.Context, to, subject, templateName string, replacements map[string]interface{}) error {
	body, err := s.executeTemplate(templateName, replacements)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)

	// Prepare message
	msg := []byte("From: " + s.fromEmail + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
		body)

	// Create TLS connection
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // IMPORTANT: Should be false in production with proper certs
		ServerName:         s.smtpHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.smtpHost)
	if err != nil {
		return fmt.Errorf("SMTP client init failed: %w", err)
	}
	defer client.Quit()

	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.smtpHost)

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}

	if err = client.Mail(s.fromEmail); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA write init failed: %w", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		return fmt.Errorf("DATA write failed: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("DATA close failed: %w", err)
	}

	return nil
}

// SendEmailWithAttachment sends an email with an attachment and HTML content from a template.
func (s *emailService) SendEmailWithAttachment(ctx context.Context, to, subject, templateName string, replacements map[string]interface{}, attachmentFilename string, attachmentContent []byte) error {
	htmlBody, err := s.executeTemplate(templateName, replacements)
	if err != nil {
		return err
	}

	// Create a new multipart message
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	boundary := writer.Boundary()

	// Set email headers
	header := make(map[string]string)
	header["From"] = s.fromEmail
	header["To"] = to
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = fmt.Sprintf("multipart/mixed; boundary=%s", boundary)

	// Write email headers
	emailHeaders := ""
	for k, v := range header {
		emailHeaders += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	emailHeaders += "\r\n"

	// Write HTML part
	htmlPartHeader := make(textproto.MIMEHeader)
	htmlPartHeader.Set("Content-Type", "text/html; charset=\"UTF-8\"")
	htmlPartWriter, err := writer.CreatePart(htmlPartHeader)
	if err != nil {
		return fmt.Errorf("failed to create HTML part: %w", err)
	}
	htmlPartWriter.Write([]byte(htmlBody))

	// Add attachment part
	attachmentPartHeader := make(textproto.MIMEHeader)
	attachmentPartHeader.Set("Content-Type", "application/octet-stream")
	attachmentPartHeader.Set("Content-Transfer-Encoding", "base64")
	attachmentPartHeader.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachmentFilename))
	attachmentPartWriter, err := writer.CreatePart(attachmentPartHeader)
	if err != nil {
		return fmt.Errorf("failed to create attachment part: %w", err)
	}
	encodedAttachment := base64.StdEncoding.EncodeToString(attachmentContent)
	attachmentPartWriter.Write([]byte(encodedAttachment))

	writer.Close()

	// Combine headers and multipart body
	fullMessage := []byte(emailHeaders + b.String())

	// SMTP authentication
	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.smtpHost)

	// Send the email
	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)
	err = smtp.SendMail(addr, auth, s.fromEmail, []string{to}, fullMessage)
	if err != nil {
		return fmt.Errorf("failed to send email with attachment to %s: %w", to, err)
	}
	return nil
}

// SendBulkEmail sends the same templated email to multiple recipients.
func (s *emailService) SendBulkEmail(ctx context.Context, toEmails []string, subject, templateName string, replacements map[string]interface{}) error {
	// For simplicity and better deliverability, send individual emails in a loop.
	for _, recipient := range toEmails {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := s.SendEmail(ctx, recipient, subject, templateName, replacements)
			if err != nil {
				log.Printf("Warning: Failed to send bulk email to %s: %v\n", recipient, err)
			}
		}
	}
	return nil
}
