// Copyright 2013 Adam Presley. All rights reserved
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package data

/*
MailItemStruct is a struct describing a parsed mail item. This is
populated after an incoming client connection has finished
sending mail data to this server.
*/
type MailItemStruct struct {
	DateSent    string   `json:"dateSent"`
	FromAddress string   `json:"fromAddress"`
	ToAddresses []string `json:"toAddresses"`
	Subject     string   `json:"subject"`
	XMailer     string   `json:"xmailer"`
	Body        string   `json:"body"`
}
