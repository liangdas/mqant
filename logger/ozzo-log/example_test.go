// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package log_test

import "mqant/logger/ozzo-log"

func ExampleLogger_Error() {
	logger := log.NewLogger()

	logger.Targets = append(logger.Targets, log.NewConsoleTarget())

	logger.Open()

	// log without formatting
	logger.Error("a plain message")
	// log with formatting
	logger.Error("the value is: %v", 100)
}

func ExampleNewConsoleTarget() {
	logger := log.NewLogger()

	// creates a ConsoleTarget with color mode being disabled
	target := log.NewConsoleTarget()
	target.ColorMode = false

	logger.Targets = append(logger.Targets, target)

	logger.Open()

	// ... logger is ready to use ...
}

func ExampleNewNetworkTarget() {
	logger := log.NewLogger()

	// creates a NetworkTarget which uses tcp network and address :10234
	target := log.NewNetworkTarget()
	target.Network = "tcp"
	target.Address = ":10234"

	logger.Targets = append(logger.Targets, target)

	logger.Open()

	// ... logger is ready to use ...
}

func ExampleNewMailTarget() {
	logger := log.NewLogger()

	// creates a MailTarget which sends emails to admin@example.com
	target := log.NewMailTarget()
	target.Host = "smtp.example.com"
	target.Username = "foo"
	target.Password = "bar"
	target.Subject = "log messages for foobar"
	target.Sender = "admin@example.com"
	target.Recipients = []string{"admin@example.com"}

	logger.Targets = append(logger.Targets, target)

	logger.Open()

	// ... logger is ready to use ...
}

func ExampleNewFileTarget() {
	logger := log.NewLogger()

	// creates a FileTarget which keeps log messages in the app.log file
	target := log.NewFileTarget()
	target.FileName = "app.log"

	logger.Targets = append(logger.Targets, target)

	logger.Open()

	// ... logger is ready to use ...
}
