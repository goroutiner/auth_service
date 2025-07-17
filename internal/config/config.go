package config

import (
	"os"
)

var (
	ServiceSocket = os.Getenv("SERVICE_SOCKET") // Путь к сокету сервиса, задается через переменную окружения SERVICE_SOCKET.
	Mode          = os.Getenv("MODE")           // Режим работы сервиса, задается через переменную окружения MODE.
	PsqlUrl       = os.Getenv("PSQL_URL")       // URL для подключения к PostgreSQL, задается через переменную окружения PSQL_URL.
	Secret        = os.Getenv("SECRET")         // Секретный ключ для шифрования, задается через переменную окружения SECRET.

	SenderEmail   = os.Getenv("SENDER_EMAIL")   // Email отправителя, задается через переменную окружения SENDER_EMAIL.
	PasswordEmail = os.Getenv("PASSWORD_EMAIL") // Пароль для email отправителя, задается через переменную окружения PASSWORD_EMAIL.
	SmtpHost      = os.Getenv("SMTP_HOST")      // Хост SMTP сервера, задается через переменную окружения SMTP_HOST.
	SmtpPort      = os.Getenv("SMTP_PORT")      // Порт SMTP сервера, задается через переменную окружения SMTP_PORT.

	MaxTokensPerUser = os.Getenv("MAX_TOKENS_PER_USER") // Максимальное количество активных refresh-токенов для одного пользователя.
	RateLimit        = os.Getenv("RATE_LIMIT")          // Ограничение RPS (запросов в секунду) для пользователя.
	BufferLimit      = os.Getenv("BUFFER_LIMIT")        // Ёмкость "ведра" запросов, которые могут обрабатываться поверх RPS ограничения за раз.

	CleanupInterval = os.Getenv("CLEANUP_INTERVAL") // Интервал для чистки словаря с лимитерами неактивных пользователей (в минутах).
	InactivityLimit = os.Getenv("INACTIVITY_LIMIT") // Время, через которое пользователь становится неактивным (в минутах).
)
