// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package log

import (
	"errors"
	"fmt"
	"io"
	"net/smtp"
	"strings"
)

// MailTarget sends log messages in emails via an SMTP server.
type MailTarget struct {
	*Filter
	Host       string   // SMTP server address
	Username   string   // SMTP server login username
	Password   string   // SMTP server login password
	Subject    string   // the mail subject
	Sender     string   // the mail sender
	Recipients []string // the mail recipients
	BufferSize int      // the size of the message channel.

	entries chan *Entry
	close   chan bool
}

// NewMailTarget creates a MailTarget.
// The new MailTarget takes these default options:
// MaxLevel: LevelDebug, BufferSize: 1024.
// You must specify these fields: Host, Username, Subject, Sender, and Recipients.
func NewMailTarget() *MailTarget {
	return &MailTarget{
		Filter:     &Filter{MaxLevel: LevelDebug},
		BufferSize: 1024,
		close:      make(chan bool, 0),
	}
}

// Open prepares MailTarget for processing log messages.
func (t *MailTarget) Open(errWriter io.Writer) error {
	t.Filter.Init()
	if t.Host == "" {
		return errors.New("MailTarget.Host must be specified")
	}
	if t.Username == "" {
		return errors.New("MailTarget.Username must be specified")
	}
	if t.Subject == "" {
		return errors.New("MailTarget.Subject must be specified")
	}
	if t.Sender == "" {
		return errors.New("MailTarget.Sender must be specified")
	}
	if len(t.Recipients) == 0 {
		return errors.New("MailTarget.Recipients must be specified")
	}
	if t.BufferSize < 0 {
		return errors.New("MailTarget.BufferSize must be no less than 0")
	}
	t.entries = make(chan *Entry, t.BufferSize)

	go t.sendMessages(errWriter)

	return nil
}

// Process puts filtered log messages into a channel for sending in emails.
func (t *MailTarget) Process(e *Entry) {
	if t.Allow(e) {
		select {
		case t.entries <- e:
		default:
		}
	}
}

// Close closes the mail target.
func (t *MailTarget) Close() {
	<-t.close
}

func (t *MailTarget) sendMessages(errWriter io.Writer) {
	auth := smtp.PlainAuth(
		"",
		t.Username,
		t.Password,
		strings.Split(t.Host, ":")[0],
	)
	for {
		entry := <-t.entries
		if entry == nil {
			t.close <- true
			break
		}
		if err := t.write(auth, entry.String()+"\n"); err != nil {
			fmt.Fprintf(errWriter, "MailTarget write error: %v\n", err)
		}
	}
}

func (t *MailTarget) write(auth smtp.Auth, message string) error {
	msg := fmt.Sprintf("To: %v\r\nFrom: %v\r\nSubject: %v\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%v",
		strings.Join(t.Recipients, ";"),
		t.Sender,
		t.Subject,
		message,
	)
	return smtp.SendMail(t.Host, auth, t.Sender, t.Recipients, []byte(msg))
}
