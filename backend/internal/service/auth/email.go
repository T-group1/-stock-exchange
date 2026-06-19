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
	frontendURL string // <--- ДОБАВЛЕНО
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
<body>
    <h2>Добро пожаловать!</h2>
    <p>Для подтверждения вашего email адреса, пожалуйста, перейдите по ссылке:</p>
    <a href="%s">Подтвердить Email</a>
    <p>Или скопируйте эту ссылку в браузер:</p>
    <p>%s</p>
    <p>Ссылка действительна в течение 24 часов.</p>
    <hr>
    <p>Если вы не регистрировались, просто проигнорируйте это письмо.</p>
</body>
</html>
`, verifyURL, verifyURL)

	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	// ИСПРАВЛЕНО: rn заменено на \r\n (стандарт SMTP)
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
