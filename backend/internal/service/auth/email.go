package auth

import (
	"fmt"
	"log"
	"net/smtp"
	"strconv"
)

type EmailService struct {
	host        string
	port        string
	username    string
	password    string
	from        string
	frontendURL string
}

func NewEmailService(host, port, username, password, from, frontendURL string) *EmailService {
	return &EmailService{
		host:        host,
		port:        port,
		username:    username,
		password:    password,
		from:        from,
		frontendURL: frontendURL,
	}
}

func (s *EmailService) SendVerificationEmail(to, token string) error {
	// Валидация порта если SMTP настроен
	if s.username != "" && s.password != "" {
		if s.port == "" {
			return fmt.Errorf("SMTP port is not configured")
		}
		if _, err := strconv.Atoi(s.port); err != nil {
			return fmt.Errorf("invalid SMTP port: %s", s.port)
		}
	}

	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.frontendURL, token)

	// Если SMTP не настроен, используем dev mode
	if s.username == "" || s.password == "" {
		log.Printf("[DEV MODE] Email to %s: Подтверждение регистрации", to)
		log.Printf("[DEV MODE] Verification link: %s", verifyURL)
		return nil
	}

	subject := "Подтверждение регистрации на бирже валют"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Подтверждение регистрации</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2c3e50;">Подтверждение регистрации</h1>
        <h2 style="color: #34495e;">Добро пожаловать!</h2>
        <p>Для подтверждения вашего email адреса, пожалуйста, перейдите по ссылке:</p>
        <p style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #3498db; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Подтвердить Email</a>
        </p>
        <p>Или скопируйте эту ссылку в браузер:</p>
        <p style="background-color: #f8f9fa; padding: 10px; border-radius: 3px; word-break: break-all;">%s</p>
        <p style="color: #7f8c8d; font-size: 14px;">Ссылка действительна в течение 24 часов.</p>
        <hr style="border: none; border-top: 1px solid #ecf0f1; margin: 20px 0;">
        <p style="color: #95a5a6; font-size: 12px;">Если вы не регистрировались, просто проигнорируйте это письмо.</p>
    </div>
</body>
</html>
`, verifyURL, verifyURL)

	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", s.from, to, subject, body)

	addr := s.host + ":" + s.port
	err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return err
	}

	log.Printf("Verification email sent to %s", to)
	return nil
}

// SendAlertEmail отправляет уведомление о срабатывании алерта
func (s *EmailService) SendAlertEmail(to, title, message string) error {
	// Валидация порта если SMTP настроен
	if s.username != "" && s.password != "" {
		if s.port == "" {
			return fmt.Errorf("SMTP port is not configured")
		}
		if _, err := strconv.Atoi(s.port); err != nil {
			return fmt.Errorf("invalid SMTP port: %s", s.port)
		}
	}

	// Если SMTP не настроен, используем dev mode
	if s.username == "" || s.password == "" {
		log.Printf("[DEV MODE] Alert email to %s: %s", to, title)
		log.Printf("[DEV MODE] Message: %s", message)
		return nil
	}

	subject := fmt.Sprintf("🔔 %s", title)
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #e74c3c;">🔔 %s</h1>
        <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0;">
            <p style="font-size: 16px; margin: 0;">%s</p>
        </div>
        <p style="color: #7f8c8d; font-size: 14px;">Вы получили это уведомление, потому что настроили алерт на изменение курса валюты.</p>
        <hr style="border: none; border-top: 1px solid #ecf0f1; margin: 20px 0;">
        <p style="color: #95a5a6; font-size: 12px;">Если вы не хотите получать такие уведомления, измените настройки в личном кабинете.</p>
    </div>
</body>
</html>
`, title, title, message)

	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", s.from, to, subject, body)

	addr := s.host + ":" + s.port
	err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("Failed to send alert email to %s: %v", to, err)
		return err
	}

	log.Printf("Alert email sent to %s: %s", to, title)
	return nil
}
