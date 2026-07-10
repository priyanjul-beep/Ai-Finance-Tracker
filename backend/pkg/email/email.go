// Package email provides a transactional email service backed by SMTP.
package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/priyanjul/ai-finance-tracker/internal/dto"
)

// Service sends transactional emails over SMTP.
type Service struct {
	host     string
	port     string
	username string
	password string
	from     string
}

// New creates a new SMTP email service.
func New(host, port, username, password, from string) *Service {
	return &Service{host: host, port: port, username: username, password: password, from: from}
}

// SendVerification sends an email-verification link.
func (s *Service) SendVerification(_ context.Context, to, link string) error {
	subject := "Verify your AI Finance Tracker email"
	body := fmt.Sprintf(`
Hi there,

Please verify your email address by clicking the link below:

%s

This link expires in 24 hours.

– AI Finance Tracker Team
`, link)
	return s.send(to, subject, body)
}

// SendPasswordReset sends a password-reset link.
func (s *Service) SendPasswordReset(_ context.Context, to, link string) error {
	subject := "Reset your AI Finance Tracker password"
	body := fmt.Sprintf(`
Hi there,

You requested a password reset. Click the link below to set a new password:

%s

This link expires in 1 hour. If you didn't request this, please ignore this email.

– AI Finance Tracker Team
`, link)
	return s.send(to, subject, body)
}

// SendWelcome sends a welcome email after successful signup.
func (s *Service) SendWelcome(_ context.Context, to, name string) error {
	subject := "Welcome to AI Finance Tracker 🎉"
	body := fmt.Sprintf(`
Hi %s,

Welcome to AI Finance Tracker! You're all set to track your finances intelligently.

Get started by logging your first expense. Your AI assistant is ready to help!

– AI Finance Tracker Team
`, name)
	return s.send(to, subject, body)
}

// SendBudgetAlert sends a budget overrun warning.
func (s *Service) SendBudgetAlert(_ context.Context, to, message string) error {
	return s.send(to, "Budget Alert – AI Finance Tracker", message)
}

// SendWeeklySummary sends the weekly financial summary.
func (s *Service) SendWeeklySummary(_ context.Context, to, summary string) error {
	return s.send(to, "Your Weekly Financial Summary", summary)
}

// SendMonthlyReport sends a monthly financial report email.
func (s *Service) SendMonthlyReport(_ context.Context, to string, report *dto.MonthlyReportDTO) error {
	body := fmt.Sprintf(`
Hi there,

Here's your financial summary for %d/%d:

  Income:   ₹%.0f
  Expenses: ₹%.0f
  Savings:  ₹%.0f (%.1f%%)

Log in to AI Finance Tracker to see the full breakdown.

– AI Finance Tracker Team
`, report.Month, report.Year, report.TotalIncome, report.TotalExpense, report.TotalSavings, report.SavingsRate)
	return s.send(to, fmt.Sprintf("Your Monthly Report – %d/%d", report.Month, report.Year), body)
}

// send is the low-level SMTP dispatcher.
func (s *Service) send(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	msg := buildMessage(s.from, to, subject, body)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("email: send to %s: %w", to, err)
	}
	return nil
}

func buildMessage(from, to, subject, body string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("From: %s\r\n", from))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", to))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(body)
	return sb.String()
}
