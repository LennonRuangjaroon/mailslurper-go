// Copyright 2013-2014 Adam Presley. All rights reserved
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package smtp

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/adampresley/mailslurper/admin/model"
	_ "github.com/mattn/go-sqlite3"
)

/*
Structure for holding a persistent database connection.
*/
type MailStorage struct {
	Db *sql.DB
}

// Global variable for our server's database connection
var Storage MailStorage

/*
Open a connection to a SQLite database. This will attempt to delete any
existing database file and create a new one with a blank table for holding
mail data.
*/
func (ms *MailStorage) Connect(filename string) error {
	os.Remove(filename)

	/*
	 * Create the connection
	 */
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return err
	}

	ms.Db = db

	/*
	 * Create the mailitem table.
	 */
	sql := `
		CREATE TABLE mailitem (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			dateSent TEXT,
			fromAddress TEXT,
			toAddressList TEXT,
			subject TEXT,
			xmailer TEXT,
			body TEXT,
			contentType TEXT,
			boundary TEXT
		);
	`

	_, err = ms.Db.Exec(sql)
	if err != nil {
		return err
	}

	sql = `
		CREATE TABLE attachment (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			mailItemId INTEGER,
			fileName TEXT,
			contentType TEXT,
			content TEXT
		);
	`

	_, err = ms.Db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

/*
Close a SQLite database connection.
*/
func (ms *MailStorage) Disconnect() {
	ms.Db.Close()
}

/*
Listens for messages on a channel for mail messages to be written
to disk. This channel takes in MailItemStruct mail items.
*/
func (ms *MailStorage) StartWriteListener(dbWriteChannel chan MailItemStruct) {
	for {
		mailItem := <-dbWriteChannel

		/*
		 * Create a transaction and insert the new mail item
		 */
		transaction, err := ms.Db.Begin()
		if err != nil {
			panic(fmt.Sprintf("Error starting insert transaction: %s", err))
		}

		/*
		 * Insert the mail item
		 */
		statement, err := transaction.Prepare("INSERT INTO mailitem (dateSent, fromAddress, toAddressList, subject, xmailer, body, contentType, boundary) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			panic(fmt.Sprintf("Error preparing insert statement: %s", err))
		}

		//defer statement.Close()

		result, err := statement.Exec(
			mailItem.DateSent,
			mailItem.FromAddress,
			strings.Join(mailItem.ToAddresses, "; "),
			mailItem.Subject,
			mailItem.XMailer,
			mailItem.Body,
			mailItem.ContentType,
			mailItem.Boundary,
		)

		if err != nil {
			panic(fmt.Sprintf("Error executing insert statement: %s", err))
		}

		statement.Close()
		mailItemId, _ := result.LastInsertId()
		mailItem.Id = int(mailItemId)

		/*
		 * Insert attachments
		 */
		for index := 0; index < len(mailItem.Attachments); index++ {
			statement, err = transaction.Prepare("INSERT INTO attachment (mailItemId, fileName, contentType, content) VALUES (?, ?, ?, ?)")
			if err != nil {
				panic(fmt.Sprintf("Error preparing insert attachment statement: %s", err))
			}

			_, err = statement.Exec(
				mailItemId,
				mailItem.Attachments[index].Headers.FileName,
				mailItem.Attachments[index].Headers.ContentType,
				mailItem.Attachments[index].Contents,
			)

			if err != nil {
				panic(fmt.Sprintf("Error executing insert attachment statement: %s", err))
			}

			statement.Close()
		}

		transaction.Commit()
		log.Printf("New mail item written to database.\n\n")

		BroadcastMessageToWebsockets(mailItem)
	}
}

/*
Retrieves all stored mail items as an array of MailItemStruct items.
*/
func (ms *MailStorage) GetMails() []model.JSONMailItem {
	result := make([]model.JSONMailItem, 0)

	rows, err := ms.Db.Query(`
		SELECT
			  mailItem.id AS mailItemId
			, mailitem.dateSent
			, mailitem.fromAddress
			, mailitem.toAddressList
			, mailitem.subject
			, mailitem.xmailer
			, attachment.id AS attachmentId
			, attachment.fileName
		FROM mailitem
			LEFT OUTER JOIN attachment ON mailitem.id=attachment.mailItemId
		ORDER BY mailitem.dateSent DESC
	`)

	if err != nil {
		panic("Error running query to get mail items")
	}

	defer rows.Close()

	currentMailItemId := 0
	attachments := make([]model.JSONAttachment, 0)
	newItem := model.JSONMailItem{}

	for rows.Next() {
		var mailItemId int
		var dateSent string
		var fromAddress string
		var toAddressList string
		var subject string
		var xmailer string
		var attachmentId int
		var fileName string

		rows.Scan(&mailItemId, &dateSent, &fromAddress, &toAddressList, &subject, &xmailer, &attachmentId, &fileName)

		/*
		 * If this is our first iteration then we haven't looked at a
		 * current item yet
		 */
		if currentMailItemId == 0 {
			currentMailItemId = mailItemId
		}

		/*
		 * There will be multiple records per mail item if there are
		 * multiple attachments. As such make sure we are getting all
		 * the IDs first, and the mail item only once.
		 */
		if currentMailItemId > 0 && currentMailItemId == mailItemId {
			if attachmentId > 0 {
				attachments = append(attachments, model.JSONAttachment{ Id: attachmentId, FileName: fileName, })
			}

			newItem = model.JSONMailItem{
				Id:              mailItemId,
				DateSent:        dateSent,
				FromAddress:     fromAddress,
				ToAddresses:     strings.Split(toAddressList, "; "),
				Subject:         subject,
				XMailer:         xmailer,
				Body:            "",
				ContentType:     "",
				AttachmentCount: 0,
				Attachments:     nil,
			}
		} else {
			newItem.Attachments = attachments
			newItem.AttachmentCount = len(attachments)

			result = append(result, newItem)
			attachments = make([]model.JSONAttachment, 0)

			if attachmentId > 0 {
				attachments = append(attachments, model.JSONAttachment{ Id: attachmentId, FileName: fileName, })
			}

			newItem = model.JSONMailItem{
				Id:              mailItemId,
				DateSent:        dateSent,
				FromAddress:     fromAddress,
				ToAddresses:     strings.Split(toAddressList, "; "),
				Subject:         subject,
				XMailer:         xmailer,
				Body:            "",
				ContentType:     "",
				AttachmentCount: 0,
				Attachments:     nil,
			}

			if currentMailItemId != mailItemId {
				currentMailItemId = mailItemId
			}

		}
	}

	newItem.Attachments = attachments
	newItem.AttachmentCount = len(attachments)
	result = append(result, newItem)

	return result
}

func (ms *MailStorage) GetAttachment(id int) map[string]string {
	rows, err := ms.Db.Query(`
		SELECT
			  fileName TEXT
			, contentType TEXT
			, content TEXT
		FROM attachment
		WHERE
			id=?
	`, id)

	if err != nil {
		panic("Error running query to get attachment")
	}

	defer rows.Close()

	result := make(map[string]string)

	for rows.Next() {
		var fileName string
		var contentType string
		var content string

		rows.Scan(&fileName, &contentType, &content)

		result["fileName"] = fileName
		result["contentType"] = contentType
		result["content"] = content
	}

	return result
}

/*
Retrieves a single mail item and its attachments.
*/
func (ms *MailStorage) GetMail(id int) model.JSONMailItem {
	rows, err := ms.Db.Query(`
		SELECT
			  mailItem.id AS mailItemId
			, mailitem.dateSent
			, mailitem.fromAddress
			, mailitem.toAddressList
			, mailitem.subject
			, mailitem.xmailer
			, mailitem.body
			, mailitem.contentType
			, attachment.id AS attachmentId
			, attachment.fileName
		FROM mailitem
			LEFT OUTER JOIN attachment ON mailitem.id=attachment.mailItemId
		WHERE mailItem.id=?
	`, id)

	if err != nil {
		panic("Error running query to get mail item")
	}

	defer rows.Close()

	result := model.JSONMailItem{}
	attachments := make([]model.JSONAttachment, 0)

	for rows.Next() {
		var mailItemId int
		var dateSent string
		var fromAddress string
		var toAddressList string
		var subject string
		var xmailer string
		var body string
		var contentType string
		var attachmentId int
		var fileName string

		rows.Scan(&mailItemId, &dateSent, &fromAddress, &toAddressList, &subject, &xmailer, &body, &contentType, &attachmentId, &fileName)

		if attachmentId > 0 {
			attachments = append(attachments, model.JSONAttachment{ Id: attachmentId, FileName: fileName })
		}

		result = model.JSONMailItem{
			Id:              mailItemId,
			DateSent:        dateSent,
			FromAddress:     fromAddress,
			ToAddresses:     strings.Split(toAddressList, "; "),
			Subject:         subject,
			XMailer:         xmailer,
			Body:            body,
			ContentType:     contentType,
			AttachmentCount: len(attachments),
			Attachments:     attachments,
		}
	}

	return result
}