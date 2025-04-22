# Authentication Service 🔐

**Authentication Service** — это сервис для управления аутентификацией пользователей. Он предоставляет API для генерации и обновления токенов доступа.

---

## 📋 Возможности

- Генерация токенов доступа для пользователей.
- Обновление токенов доступа.
- Поддержка двух режимов хранения данных:
  - **in-memory** (для демонстрации или тестирования).
  - **PostgreSQL** (для продакшн-окружения).

---

### 🔧 Предварительная настройка переменных окружений в файле `compose.yaml`:

- Переменные окружения для **auth_service**:
```yaml
environment:
      SERVICE_SOCKET: ":8080" # сокет соединения, на котором работает сервис 
      MODE: "postgres" # режим работы сервиса 
      PSQL_URL: "postgres://root:password@postgres:5432/mydb?sslmode=disable" # адрес Postgres
      SECRET: "secret" # секрет для подписи jwt токена
      SENDER_EMAIL: "" # email, с которого будут отправлятся предупреждения пользователям
      PASSWORD_EMAIL: "" # пароль от почты 
      SMTP_HOST: "" # адрес хоста, на котором развернут SMTP-сервер
```

- Переменные окружения для **postgres**:

```yaml
environment:
      POSTGRES_USER: "root"
      POSTGRES_PASSWORD: "password"
      POSTGRES_DB: "mydb"
```

---

## 🐳 Запуск через Docker Compose

1. Убедитесь, что у вас установлен **Docker** и **Docker Compose**.
2. Склонируйте репозиторий:

   ```sh
   git clone https://github.com/goroutiner/auth_service.git
   cd auth_service
   ```

---

## ✅⭕ Инструкция по запуску тестов

- Для запуска unit-тестов выполните команду:

```sh
make unit-tests
```
- Для запуска integration-тестов убедитесь, что запущен Docker и выполните команду:

```sh
make integration-tests
```
---

## 🛠️ Технические ресурсы

- **Язык программирования**: Go (Golang)
- **База данных**: PostgreSQL (опционально)
- **Библиотеки**:
  - [jmoiron/sqlx](https://github.com/jmoiron/sqlx) для взаимодействия с базами данных.
  - [github.com/jackc/pgx/v5/stdlib](https://github.com/jackc/pgx) и [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) драйвера для PosgreSQL и SQLite
  - [golang-jwt/jwt](https://github.com/golang-jwt/jwt) для работы с JWT-токенами
  - [go-gomail/gomail](https://github.com/go-gomail/gomail) для отправки предупреждающих сообщений пользователям 
  - [stretchr/testify](https://github.com/stretchr/testify) для написания тестов
  - [vektra/mockery](https://github.com/vektra/mockery) для генерации mocks
  - [testcontainers/testcontainers-go](https://github.com/testcontainers/testcontainers-go?tab=readme-ov-file) для запуска тестовых контейнеров
