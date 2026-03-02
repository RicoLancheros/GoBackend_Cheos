package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"crypto/tls"
)

type EmailService struct {
	// SMTP (local)
	smtpHost     string
	smtpPort     string
	smtpEmail    string
	smtpPassword string

	// Resend (produccion)
	resendAPIKey string
	resendFrom   string

	frontendURL string
	useResend   bool
}

func NewEmailService(cfg EmailConfig) *EmailService {
	// Si hay credenciales SMTP, usar SMTP (desarrollo local)
	// Si no, usar Resend (produccion en Render)
	useResend := cfg.SMTPEmail == "" || cfg.SMTPPassword == ""

	if useResend {
		log.Printf("[EMAIL] Modo: RESEND (HTTP API) - from=%s", cfg.ResendFrom)
	} else {
		log.Printf("[EMAIL] Modo: SMTP (Gmail) - email=%s, host=%s:%s", cfg.SMTPEmail, cfg.SMTPHost, cfg.SMTPPort)
	}

	return &EmailService{
		smtpHost:     cfg.SMTPHost,
		smtpPort:     cfg.SMTPPort,
		smtpEmail:    cfg.SMTPEmail,
		smtpPassword: cfg.SMTPPassword,
		resendAPIKey: cfg.ResendAPIKey,
		resendFrom:   cfg.ResendFrom,
		frontendURL:  cfg.FrontendURL,
		useResend:    useResend,
	}
}

// EmailConfig agrupa las configuraciones de email
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPEmail    string
	SMTPPassword string
	ResendAPIKey string
	ResendFrom   string
	FrontendURL  string
}

// SendPasswordResetEmail envia un correo con el enlace para restablecer la contrasena
func (s *EmailService) SendPasswordResetEmail(toEmail, toName, resetToken string) error {
	log.Printf("[EMAIL] Enviando email de reset a: %s (modo: %s)", toEmail, s.modeLabel())

	resetURL := s.frontendURL + "/reset-password?token=" + resetToken

	body := passwordResetTemplate
	body = strings.ReplaceAll(body, "{NOMBRE}", toName)
	body = strings.ReplaceAll(body, "{RESET_URL}", resetURL)

	var err error
	if s.useResend {
		err = s.sendWithResend(toEmail, body)
	} else {
		err = s.sendWithSMTP(toEmail, body)
	}

	if err != nil {
		return err
	}

	log.Printf("[EMAIL] Email enviado exitosamente a: %s", toEmail)
	return nil
}

func (s *EmailService) modeLabel() string {
	if s.useResend {
		return "RESEND"
	}
	return "SMTP"
}

// ==================== RESEND (HTTP) ====================

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (s *EmailService) sendWithResend(toEmail, htmlBody string) error {
	payload := resendRequest{
		From:    fmt.Sprintf("Cheos Cafe <%s>", s.resendFrom),
		To:      []string{toEmail},
		Subject: "Cheos Cafe - Restablecer contraseña",
		HTML:    htmlBody,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error serializando payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.resendAPIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error enviando request a Resend: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		log.Printf("[EMAIL] Resend error %d: %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("resend error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// ==================== SMTP (Gmail) ====================

func (s *EmailService) sendWithSMTP(toEmail, htmlBody string) error {
	message := fmt.Sprintf(
		"From: Cheos Cafe <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: Cheos Cafe - Restablecer contraseña\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		s.smtpEmail, toEmail, htmlBody,
	)

	addr := s.smtpHost + ":" + s.smtpPort

	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("error conectando a %s: %w", addr, err)
	}

	conn.SetDeadline(time.Now().Add(15 * time.Second))

	client, err := smtp.NewClient(conn, s.smtpHost)
	if err != nil {
		conn.Close()
		return fmt.Errorf("error creando cliente SMTP: %w", err)
	}
	defer client.Quit()

	tlsConfig := &tls.Config{ServerName: s.smtpHost}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("error STARTTLS: %w", err)
	}

	auth := smtp.PlainAuth("", s.smtpEmail, s.smtpPassword, s.smtpHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("error autenticando: %w", err)
	}

	if err := client.Mail(s.smtpEmail); err != nil {
		return fmt.Errorf("error MAIL FROM: %w", err)
	}
	if err := client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("error RCPT TO: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("error DATA: %w", err)
	}
	if _, err := w.Write([]byte(message)); err != nil {
		return fmt.Errorf("error escribiendo mensaje: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("error cerrando writer: %w", err)
	}

	return nil
}

// ==================== TEMPLATE ====================

const passwordResetTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; background-color: #f5f0eb; padding: 20px;">
  <div style="max-width: 500px; margin: 0 auto; background: white; border-radius: 12px; padding: 30px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
    <h1 style="color: #6F4E37; text-align: center;">Cheos Cafe</h1>
    <p>Hola <strong>{NOMBRE}</strong>,</p>
    <p>Recibimos una solicitud para restablecer la contraseña de tu cuenta.</p>
    <p>Haz clic en el siguiente botón para crear una nueva contraseña:</p>
    <div style="text-align: center; margin: 25px 0;">
      <a href="{RESET_URL}" style="background-color: #6F4E37; color: white; padding: 12px 30px; text-decoration: none; border-radius: 8px; font-size: 16px;">
        Restablecer contraseña
      </a>
    </div>
    <p style="color: #888; font-size: 13px;">Este enlace expira en 15 minutos.</p>
    <p style="color: #888; font-size: 13px;">Si no solicitaste este cambio, puedes ignorar este correo. Tu contraseña no será modificada.</p>
    <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
    <p style="color: #aaa; font-size: 11px; text-align: center;">Cheos Cafe - Café de especialidad colombiano</p>
  </div>
</body>
</html>`
