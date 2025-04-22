package config

import (
	"os"
	"time"

	"golang.org/x/time/rate"
)

var (
	ServiceSocket = os.Getenv("SERVICE_SOCKET") // путь к сокету сервиса, задается через переменную окружения SERVICE_SOCKET
	Mode          = os.Getenv("MODE")           // режим работы сервиса, задается через переменную окружения MODE
	PsqlUrl       = os.Getenv("PSQL_URL")       // URL для подключения к PostgreSQL, задается через переменную окружения PSQL_URL
	Secret        = os.Getenv("SECRET")         // секретный ключ для шифрования, задается через переменную окружения SECRET

	SenderEmail   = os.Getenv("SENDER_EMAIL")   // email отправителя, задается через переменную окружения SENDER_EMAIL
	PasswordEmail = os.Getenv("PASSWORD_EMAIL") // пароль для email отправителя, задается через переменную окружения PASSWORD_EMAIL
	SmtpHost      = os.Getenv("SMTP_HOST")      // хост SMTP сервера, задается через переменную окружения SMTP_HOST
	SmtpPort      = os.Getenv("SMTP_PORT")      // порт SMTP сервера, задается через переменную окружения SMTP_PORT

	RateLimit   = rate.Limit(20) // ограничение RPS (запросов в секунду) для пользователя
	BufferLimit = 40             // емкость "ведра" запросов, которые могут обрабатываться поверх RPS ограничения за раз

	CleanupInterval = 1 * time.Minute // интервал для чистки словаря с лимитерами неактивных пользователей
	InactivityLimit = 5 * time.Minute // время, через которое пользователь становится неактивным
)
