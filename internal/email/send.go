package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
)

const (
	MIN_TIME_BETWEEN_EMAILS_MSEC = 100
	LOCK_KEY                     = "internal.email"
)

var (
	// A reusable SMTP connection.
	// Should only be used of the lock is acquired.
	smtpConnection *smtp.Client = nil

	// The last time an email was sent.
	// Used the throttle the number of emails sent.
	lastEmailTime timestamp.Timestamp = timestamp.Zero()

	// Messages that are stored (instead of sent) in testing mode.
	testMessages []*Message = nil

	// A timer to regularly check for idle connections.
	ticker *time.Ticker = nil

	// True if the email systems has been initialized.
	// We delay initialization so that config values can be set.
	isInitialized bool = false
)

func GetTestMessages() []*Message {
	return testMessages
}

func ClearTestMessages() {
	testMessages = nil
}

func Send(to []string, subject string, body string, html bool) error {
	return SendMessage(&Message{
		To:      to,
		Subject: subject,
		Body:    body,
		HTML:    html,
	})
}

// Send a message.
// Instead of using the simpler smtp.SendMail(),
// we will be dialing a SMTP connection to be reused over multiple messages.
// This prevents some email providers (like gmail) from complaining about too many SMTP connections.
func SendMessage(message *Message) error {
	// In testing mode, just store the message.
	if config.TESTING_MODE.Get() {
		testMessages = append(testMessages, message)
		return nil
	}

	// Before acquiring the email lock, initialize the email system.
	initialize()

	// Only send one email at a time.
	common.Lock(LOCK_KEY)
	defer common.Unlock(LOCK_KEY)

	content := message.ToContent()

	// Sleep if we are sending emails too fast.
	timeSinceLastEmail := timestamp.Now() - lastEmailTime
	sleepDuration := timestamp.FromMSecs(MIN_TIME_BETWEEN_EMAILS_MSEC) - timeSinceLastEmail
	if sleepDuration > 0 {
		time.Sleep(sleepDuration.ToGoTimeDuration())
	}

	lastEmailTime = timestamp.Now()

	return sendEmail(message.To, content)
}

// Send the message.
// Should only be called with the email lock.
func sendEmail(to []string, content []byte) error {
	err := ensureConnection()
	if err != nil {
		return err
	}

	err = smtpConnection.Mail(config.EMAIL_FROM.Get())
	if err != nil {
		return fmt.Errorf("Failed to issue SMTP MAIL command: '%w'.", err)
	}

	for _, address := range to {
		err = smtpConnection.Rcpt(address)
		if err != nil {
			return fmt.Errorf("Failed to issue SMTP RCPT command to '%s': '%w'.", address, err)
		}
	}

	writter, err := smtpConnection.Data()
	if err != nil {
		return fmt.Errorf("Failed to issue SMTP DATA command: '%w'.", err)
	}
	defer writter.Close()

	_, err = writter.Write(content)
	if err != nil {
		return fmt.Errorf("Failed to write message content: '%w'.", err)
	}

	return nil
}

// Ensure that |smtpConnection| is valid, or return an error.
// Should only be called with the email lock.
func ensureConnection() error {
	if smtpConnection != nil {
		return nil
	}

	host := config.EMAIL_HOST.Get()
	serverAddress := fmt.Sprintf("%s:%s", host, config.EMAIL_PORT.Get())

	tempConnection, err := smtp.Dial(serverAddress)
	if err != nil {
		return fmt.Errorf("Could not dial SMTP server '%s': '%w'.", serverAddress, err)
	}

	username := config.EMAIL_USER.Get()
	authInfo := smtp.PlainAuth("", username, config.EMAIL_PASS.Get(), config.EMAIL_HOST.Get())

	// Encrypt if allowed.
	ok, _ := tempConnection.Extension("STARTTLS")
	if ok {
		config := &tls.Config{
			ServerName: host,
		}

		err = tempConnection.StartTLS(config)
		if err != nil {
			return fmt.Errorf("Failed to issue SMTP STARTTLS comment to server '%s': '%w'.", serverAddress, err)
		}
	} else {
		log.Warn("SMTP server '%s' does not support encryption.", serverAddress)
	}

	// Auth() will close the connection on error.
	err = tempConnection.Auth(authInfo)
	if err != nil {
		return fmt.Errorf("Could not auth user '%s' against SMTP server '%s': '%w'.", username, serverAddress, err)
	}

	// The connection is ready to use.
	smtpConnection = tempConnection

	return nil
}

// Initialize the email system.
// May be called many times (only the first one will happen).
// Will acquire the email lock.
// Regularly check if the connections should be closed.
// Note that this is not a Go-level initialization.
func initialize() {
	common.Lock(LOCK_KEY)
	defer common.Unlock(LOCK_KEY)

	// Already initialized.
	if isInitialized {
		return
	}

	ticker = time.NewTicker(time.Duration(config.EMAIL_SMTP_IDLE_TIMEOUT_MS.Get()) * time.Millisecond)
	go func() {
		for range ticker.C {
			err := Close()
			if err != nil {
				log.Warn("Failed to close SMTP connection.", err)
			}
		}
	}()

	isInitialized = true
}

// Close the SMTP connection.
// Will acquire the email lock.
func Close() error {
	common.Lock(LOCK_KEY)
	defer common.Unlock(LOCK_KEY)

	// Already closed.
	if smtpConnection == nil {
		return nil
	}

	// Quit() will contact the server and then close the connection.
	err := smtpConnection.Quit()

	// Once Quit() has been called, we will not use the connection even if there was an error.
	smtpConnection = nil

	if err != nil {
		return fmt.Errorf("Failed to issue SMTP QUIT command: '%w'.", err)
	}

	return nil
}
