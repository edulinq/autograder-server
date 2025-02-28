package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/lockmanager"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
)

const (
	LOCK_KEY = "internal.email"
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
	return SendFull(to, nil, nil, subject, body, html)
}

func SendFull(to []string, cc []string, bcc []string, subject string, body string, html bool) error {
	return SendMessage(&Message{
		To:      to,
		CC:      cc,
		BCC:     bcc,
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
	if config.UNIT_TESTING_MODE.Get() {
		testMessages = append(testMessages, message)
		return nil
	}

	// Only send one email at a time.
	lockmanager.Lock(LOCK_KEY)
	defer lockmanager.Unlock(LOCK_KEY)

	// Sleep if we are sending emails too fast.
	timeSinceLastEmail := timestamp.Now() - lastEmailTime
	sleepDuration := timestamp.FromMSecs(int64(config.EMAIL_MIN_PERIOD.Get())) - timeSinceLastEmail
	if sleepDuration > 0 {
		time.Sleep(sleepDuration.ToGoTimeDuration())
	}

	lastEmailTime = timestamp.Now()

	err := sendMessageInternal(message)
	if err != nil {
		log.Warn("Failed to send email.", message)
		return err
	}

	log.Trace("Sent email.", message)
	return nil
}

// Send the message.
// Should only be called with the email lock.
func sendMessageInternal(message *Message) error {
	err := ensureConnection()
	if err != nil {
		return err
	}

	err = smtpConnection.Mail(config.EMAIL_FROM.Get())
	if err != nil {
		return fmt.Errorf("Failed to issue SMTP MAIL command: '%w'.", err)
	}

	allRCPTs := append(message.To, append(message.CC, message.BCC...)...)
	for _, address := range allRCPTs {
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

	_, err = writter.Write(message.ToContent())
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
			tempConnection.Close()
			return fmt.Errorf("Failed to issue SMTP STARTTLS comment to server '%s': '%w'.", serverAddress, err)
		}
	} else {
		log.Warn("SMTP server '%s' does not support encryption.", serverAddress)
	}

	// Auth() will close the connection on error.
	err = tempConnection.Auth(authInfo)
	if err != nil {
		tempConnection.Close()
		return fmt.Errorf("Could not auth user '%s' against SMTP server '%s': '%w'.", username, serverAddress, err)
	}

	// The connection is ready to use.
	smtpConnection = tempConnection

	// Ensure we close the connection when it is idle.
	go closeIdleConnection()

	return nil
}

// Continually check for a connection that needs to be closed.
func closeIdleConnection() {
	done := false

	for {
		func() {
			lockmanager.Lock(LOCK_KEY)
			defer lockmanager.Unlock(LOCK_KEY)

			// The email system was closed.
			if smtpConnection == nil {
				done = true
				return
			}

			// Check for idle timeout.
			timeDelta := timestamp.Now() - lastEmailTime
			if timeDelta.ToMSecs() >= int64(config.EMAIL_SMTP_IDLE_TIMEOUT_MS.Get()) {
				err := closeWithLock()
				if err != nil {
					log.Warn("Failed to close SMTP connection.", err)
				}
				done = true
				return
			}
		}()

		if done {
			break
		}

		// If we are not done, schedule the next check.
		time.Sleep(time.Duration(config.EMAIL_SMTP_IDLE_TIMEOUT_MS.Get()) * time.Millisecond)
	}

}

// Close the SMTP connection.
// Will acquire the email lock.
func Close() error {
	lockmanager.Lock(LOCK_KEY)
	defer lockmanager.Unlock(LOCK_KEY)

	return closeWithLock()
}

func closeWithLock() error {
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
