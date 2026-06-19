package auth

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

type EmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewEmailService(host, port, username, password, from string) *EmailService {
	return &EmailService{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *EmailService) SendVerificationEmail(to, token string) error {
	// Если SMTP не настроен, используем dev mode
	if s.username == "" || s.password == "" {
		log.Printf("[DEV MODE] Email to %s: Подтверждение регистрации", to)
		log.Printf("[DEV MODE] Verification link: http://localhost:5173/verify-email?token=%s", token)
		return nil
	}

	subject := "Подтверждение регистрации на бирже валют"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; padding: 20px;">
			<h2>Добро пожаловать!</h2>
			<p>Для подтверждения вашего email адреса, пожалуйста, перейдите по ссылке:</p>
			<p>
				<a href="http://localhost:5173/verify-email?token=%s" 
				   style="background-color: #2563eb; color: white; padding: 10px 20px; 
						  text-decoration: none; border-radius: 5px; display: inline-block;">
					Подтвердить Email
				</a>
			</p>
			<p>Или скопируйте эту ссылку в браузер:</p>
			<p style="word-break: break-all;">http://localhost:5173/verify-email?token=%s</p>
			<p>Ссылка действительна в течение 24 часов.</p>
			<hr>
			<p style="color: #666; font-size: 12px;">Если вы не регистрировались, просто проигнорируйте это письмо.</p>
		</body>
		</html>
	`, token, token)

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