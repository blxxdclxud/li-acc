package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// MailHogMessage - MailHog API structures
type MailHogMessage struct {
	ID   string `json:"ID"`
	From struct {
		Mailbox string `json:"Mailbox"`
		Domain  string `json:"Domain"`
	} `json:"From"`
	To []struct {
		Mailbox string `json:"Mailbox"`
		Domain  string `json:"Domain"`
	} `json:"To"`
	Content struct {
		Headers map[string][]string `json:"Headers"`
		Body    string              `json:"Body"`
	} `json:"Content"`
	MIME struct {
		Parts []struct {
			Headers struct {
				ContentType        []string `json:"Content-Type"`
				ContentDisposition []string `json:"Content-Disposition"`
			} `json:"Headers"`
			Body string `json:"Body"`
		} `json:"Parts"`
	} `json:"MIME"`
}

type MailHogResponse struct {
	Total int              `json:"total"`
	Count int              `json:"count"`
	Start int              `json:"start"`
	Items []MailHogMessage `json:"items"`
}

// getMailHogEmails fetches all emails from MailHog with pagination
func getMailHogEmails(t *testing.T, env *TestEnvironment) []MailHogMessage {
	t.Helper()

	var allEmails []MailHogMessage
	start := 0
	limit := 250 // MailHog's maximum per request

	for {
		url := fmt.Sprintf("%s/api/v2/messages?start=%d&limit=%d",
			env.MailHogAPIURL, start, limit)

		resp, err := http.Get(url)
		require.NoError(t, err, "Failed to fetch MailHog emails")

		var result MailHogResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		require.NoError(t, err, "Failed to decode MailHog response")

		// Append current page
		allEmails = append(allEmails, result.Items...)

		// Log progress for large fetches
		if result.Total > 100 {
			t.Logf("Fetched %d/%d emails from MailHog", len(allEmails), result.Total)
		}

		// Stop if no more items or we've fetched all
		if len(result.Items) == 0 || len(allEmails) >= result.Total {
			break
		}

		// Move to next page
		start += limit
	}

	return allEmails
}

// clearMailHog deletes all emails from MailHog
func clearMailHog(t *testing.T, env *TestEnvironment) {
	t.Helper()

	req, err := http.NewRequest("DELETE", env.MailHogAPIURL+"/api/v1/messages", nil)
	require.NoError(t, err, "Failed to create DELETE request")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "Failed to clear MailHog")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode,
		"MailHog clear should return 200, got %d", resp.StatusCode)

	// Verify MailHog is actually empty
	emails := getMailHogEmails(t, env)
	require.Empty(t, emails, "MailHog should be empty after clearing")

	t.Log("MailHog cleared successfully")
}

// findEmailByRecipient finds email sent to specific recipient
func findEmailByRecipient(messages []MailHogMessage, email string) *MailHogMessage {
	for _, msg := range messages {
		for _, to := range msg.To {
			if to.Mailbox+"@"+to.Domain == email {
				return &msg
			}
		}
	}
	return nil
}

// countEmailAttachments counts PDF attachments in an email
func countEmailAttachments(msg *MailHogMessage) int {
	count := 0
	for _, part := range msg.MIME.Parts {
		// Check if this part is an attachment
		for _, ct := range part.Headers.ContentType {
			// Match "application/pdf" with or without parameters
			if strings.HasPrefix(ct, "application/pdf") {
				count++
				break // Count this part only once
			}
		}
	}
	return count
}
