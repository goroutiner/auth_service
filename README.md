[![Authentication service](https://github.com/goroutiner/auth_service/actions/workflows/ci-cd.yaml/badge.svg?branch=main)](https://github.com/alexey0b/auth_service/actions/workflows/ci-cd.yaml) 
[![codecov](https://codecov.io/gh/alexey0b/auth_service/graph/badge.svg)](https://codecov.io/gh/goroutiner/auth_service)


<h3 align="center">
  <div align="center">
    <h1>Authentication Service 🔐</h1>
  </div>
</h3>

**Authentication Service** — это сервис для управления аутентификацией пользователей. Он предоставляет API для генерации и обновления токенов доступа.

---

## 📋 Возможности

- Генерация токенов доступа для пользователей.
- Обновление токенов доступа.
- Поддержка двух режимов хранения данных:
  - **in-memory** (для демонстрации или тестирования).
  - **PostgreSQL** (для продакшн-окружения).

---

## 🔥 API Эндпоинты

1️⃣ **Генерация токенов**

**GET** `/api/auth/{user_id}`

**Тело ответа**:

```json
{
  "access_token": "your_access_token",
  "refresh_token": "your_refresh_token"
}
```

2️⃣ **Обновление токенов**

**POST** `/api/auth/refresh`

**Тело запроса**:

```json
{
  "access_token": "your_access_token",
  "refresh_token": "your_refresh_token"
}
```

**Тело ответа**:

```json
{
  "access_token": "new_access_token",
  "refresh_token": "new_refresh_token"
}
```

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
  MAX_TOKENS_PER_USER: 5 # максимальное количество активных refresh-токенов для одного пользователя
  RATE_LIMIT: 20 # значение RPS на пользователя
  BUFFER_LIMIT: 40 # вместимость буфера запросов
  CLEANUP_INTERVAL: 1 # интервал для чистки словаря с лимитерами неактивных пользователей (в минутах)
  INACTIVITY_LIMIT: 5 # период неактивности пользователя (в минутах)
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
2. Скопируйте репозиторий:

```sh
git clone https://github.com/goroutiner/auth_service.git
cd auth_service
```

3. Запуск сервиса выполняется с помощью команды:

```sh
make run
```

---

## ✅⭕ Инструкция по запуску тестов

_Перед запуском тестов должен быть запущен Docker!_

- Для запуска тестирования `handlers` выполните команду:

```sh
make test-handlers
```

- Для запуска тестирования `services` выполните команду:

```sh
make test-services
```

- Для запуска тестирования `database` выполните команду:

```sh
make test-database
```

- Для запуска тестирования `memory` выполните команду:

```sh
make test-memory
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
  - [testcontainers/testcontainers-go](https://github.com/testcontainers) для запуска тестовых контейнеров
  - [mailhog/MailHog](https://github.com/mailhog) для тестирования отправки писем по электронной почте
