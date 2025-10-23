package service

import (
	"context"
	"errors"
	"fmt"
	"li-acc/internal/model"
	"li-acc/pkg/sender"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/gomail.v2"
)

//
// ===== Mock implementations =====
//

// mockSender implements sender.MailSender for testing.
type mockSender struct {
	results map[string]error // recipient -> error (nil = success)
}

func (m *mockSender) SendEmail(msg *gomail.Message, status chan sender.EmailStatus, _ bool) {
	to := msg.GetHeader("To")
	recipient := to[0]
	if err, ok := m.results[recipient]; ok && err != nil {
		status <- sender.EmailStatus{
			Status:    sender.Error,
			StatusMsg: fmt.Sprintf("failed to send to %s", recipient),
			Cause:     err,
			Msg:       msg,
		}
		return
	}

	status <- sender.EmailStatus{
		Status: sender.Success,
		Msg:    msg,
	}
}

func (m *mockSender) GetSenderEmail() string {
	return "mock@sender.com"
}

//
// ===== Helper builder =====
//

func newTestMail(recipients ...string) model.Mail {
	return model.Mail{
		Subject: "Test Subject",
		Body:    "Hello there!",
		From:    "mock@sender.com",
		To:      recipients,
		AttachmentPaths: map[string]string{
			"ok@example.com":   "/tmp/ok.pdf",
			"fail@example.com": "/tmp/fail.pdf",
		},
	}
}

//
// ===== Tests =====
//

func TestSendMails_AllSuccess(t *testing.T) {
	ctx := context.Background()
	mock := &mockSender{results: map[string]error{
		"ok@example.com": nil,
	}}
	s := &mailService{sender: mock}

	mail := newTestMail("ok@example.com")
	sentCount, err := s.SendMails(ctx, mail)
	require.NoError(t, err)
	require.Equal(t, sentCount, 1)
}

func TestSendMails_SomeFailures(t *testing.T) {
	ctx := context.Background()
	mock := &mockSender{results: map[string]error{
		"ok@example.com":   nil,
		"fail@example.com": errors.New("smtp timeout"),
	}}

	s := &mailService{sender: mock}
	mail := newTestMail("ok@example.com", "fail@example.com")

	sentCount, err := s.SendMails(ctx, mail)
	require.Error(t, err)
	require.Equal(t, sentCount, 1)

	var sendErr *EmailSendingError
	require.ErrorAs(t, err, &sendErr)
	require.Len(t, sendErr.MapReceiverCause, 1)
	require.Contains(t, sendErr.MapReceiverCause, "fail@example.com")
	require.Equal(t, "smtp timeout", sendErr.MapReceiverCause["fail@example.com"])
}

func TestSendMails_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	mock := &mockSender{results: map[string]error{}}
	s := &mailService{sender: mock}

	mail := newTestMail("a@example.com", "b@example.com")

	sentCount, err := s.SendMails(ctx, mail)
	require.Error(t, err)
	require.Contains(t, err.Error(), "operation canceled")
	require.Equal(t, sentCount, 0)
}

func TestSendMails_NoRecipients(t *testing.T) {
	ctx := context.Background()
	mock := &mockSender{}
	s := &mailService{sender: mock}

	mail := newTestMail() // empty "To"
	mail.To = nil

	sentCount, err := s.SendMails(ctx, mail)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no recipients")
	require.Equal(t, sentCount, 0)
}

func TestSendMails_AttachmentMissing(t *testing.T) {
	ctx := context.Background()
	mock := &mockSender{results: map[string]error{
		"ok@example.com": nil,
	}}
	s := &mailService{sender: mock}

	mail := newTestMail("ok@example.com")
	// Remove attachment to simulate GetAttachmentPath error
	mail.AttachmentPaths = nil

	sentCount, err := s.SendMails(ctx, mail)
	// fatal error, since attachment is the most important in the mail
	require.Error(t, err)
	require.Equal(t, sentCount, 0)
}

func TestGetSenderEmail(t *testing.T) {
	mock := &mockSender{}
	s := &mailService{sender: mock}
	require.Equal(t, "mock@sender.com", s.GetSenderEmail())
}

//
// ===== Edge Case: multiple failures aggregated =====
//

func TestSendMails_MultipleFailuresAggregated(t *testing.T) {
	ctx := context.Background()
	mock := &mockSender{results: map[string]error{
		"a@example.com": errors.New("auth error"),
		"b@example.com": errors.New("conn reset"),
	}}

	s := &mailService{sender: mock}
	mail := newTestMail("a@example.com", "b@example.com")

	sentCount, err := s.SendMails(ctx, mail)
	require.Error(t, err)
	require.Equal(t, sentCount, 0)

	var sendErr *EmailSendingError
	require.ErrorAs(t, err, &sendErr)
	require.Len(t, sendErr.MapReceiverCause, 2)
}
