package services

import (
	"auth_service/internal/config"
	"fmt"
	"strconv"

	"gopkg.in/gomail.v2"
)

// SendWarningMsg отправляет предупреждающее сообщение на указанный email.
// Сообщение содержит информацию о попытке обновления токена с нового IP-адреса.
func SendWarningMsg(userEmail string, issuedIp string, ip string) error {
	msg := gomail.NewMessage()

	msg.SetHeader("From", config.SenderEmail)
	msg.SetHeader("To", userEmail)
	msg.SetHeader("Subject", "Подозрительная активность: IP-адрес изменён")

	body := fmt.Sprintf(`
        <p>Была предпринята попытка обновления токена с нового IP-адреса:</p>
        <ul>
            <li><strong>Старый IP:</strong> %s</li>
            <li><strong>Новый IP:</strong> %s</li>
        </ul>
        <p>Если это были не вы, рекомендуем сменить пароль и завершить все активные сессии.</p>
    `, issuedIp, ip)

	msg.SetBody("text/html", body)

	smtPort, err := strconv.Atoi(config.SmtpPort)
	if err != nil {
		return fmt.Errorf("failed to converte 'smtPort': %w", err)
	}

	serv := gomail.NewDialer(config.SmtpHost, smtPort, config.SenderEmail, config.PasswordEmail)
	if err := serv.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email message^ %w", err)
	}

	return nil
}

// CheckConfigVar проверяет наличие необходимых переменных окружения для отправки email.
// Если какая-либо переменная не установлена, возвращается ошибка с описанием отсутствующей переменной.
func CheckConfigVar() error {
	var err error

	switch {
	case config.SenderEmail == "":
		err = fmt.Errorf("'SENDER_EMAIL' is not set in the environment variables")
	case config.SmtpPort == "":
		err = fmt.Errorf("'SMTP_PORT' is not set in the environment variables")
	case config.SmtpHost == "":
		err = fmt.Errorf("'SMTP_HOST' is not set in the environment variables")
	case config.PasswordEmail == "":
		err = fmt.Errorf("'PASSWORD_EMAIL' is not set in the environment variables")
	}

	return err
}
