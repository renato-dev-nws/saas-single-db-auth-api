package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds SMTP configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
	AppName  string
	BaseURL  string // e.g., https://mysaas.com
}

// Service handles email sending and template rendering
type Service struct {
	cfg Config
	db  *pgxpool.Pool
}

// NewService creates a new email service
func NewService(cfg Config, db *pgxpool.Pool) *Service {
	return &Service{cfg: cfg, db: db}
}

// BaseURL returns the configured base URL
func (s *Service) BaseURL() string {
	return s.cfg.BaseURL
}

// Template represents an email template from the database
type Template struct {
	ID       string
	Slug     string
	Subject  string
	BodyHTML string
}

// GetTemplate loads a template by slug from the database
func (s *Service) GetTemplate(ctx context.Context, slug string) (*Template, error) {
	var t Template
	err := s.db.QueryRow(ctx,
		`SELECT id, slug, subject, body_html FROM email_templates WHERE slug = $1 AND is_active = true`,
		slug,
	).Scan(&t.ID, &t.Slug, &t.Subject, &t.BodyHTML)
	if err != nil {
		return nil, fmt.Errorf("template '%s' not found: %w", slug, err)
	}
	return &t, nil
}

// RenderTemplate replaces {{key}} placeholders with values
func RenderTemplate(template string, vars map[string]string) string {
	result := template
	for key, value := range vars {
		result = strings.ReplaceAll(result, "{{"+key+"}}", value)
	}
	return result
}

// SendWithTemplate loads a template, renders it, and sends the email
func (s *Service) SendWithTemplate(ctx context.Context, to, templateSlug string, vars map[string]string) error {
	tmpl, err := s.GetTemplate(ctx, templateSlug)
	if err != nil {
		return err
	}

	// Always inject app_name
	if _, ok := vars["app_name"]; !ok {
		vars["app_name"] = s.cfg.AppName
	}

	subject := RenderTemplate(tmpl.Subject, vars)
	body := RenderTemplate(tmpl.BodyHTML, vars)

	return s.Send(to, subject, body)
}

// Send sends an HTML email via SMTP
func (s *Service) Send(to, subject, htmlBody string) error {
	if s.cfg.Host == "" {
		// SMTP not configured — log and skip (useful for dev)
		fmt.Printf("[EMAIL] SMTP not configured. Would send to=%s subject=%s\n", to, subject)
		return nil
	}

	from := s.cfg.From
	addr := s.cfg.Host + ":" + s.cfg.Port

	headers := map[string]string{
		"From":         from,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=UTF-8",
	}

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	// Try TLS (port 465) or STARTTLS (port 587)
	if s.cfg.Port == "465" {
		return s.sendTLS(addr, from, to, msg.String())
	}
	return s.sendSTARTTLS(addr, from, to, msg.String())
}

// sendSTARTTLS sends via STARTTLS (port 587)
func (s *Service) sendSTARTTLS(addr, from, to, msg string) error {
	auth := smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

// sendTLS sends via implicit TLS (port 465)
func (s *Service) sendTLS(addr, from, to, msg string) error {
	tlsConfig := &tls.Config{
		ServerName: s.cfg.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}

	return client.Quit()
}

// SendWelcomeVerification sends the welcome + email verification email
func (s *Service) SendWelcomeVerification(ctx context.Context, to string, vars map[string]string) error {
	return s.SendWithTemplate(ctx, to, "welcome_verify_email", vars)
}

// SendEmailVerified sends the confirmation email after verification
func (s *Service) SendEmailVerified(ctx context.Context, to string, vars map[string]string) error {
	return s.SendWithTemplate(ctx, to, "email_verified", vars)
}
