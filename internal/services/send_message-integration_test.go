package services_test

import (
	"auth_service/internal/config"
	"auth_service/internal/services"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/quotedprintable"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Messages представляет структуру для хранения писем, полученных из Mailhog API.
type Messages struct {
	Total float64 `json:"Total"` // общее количество писем
	Items []struct {
		Content struct {
			Headers struct {
				From []string `json:"From"` // отправитель письма
				To   []string `json:"To"`   // получатель письма
			} `json:"Headers"`
			Body string `json:"Body"` // тело письма
		} `json:"Content"`
	} `json:"items"`
}

var (
	host     string
	webPort  string
	smtpPort string
)

// TestMain запускает контейнер Mailhog для тестов и завершает его после выполнения всех тестов.
func TestMain(m *testing.M) {
	ctx := context.Background()
	buildContext, err := filepath.Abs("./docker")
	if err != nil {
		log.Fatalf("Failed to resolve absolute path: %v\n", err)
	}

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context: buildContext,
		},
		WaitingFor: wait.ForListeningPort("8025/tcp"),
	}

	mailhogC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatal(err)
	}

	host, err = mailhogC.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}

	mappedPort, err := mailhogC.MappedPort(ctx, "1025")
	if err != nil {
		log.Fatal(err)
	}
	smtpPort = mappedPort.Port()

	config.SmtpPort = mappedPort.Port()
	config.SmtpHost = host
	config.SenderEmail = "security@gmail.com"

	mappedWebPort, err := mailhogC.MappedPort(ctx, "8025")
	if err != nil {
		log.Fatal(err)
	}
	webPort = mappedWebPort.Port()

	code := m.Run()

	mailhogC.Terminate(ctx)
	os.Exit(code)
}

// TestSendWarningMsg проверяет функцию отправки предупреждающего письма пользователю.
func TestSendWarningMsg(t *testing.T) {
	testUserEmail := "user@gmail.com"
	issuedIp := "122.124.123"
	ip := "122.124.129"
	t.Run("successful sending", func(t *testing.T) {
		msgs, err := getMsgs()
		require.NoError(t, err)

		err = services.SendWarningMsg(testUserEmail, issuedIp, ip)
		require.NoError(t, err)

		actualMsgs, err := getMsgs()
		require.NoError(t, err)
		require.Greater(t, len(actualMsgs.Items), 0, "Письмо не было отправлено")

		require.Equal(t, msgs.Total+1, actualMsgs.Total)
	})

	t.Run("invalid smtp port", func(t *testing.T) {
		defer updateConfig()

		config.SmtpPort = "not_a_number"
		err := services.SendWarningMsg(testUserEmail, issuedIp, ip)
		require.ErrorContains(t, err, "failed to converte 'smtPort'")
	})

	t.Run("check email body and recipient", func(t *testing.T) {
		err := services.SendWarningMsg(testUserEmail, issuedIp, ip)
		require.NoError(t, err)

		msgs, err := getMsgs()
		require.NoError(t, err)
		require.Greater(t, len(msgs.Items), 0, "Письмо не было отправлено")

		headers := msgs.Items[len(msgs.Items)-1].Content.Headers
		require.Equal(t, testUserEmail, headers.To[0])
		require.Equal(t, config.SenderEmail, headers.From[0])

		contentBody := msgs.Items[len(msgs.Items)-1].Content.Body
		body, err := io.ReadAll(quotedprintable.NewReader(strings.NewReader(contentBody)))
		require.NoError(t, err)
		expectedLines := []string{
			"Была предпринята попытка обновления токена с нового IP-адреса:",
			fmt.Sprintf("Старый IP:</strong> %s", issuedIp),
			fmt.Sprintf("Новый IP:</strong> %s", ip),
		}
		for _, line := range expectedLines {
			require.Contains(t, string(body), line)
		}
	})
}

// TestCheckConfigVar проверяет функцию checkConfigVar на отсутствие обязательных переменных конфигурации.
func TestCheckConfigVar(t *testing.T) {
	t.Run("empty sender email", func(t *testing.T) {
		defer updateConfig()

		config.SenderEmail = ""
		config.SmtpHost = host
		config.SmtpPort = webPort
		config.PasswordEmail = "secret"
		err := services.CheckConfigVar()
		require.ErrorContains(t, err, "'SENDER_EMAIL' is not set in the environment variables")
	})

	t.Run("empty smtp host", func(t *testing.T) {
		defer updateConfig()

		config.SmtpHost = ""
		config.SmtpPort = webPort
		config.SenderEmail = "security@gmail.com"
		config.PasswordEmail = "secret"
		err := services.CheckConfigVar()
		require.ErrorContains(t, err, "'SMTP_HOST' is not set in the environment variables")
	})

	t.Run("empty smptp port", func(t *testing.T) {
		defer updateConfig()

		config.SmtpPort = ""
		config.SmtpHost = host
		config.SenderEmail = "security@gmail.com"
		config.PasswordEmail = "secret"

		err := services.CheckConfigVar()
		require.ErrorContains(t, err, "'SMTP_PORT' is not set in the environment variables")
	})

	t.Run("empty password email", func(t *testing.T) {
		defer updateConfig()

		config.PasswordEmail = ""
		config.SmtpHost = host
		config.SmtpPort = webPort
		config.SenderEmail = "security@gmail.com"

		err := services.CheckConfigVar()
		require.ErrorContains(t, err, "'PASSWORD_EMAIL' is not set in the environment variables")
	})
}

// getMsgs получает содержимое письма в декодированном виде.
func getMsgs() (*Messages, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%s/api/v2/messages", host, webPort))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var msgs Messages
	if err := json.NewDecoder(resp.Body).Decode(&msgs); err != nil {
		return nil, err
	}

	return &msgs, nil
}

// updateConfig обновляет параметры конфигурации на исходные.
func updateConfig() {
	config.SmtpPort = smtpPort
	config.SmtpHost = host
	config.SenderEmail = "security@gmail.com"
	config.PasswordEmail = "secret"
}
