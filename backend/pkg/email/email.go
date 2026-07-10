// Package email provides a transactional email service backed by SMTP.
// All emails are sent as responsive HTML with plain-text fallback.
package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"time"

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

// ─── Public API ──────────────────────────────────────────────────────────────

// SendVerification sends an email-verification link.
func (s *Service) SendVerification(_ context.Context, to, link string) error {
	return s.sendHTML(to, "Verify your AI Finance Tracker email", s.renderTemplate("verify", verificationTemplate, map[string]interface{}{"Link": link}))
}

// SendPasswordReset sends a password-reset link.
func (s *Service) SendPasswordReset(_ context.Context, to, link string) error {
	return s.sendHTML(to, "Reset your AI Finance Tracker password", s.renderTemplate("reset", passwordResetTemplate, map[string]interface{}{"Link": link}))
}

// SendWelcome sends a welcome email (delegates to HTML version).
func (s *Service) SendWelcome(_ context.Context, to, name string) error {
	return s.SendWelcomeHTML(context.Background(), to, name)
}

// SendWelcomeHTML sends a fully-templated responsive HTML welcome email.
func (s *Service) SendWelcomeHTML(_ context.Context, to, name string) error {
	return s.sendHTML(to, "Welcome to AI Finance Tracker 🎉", s.renderTemplate("welcome", welcomeTemplate, map[string]interface{}{"Name": name}))
}

// SendBudgetAlert sends a simple text budget alert (legacy).
func (s *Service) SendBudgetAlert(_ context.Context, to, message string) error {
	return s.sendHTML(to, "Budget Alert – AI Finance Tracker", wrapSimple(message))
}

// SendWeeklySummary sends the weekly financial summary.
func (s *Service) SendWeeklySummary(_ context.Context, to, summary string) error {
	return s.sendHTML(to, "Your Weekly Financial Summary", wrapSimple(summary))
}

// SendMonthlyReport sends a monthly report email.
func (s *Service) SendMonthlyReport(_ context.Context, to string, _ *dto.MonthlyReportDTO) error {
	return s.sendHTML(to, "Your Monthly Financial Report", wrapSimple("Your monthly financial report is ready. Log in to view your full report."))
}

// SendBudgetWarningHTML sends a rich budget warning email (90% / custom threshold).
func (s *Service) SendBudgetWarningHTML(_ context.Context, to, name, category string,
	budgetAmount, spent, remaining, threshold float64,
	month, year, daysLeft int,
) error {
	subject := fmt.Sprintf("⚠️ Budget Usage Alert – %.0f%% Reached", threshold)
	data := budgetWarningData{
		Name:         name,
		Category:     strings.Title(strings.ReplaceAll(category, "_", " ")),
		BudgetAmount: budgetAmount,
		Spent:        spent,
		Remaining:    remaining,
		Threshold:    threshold,
		Percent:      spent / budgetAmount * 100,
		Month:        time.Month(month).String(),
		Year:         year,
		DaysLeft:     daysLeft,
	}
	return s.sendHTML(to, subject, s.renderTemplate("budget_warning", budgetWarningTemplate, data))
}

// SendBudgetExceededHTML sends a rich budget exceeded email.
func (s *Service) SendBudgetExceededHTML(_ context.Context, to, name, category string,
	budgetAmount, spent, overspent float64,
) error {
	data := budgetExceededData{
		Name:         name,
		Category:     strings.Title(strings.ReplaceAll(category, "_", " ")),
		BudgetAmount: budgetAmount,
		Spent:        spent,
		Overspent:    overspent,
	}
	return s.sendHTML(to, "🚨 Budget Exceeded – AI Finance Tracker", s.renderTemplate("budget_exceeded", budgetExceededTemplate, data))
}

// ─── Template data structs ───────────────────────────────────────────────────

type budgetWarningData struct {
	Name, Category, Month string
	BudgetAmount, Spent, Remaining, Threshold, Percent float64
	Year, DaysLeft                                     int
}

type budgetExceededData struct {
	Name, Category           string
	BudgetAmount, Spent, Overspent float64
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func (s *Service) renderTemplate(name, tmplStr string, data interface{}) string {
	tmpl := template.Must(template.New(name).Parse(tmplStr))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "<p>Error rendering email template</p>"
	}
	return buf.String()
}

func (s *Service) sendHTML(to, subject, htmlBody string) error {
	if s.host == "" {
		fmt.Printf("[email] SMTP not configured — skipping email to=%s subject=%s\n", to, subject)
		return nil
	}
	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("From: %s\r\n", s.from))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", to))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(htmlBody)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(sb.String())); err != nil {
		return fmt.Errorf("email: send to %s: %w", to, err)
	}
	return nil
}

func wrapSimple(msg string) string {
	return fmt.Sprintf(`<!DOCTYPE html><html><body style="font-family:sans-serif;color:#374151;padding:32px;max-width:600px;margin:auto;"><p>%s</p><p style="color:#9ca3af;font-size:12px;margin-top:32px;">AI Finance Tracker</p></body></html>`, msg)
}

// ─── HTML Email Templates ────────────────────────────────────────────────────

const welcomeTemplate = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Welcome</title></head>
<body style="margin:0;padding:0;background:#f3f4f6;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="background:#f3f4f6;padding:32px 16px;">
<tr><td align="center"><table width="600" cellpadding="0" cellspacing="0" style="background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 4px 6px rgba(0,0,0,.07);">
<tr><td style="background:linear-gradient(135deg,#4f46e5,#7c3aed);padding:40px 32px;text-align:center;">
<h1 style="margin:0;color:#fff;font-size:28px;font-weight:700;">🎉 Welcome aboard!</h1>
<p style="margin:8px 0 0;color:#c4b5fd;font-size:16px;">AI Finance Tracker</p></td></tr>
<tr><td style="padding:32px;">
<p style="color:#374151;font-size:16px;line-height:1.6;margin:0 0 16px;">Hi <strong>{{.Name}}</strong>,</p>
<p style="color:#374151;font-size:15px;line-height:1.6;margin:0 0 24px;">Your account has been created successfully. We're excited to help you take control of your finances with AI-powered insights. 🚀</p>
<h2 style="color:#111827;font-size:18px;font-weight:600;margin:0 0 16px;">Get started in 3 easy steps</h2>
<table width="100%" cellpadding="0" cellspacing="8"><tr><td>
<div style="background:#f5f3ff;border-left:4px solid #4f46e5;border-radius:8px;padding:16px 20px;margin-bottom:12px;">
<p style="margin:0;color:#4f46e5;font-size:12px;font-weight:600;text-transform:uppercase;">Step 1</p>
<p style="margin:4px 0 0;color:#111827;font-size:15px;font-weight:600;">💸 Add your first expense</p>
<p style="margin:4px 0 0;color:#6b7280;font-size:13px;">Track every spend by category, merchant, or tag.</p></div>
<div style="background:#eff6ff;border-left:4px solid #2563eb;border-radius:8px;padding:16px 20px;margin-bottom:12px;">
<p style="margin:0;color:#2563eb;font-size:12px;font-weight:600;text-transform:uppercase;">Step 2</p>
<p style="margin:4px 0 0;color:#111827;font-size:15px;font-weight:600;">🎯 Create your first budget</p>
<p style="margin:4px 0 0;color:#6b7280;font-size:13px;">Set spending limits and get alerts before you overspend.</p></div>
<div style="background:#f0fdf4;border-left:4px solid #16a34a;border-radius:8px;padding:16px 20px;margin-bottom:24px;">
<p style="margin:0;color:#16a34a;font-size:12px;font-weight:600;text-transform:uppercase;">Step 3</p>
<p style="margin:4px 0 0;color:#111827;font-size:15px;font-weight:600;">🤖 Explore AI features</p>
<p style="margin:4px 0 0;color:#6b7280;font-size:13px;">Scan receipts, use voice input, and get AI-powered financial insights.</p></div>
</td></tr></table>
<p style="color:#6b7280;font-size:13px;margin:0;"><strong style="color:#374151;">– AI Finance Tracker Team</strong></p>
</td></tr>
<tr><td style="background:#f9fafb;padding:20px 32px;text-align:center;border-top:1px solid #e5e7eb;">
<p style="margin:0;color:#9ca3af;font-size:12px;">AI Finance Tracker · Manage your finances intelligently</p></td></tr>
</table></td></tr></table></body></html>`

const verificationTemplate = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background:#f3f4f6;font-family:sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="padding:32px 16px;"><tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="background:#fff;border-radius:12px;overflow:hidden;">
<tr><td style="background:linear-gradient(135deg,#4f46e5,#7c3aed);padding:32px;text-align:center;">
<h1 style="margin:0;color:#fff;font-size:24px;font-weight:700;">Verify your email ✉️</h1></td></tr>
<tr><td style="padding:32px;">
<p style="color:#374151;font-size:15px;line-height:1.6;margin:0 0 24px;">Please click the button below to verify your email address. This link expires in <strong>24 hours</strong>.</p>
<table cellpadding="0" cellspacing="0" style="margin:0 auto 24px;"><tr>
<td style="background:#4f46e5;border-radius:8px;padding:14px 32px;text-align:center;">
<a href="{{.Link}}" style="color:#fff;font-size:15px;font-weight:600;text-decoration:none;">Verify Email Address</a></td></tr></table>
<p style="color:#9ca3af;font-size:12px;margin:0;">If you didn't create an account, you can safely ignore this email.</p>
</td></tr><tr><td style="background:#f9fafb;padding:20px;text-align:center;border-top:1px solid #e5e7eb;">
<p style="margin:0;color:#9ca3af;font-size:12px;">AI Finance Tracker</p></td></tr>
</table></td></tr></table></body></html>`

const passwordResetTemplate = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background:#f3f4f6;font-family:sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="padding:32px 16px;"><tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="background:#fff;border-radius:12px;overflow:hidden;">
<tr><td style="background:linear-gradient(135deg,#dc2626,#991b1b);padding:32px;text-align:center;">
<h1 style="margin:0;color:#fff;font-size:24px;font-weight:700;">Reset Password 🔒</h1></td></tr>
<tr><td style="padding:32px;">
<p style="color:#374151;font-size:15px;line-height:1.6;margin:0 0 24px;">Click the button below to set a new password. This link expires in <strong>1 hour</strong>.</p>
<table cellpadding="0" cellspacing="0" style="margin:0 auto 24px;"><tr>
<td style="background:#dc2626;border-radius:8px;padding:14px 32px;">
<a href="{{.Link}}" style="color:#fff;font-size:15px;font-weight:600;text-decoration:none;">Reset My Password</a></td></tr></table>
<p style="color:#9ca3af;font-size:12px;margin:0;">If you didn't request this, please ignore this email.</p>
</td></tr><tr><td style="background:#f9fafb;padding:20px;text-align:center;border-top:1px solid #e5e7eb;">
<p style="margin:0;color:#9ca3af;font-size:12px;">AI Finance Tracker</p></td></tr>
</table></td></tr></table></body></html>`

const budgetWarningTemplate = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background:#f3f4f6;font-family:sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="padding:32px 16px;"><tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="background:#fff;border-radius:12px;overflow:hidden;">
<tr><td style="background:linear-gradient(135deg,#d97706,#b45309);padding:32px;text-align:center;">
<h1 style="margin:0;color:#fff;font-size:24px;font-weight:700;">⚠️ Budget Usage Alert</h1>
<p style="margin:8px 0 0;color:#fde68a;font-size:14px;">{{.Month}} {{.Year}} · {{.Category}} Budget</p></td></tr>
<tr><td style="padding:32px;">
<p style="color:#374151;font-size:15px;line-height:1.6;margin:0 0 24px;">
Hi <strong>{{.Name}}</strong>, you've used <strong>{{printf "%.1f" .Percent}}%</strong> of your <strong>{{.Category}}</strong> budget this month.</p>
<table width="100%" cellpadding="0" cellspacing="0" style="background:#fffbeb;border:1px solid #fcd34d;border-radius:8px;margin-bottom:24px;">
<tr>
<td style="padding:16px 20px;border-right:1px solid #fcd34d;text-align:center;">
<p style="margin:0;color:#92400e;font-size:11px;font-weight:600;text-transform:uppercase;">Budget</p>
<p style="margin:4px 0 0;color:#111827;font-size:20px;font-weight:700;">₹{{printf "%.0f" .BudgetAmount}}</p></td>
<td style="padding:16px 20px;border-right:1px solid #fcd34d;text-align:center;">
<p style="margin:0;color:#92400e;font-size:11px;font-weight:600;text-transform:uppercase;">Spent</p>
<p style="margin:4px 0 0;color:#d97706;font-size:20px;font-weight:700;">₹{{printf "%.0f" .Spent}}</p></td>
<td style="padding:16px 20px;text-align:center;">
<p style="margin:0;color:#92400e;font-size:11px;font-weight:600;text-transform:uppercase;">Remaining</p>
<p style="margin:4px 0 0;color:#16a34a;font-size:20px;font-weight:700;">₹{{printf "%.0f" .Remaining}}</p></td>
</tr></table>
{{if gt .DaysLeft 0}}<p style="color:#6b7280;font-size:13px;margin:0 0 24px;">⏱ <strong>{{.DaysLeft}} days remaining</strong> in this budget period.</p>{{end}}
<p style="color:#6b7280;font-size:13px;"><strong style="color:#374151;">– AI Finance Tracker Team</strong></p>
</td></tr><tr><td style="background:#f9fafb;padding:20px;text-align:center;border-top:1px solid #e5e7eb;">
<p style="margin:0;color:#9ca3af;font-size:12px;">AI Finance Tracker · Manage your finances intelligently</p></td></tr>
</table></td></tr></table></body></html>`

const budgetExceededTemplate = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background:#f3f4f6;font-family:sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="padding:32px 16px;"><tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="background:#fff;border-radius:12px;overflow:hidden;">
<tr><td style="background:linear-gradient(135deg,#dc2626,#991b1b);padding:32px;text-align:center;">
<h1 style="margin:0;color:#fff;font-size:24px;font-weight:700;">🚨 Budget Exceeded</h1>
<p style="margin:8px 0 0;color:#fca5a5;font-size:14px;">{{.Category}} Budget Overrun</p></td></tr>
<tr><td style="padding:32px;">
<p style="color:#374151;font-size:15px;line-height:1.6;margin:0 0 24px;">
Hi <strong>{{.Name}}</strong>, your <strong>{{.Category}}</strong> budget has been exceeded.</p>
<table width="100%" cellpadding="0" cellspacing="0" style="background:#fef2f2;border:1px solid #fca5a5;border-radius:8px;margin-bottom:24px;">
<tr>
<td style="padding:16px 20px;border-right:1px solid #fca5a5;text-align:center;">
<p style="margin:0;color:#991b1b;font-size:11px;font-weight:600;text-transform:uppercase;">Budget</p>
<p style="margin:4px 0 0;color:#111827;font-size:20px;font-weight:700;">₹{{printf "%.0f" .BudgetAmount}}</p></td>
<td style="padding:16px 20px;border-right:1px solid #fca5a5;text-align:center;">
<p style="margin:0;color:#991b1b;font-size:11px;font-weight:600;text-transform:uppercase;">Spent</p>
<p style="margin:4px 0 0;color:#dc2626;font-size:20px;font-weight:700;">₹{{printf "%.0f" .Spent}}</p></td>
<td style="padding:16px 20px;text-align:center;">
<p style="margin:0;color:#991b1b;font-size:11px;font-weight:600;text-transform:uppercase;">Overspent</p>
<p style="margin:4px 0 0;color:#dc2626;font-size:20px;font-weight:700;">+₹{{printf "%.0f" .Overspent}}</p></td>
</tr></table>
<div style="background:#f0fdf4;border-left:4px solid #16a34a;border-radius:8px;padding:16px 20px;margin-bottom:24px;">
<p style="margin:0;color:#15803d;font-size:13px;font-weight:600;">💡 Suggested actions</p>
<ul style="margin:8px 0 0;padding-left:20px;color:#374151;font-size:13px;line-height:2;">
<li>Review recent {{.Category}} expenses in the app</li>
<li>Adjust your budget limit for next month</li>
<li>Set a lower custom alert threshold (e.g. 70%)</li>
<li>Use AI insights to identify where you can cut back</li></ul></div>
<p style="color:#6b7280;font-size:13px;"><strong style="color:#374151;">– AI Finance Tracker Team</strong></p>
</td></tr><tr><td style="background:#f9fafb;padding:20px;text-align:center;border-top:1px solid #e5e7eb;">
<p style="margin:0;color:#9ca3af;font-size:12px;">AI Finance Tracker · Manage your finances intelligently</p></td></tr>
</table></td></tr></table></body></html>`

