//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	sender2 "li-acc/pkg/sender"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// mailHogEnv - helper struct for "MailHog" mock SmtpPort server
type mailHogEnv struct {
	c        tc.Container
	Host     string
	SmtpPort string
	APIPort  string
	Close    func()
}

// startMailHog runs MailHog in a testcontainer
func startMailHog(t *testing.T) *mailHogEnv {
	t.Helper()

	ctx := context.Background()

	req := tc.ContainerRequest{
		Image:        "mailhog/mailhog:v1.0.1",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("1025/tcp"),
			wait.ForListeningPort("8025/tcp"),
		).WithDeadline(30 * time.Second),
	}

	cont, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// get host, port, and APIPort port
	host, err := cont.Host(ctx)
	require.NoError(t, err)

	smtpPort, err := cont.MappedPort(ctx, "1025/tcp")
	require.NoError(t, err)

	apiPort, err := cont.MappedPort(ctx, "8025/tcp")
	require.NoError(t, err)

	return &mailHogEnv{
		c:        cont,
		Host:     host,
		SmtpPort: smtpPort.Port(),
		APIPort:  apiPort.Port(),
		Close: func() {
			_ = cont.Terminate(ctx)
		},
	}
}

type MailHogItems []struct {
	Content struct {
		Headers map[string][]string `json:"headers"`
		Body    string              `json:"body"`
	} `json:"Content"`
}

// getMessagesFromMailHog fetches all messages from MailHog APIPort
func getMessagesFromMailHog(t *testing.T, env *mailHogEnv) MailHogItems {
	t.Helper()

	resp, err := http.Get(fmt.Sprintf("http://%s:%s/api/v2/messages", env.Host, env.APIPort))
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var parsed struct {
		TotalMsg int          `json:"totalMsg"`
		Items    MailHogItems `json:"items"`
	}
	err = json.Unmarshal(body, &parsed)
	require.NoError(t, err)

	return parsed.Items
}

// ---- Tests ----
func TestSendEmail_SingleRecipient(t *testing.T) {
	env := startMailHog(t)
	defer env.Close()

	senderEmail := "sender@test.com"
	recEmail := "rec@test.com"

	sender := sender2.NewSender(env.Host, mustAtoi(env.SmtpPort), senderEmail, "", false)

	subject := "Single test"
	body := "Hello from testcontainers single"
	msgStatus := sender2.FormMessage(subject, body, "", senderEmail, recEmail)

	storeForSender := false

	statusCh := make(chan sender2.EmailStatus, 1)
	var wg sync.WaitGroup
	wg.Add(1)

	go sender.SendEmail(msgStatus.Msg, statusCh, &wg, storeForSender)
	wg.Wait()

	res := <-statusCh
	require.Empty(t, res.Status)
	require.Equal(t, sender2.SuccessType, res.StatusType)

	items := getMessagesFromMailHog(t, env)
	require.NotEmpty(t, items)

	headers := items[0].Content.Headers
	require.Contains(t, headers["Subject"], subject)
	require.Contains(t, strings.Split(headers["To"][0], ", "), recEmail)
	if storeForSender {
		require.Contains(t, strings.Split(headers["To"][0], ", "), senderEmail)
	}

	require.Contains(t, items[0].Content.Body, body)
}

func TestSendEmail_MultipleRecipientsParallel(t *testing.T) {
	env := startMailHog(t)
	defer env.Close()

	senderEmail := "sender@test.com"
	recNum := 10 // number of recipients
	// generate list of recipients
	var recEmails []string
	for i := 0; i < recNum; i++ {
		recEmails = append(recEmails, fmt.Sprintf("rec%d@test.com", i+1))
	}

	sender := sender2.NewSender(env.Host, mustAtoi(env.SmtpPort), senderEmail, "", false)

	subject := "Multiple test"
	body := "Hello from testcontainers multiple"

	storeForSender := true

	statusCh := make(chan sender2.EmailStatus, len(recEmails))
	var wg sync.WaitGroup

	for _, rec := range recEmails {
		wg.Add(1)

		msgStatus := sender2.FormMessage(subject, body, "", senderEmail, rec)

		go sender.SendEmail(msgStatus.Msg, statusCh, &wg, storeForSender)
	}

	wg.Wait()
	close(statusCh)

	// check that all status messages are empty, and all statuses are success
	for status := range statusCh {
		require.Empty(t, status.Status)
		require.Equal(t, sender2.SuccessType, status.StatusType)
	}

	// check that amount of items = amount of recipients
	items := getMessagesFromMailHog(t, env)
	require.Len(t, items, len(recEmails))

	// check that all recipients received mails
	foundRecs := map[string]bool{}
	for _, item := range items {
		// check the subject and body
		headers := item.Content.Headers
		require.Contains(t, headers["Subject"], subject)
		require.Contains(t, item.Content.Body, body)

		// headers in gomail message are represented in one string, separated with ', '
		// so split the string
		receivers := strings.Split(headers["To"][0], ", ")
		for _, r := range receivers {
			foundRecs[r] = true
		}
	}

	for _, r := range recEmails {
		require.True(t, foundRecs[r], "missing recipient %s", r)
	}
}

// mustAtoi helper to avoid boilerplate in tests
func mustAtoi(s string) int {
	var port int
	_, err := fmt.Sscanf(s, "%d", &port)
	if err != nil {
		panic(err)
	}
	return port
}
