package service

import (
	"context"
	"errors"
	"fmt"
	"li-acc/internal/model"
	"li-acc/pkg/logger"
	"li-acc/pkg/sender"
	"sync"
	"time"

	"go.uber.org/zap"
)

// defaultMaxParallel is the default limit of concurrent SMTP sends.
// Consider making this configurable (env/config) for production.
const defaultMaxParallel = 10

type MailService interface {
	SendMails(ctx context.Context, mail model.Mail) error
	GetSenderEmail() string
}
type mailService struct {
	sender sender.MailSender
}

func NewMailService(smtp model.SMTP) MailService {
	return &mailService{
		sender: sender.NewSender(smtp.Host, smtp.Port, smtp.Email, smtp.Password, true),
	}
}

// SendMails sends emails in parallel with controlled concurrency.
// It validates input, logs every step, measures execution time, and aggregates errors.
//
// The method launches up to [maxParallel] concurrent senders to avoid overloading the SMTP server.
// Each goroutine reports its result into a channel [sent]. When all emails are processed,
// it collects all errors (if any) and returns a combined error message.
func (m *mailService) SendMails(ctx context.Context, mail model.Mail) error {
	start := time.Now()
	maxParallel := defaultMaxParallel // limit of simultaneous SMTP sends

	logger.Info("SendMails started",
		zap.Int("recipients_total", len(mail.To)),
		zap.String("subject", mail.Subject),
	)

	// Validate input
	if len(mail.To) == 0 {
		err := errors.New("no recipients provided")
		logger.Warn("SendMails validation failedMails", zap.Error(err))
		return fmt.Errorf("validation error: %w", err)
	}

	// concurrency controls
	semaphore := make(chan struct{}, maxParallel)
	statusChan := make(chan sender.EmailStatus, len(mail.To))
	wg := sync.WaitGroup{}

	select {
	case <-ctx.Done():
		err := ctx.Err()
		logger.Warn("SendMails aborted - context canceled", zap.Error(err))
		return fmt.Errorf("operation canceled: %w", err)
	default:
	}

	// Launch parallel sending
	for _, recipient := range mail.To {
		semaphore <- struct{}{}
		wg.Add(1)

		go func(recipient string) {
			defer func() {
				<-semaphore
				wg.Done()
			}()

			// check context in each worker
			select {
			case <-ctx.Done():
				statusChan <- sender.EmailStatus{
					Status:    sender.Error,
					StatusMsg: fmt.Sprintf("context canceled before sending to %s", recipient),
					Cause:     ctx.Err(),
				}
				return
			default:
			}

			// Prepare attachment
			attach, err := mail.GetAttachmentPath(recipient)
			// Create message
			msg, _ := sender.FormMessage(mail.Subject, mail.Body, attach, mail.From, recipient)
			if err != nil {
				statusChan <- sender.EmailStatus{
					Status:    sender.Error,
					StatusMsg: fmt.Sprintf("attachment path for %s is empty:", recipient),
					Cause:     err,
					Msg:       msg,
				}
				logger.Info("SendMails empty attachment",
					zap.String("recipient", recipient),
					zap.Error(err),
				)
				return
			}

			m.sender.SendEmail(msg, statusChan, true)
		}(recipient)
	}

	wg.Wait()
	close(statusChan)

	// Collect results
	failedMails := make(map[string]string)
	failedAttachments := make(map[string]string)
	for status := range statusChan {
		fmt.Println(status)
		if status.Status == sender.Error {
			logger.Error("SendMails failedMails to send email",
				zap.Any("recipients", status.Msg.GetHeader("To")),
				zap.String("status_msg", status.StatusMsg),
				zap.Error(status.Cause),
			)

			to := status.Msg.GetHeader("To")

			if len(to) > 0 {
				// remove sender email occurrences and pick first remaining
				for _, receiver := range to {
					if receiver == mail.From {
						continue
					}
					failedMails[receiver] = status.Cause.Error()
					attach, _ := mail.GetAttachmentPath(receiver)
					failedAttachments[receiver] = attach
				}
			}

		}
	}

	duration := time.Since(start)
	logger.Info("SendMails completed",
		zap.Int("recipients_total", len(mail.To)),
		zap.Int("failed_count", len(failedMails)),
		zap.Duration("elapsed", duration),
	)

	if len(failedMails) > 0 {
		return &EmailSendingError{MapReceiverCause: failedMails, AttachmentPaths: failedAttachments}
	}

	return nil
}

func (m *mailService) GetSenderEmail() string {
	return m.sender.GetSenderEmail()
}
