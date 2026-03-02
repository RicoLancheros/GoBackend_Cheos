package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type EmailService struct {
	host        string
	port        string
	email       string
	password    string
	frontendURL string
}

func NewEmailService(host, port, email, password, frontendURL string) *EmailService {
	log.Printf("[EMAIL] Servicio inicializado: host=%s, port=%s, email=%s, frontendURL=%s", host, port, email, frontendURL)
	return &EmailService{
		host:        host,
		port:        port,
		email:       email,
		password:    password,
		frontendURL: frontendURL,
	}
}

// SendPasswordResetEmail envia un correo con el enlace para restablecer la contrasena
func (s *EmailService) SendPasswordResetEmail(toEmail, toName, resetToken string) error {
	log.Printf("[EMAIL] Intentando enviar email de reset a: %s", toEmail)

	resetURL := s.frontendURL + "/reset-password?token=" + resetToken

	body := passwordResetTemplate
	body = strings.ReplaceAll(body, "{NOMBRE}", toName)
	body = strings.ReplaceAll(body, "{RESET_URL}", resetURL)

	message := fmt.Sprintf(
		"From: Cheos Cafe <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: Cheos Cafe - Restablecer contrasena\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		s.email, toEmail, body,
	)

	// Intentar TLS directo (puerto 465)
	log.Printf("[EMAIL] Intentando TLS directo (puerto 465)...")
	err := s.sendWithTLS(toEmail, message)
	if err != nil {
		log.Printf("[EMAIL] TLS directo FALLO: %v", err)

		// Intentar STARTTLS (puerto 587)
		log.Printf("[EMAIL] Intentando STARTTLS (puerto %s)...", s.port)
		err = s.sendWithSTARTTLS(toEmail, message)
		if err != nil {
			log.Printf("[EMAIL] STARTTLS FALLO: %v", err)
			return fmt.Errorf("ambos metodos fallaron - TLS y STARTTLS: %w", err)
		}
	}

	log.Printf("[EMAIL] Email enviado exitosamente a: %s", toEmail)
	return nil
}

// sendWithTLS usa conexion TLS directa (puerto 465)
func (s *EmailService) sendWithTLS(toEmail, message string) error {
	addr := s.host + ":465"

	// Timeout de 10 segundos para la conexion
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	tlsConfig := &tls.Config{ServerName: s.host}

	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("conexion TLS a %s: %w", addr, err)
	}
	defer conn.Close()

	// Timeout de 10 segundos para operaciones SMTP
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("cliente SMTP: %w", err)
	}
	defer client.Quit()

	auth := smtp.PlainAuth("", s.email, s.password, s.host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("autenticacion: %w", err)
	}

	if err := client.Mail(s.email); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}
	if err := client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("RCPT TO: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}
	if _, err := w.Write([]byte(message)); err != nil {
		return fmt.Errorf("escribiendo mensaje: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("cerrando writer: %w", err)
	}

	return nil
}

// sendWithSTARTTLS usa STARTTLS (puerto 587)
func (s *EmailService) sendWithSTARTTLS(toEmail, message string) error {
	addr := s.host + ":" + s.port

	// Timeout de 10 segundos para la conexion
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("conexion a %s: %w", addr, err)
	}

	// Timeout de 10 segundos para operaciones SMTP
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("cliente SMTP: %w", err)
	}
	defer client.Quit()

	tlsConfig := &tls.Config{ServerName: s.host}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("STARTTLS: %w", err)
	}

	auth := smtp.PlainAuth("", s.email, s.password, s.host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("autenticacion: %w", err)
	}

	if err := client.Mail(s.email); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}
	if err := client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("RCPT TO: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}
	if _, err := w.Write([]byte(message)); err != nil {
		return fmt.Errorf("escribiendo mensaje: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("cerrando writer: %w", err)
	}

	return nil
}

const passwordResetTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; background-color: #f5f0eb; padding: 20px;">
  <div style="max-width: 500px; margin: 0 auto; background: white; border-radius: 12px; padding: 30px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
    <h1 style="color: #6F4E37; text-align: center;">Cheos Cafe</h1>
    <p>Hola <strong>{NOMBRE}</strong>,</p>
    <p>Recibimos una solicitud para restablecer la contrasena de tu cuenta.</p>
    <p>Haz clic en el siguiente boton para crear una nueva contrasena:</p>
    <div style="text-align: center; margin: 25px 0;">
      <a href="{RESET_URL}" style="background-color: #6F4E37; color: white; padding: 12px 30px; text-decoration: none; border-radius: 8px; font-size: 16px;">
        Restablecer contrasena
      </a>
    </div>
    <p style="color: #888; font-size: 13px;">Este enlace expira en 15 minutos.</p>
    <p style="color: #888; font-size: 13px;">Si no solicitaste este cambio, puedes ignorar este correo. Tu contrasena no sera modificada.</p>
    <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
    <p style="color: #aaa; font-size: 11px; text-align: center;">Cheos Cafe - Cafe de especialidad colombiano</p>
  </div>
</body>
</html>`
