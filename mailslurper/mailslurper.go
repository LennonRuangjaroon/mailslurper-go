// Copyright 2013-3014 Adam Presley. All rights reserved
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// LOOK ITO USING https://github.com/GeertJohan/go.rice
// TO EMBED ASSETS IN A SINGLE EXE!!!
package main

import (
	"log"
	"os"
	"runtime"

	"github.com/adampresley/mailslurper/libmailslurper/configuration"
	"github.com/adampresley/mailslurper/libmailslurper/model/mailitem"
	"github.com/adampresley/mailslurper/libmailslurper/server"
	"github.com/adampresley/mailslurper/libmailslurper/storage"
	"github.com/adampresley/mailslurper/mailslurperservice/listener"

	"github.com/adampresley/sigint"
)

func main() {
	var err error
	runtime.GOMAXPROCS(runtime.NumCPU())

	/*
	 * Prepare SIGINT handler (CTRL+C)
	 */
	sigint.ListenForSIGINT(func() {
		log.Println("Shutting down...")
		os.Exit(0)
	})

	/*
	 * Load configuration
	 */
	config, err := configuration.LoadConfigurationFromFile(configuration.CONFIGURATION_FILE_NAME)
	if err != nil {
		log.Println("ERROR - There was an error reading your configuration file: ", err)
		os.Exit(0)
	}

	/*
	 * Setup global database connection handle
	 */
	databaseConnection := config.GetDatabaseConfiguration()
	err = storage.ConnectToStorage(databaseConnection)
	if err != nil {
		log.Println("ERROR - There was an error connecting to your data storage: ", err)
		os.Exit(0)
	}

	defer storage.DisconnectFromStorage()

	/*
	 * Setup the server pool
	 */
	log.Println("INFO - Worker pool configured for", config.MaxWorkers, "workers")
	pool := server.NewServerPool(config.MaxWorkers)

	/*
	 * Setup the SMTP listener
	 */
	log.Println("INFO - Setting up SMTP listener on", config.GetFullSmtpBindingAddress())

	smtpServer, err := server.SetupSmtpServerListener(config.GetFullSmtpBindingAddress())
	if err != nil {
		log.Println("ERROR - There was a problem starting the SMTP listener: ", err)
		os.Exit(0)
	}

	defer server.CloseSmtpServerListener(smtpServer)

	receiver := make(chan mailitem.MailItem, 1000)
	go server.Dispatcher(pool, smtpServer, receiver)

	/*
	 * Start the services server
	 */
	log.Println("INFO - MailSlurper starting on", config.GetFullServiceAppAddress())
	listener.StartHttpListener(listener.NewHttpListener(config.ServiceAddress, config.ServicePort))
}

