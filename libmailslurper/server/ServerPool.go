// Copyright 2013-3014 Adam Presley. All rights reserved
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package server

import(
	"errors"
	"net"

	"github.com/adampresley/mailslurper/libmailslurper/model/mailitem"
	"github.com/adampresley/mailslurper/libmailslurper/smtpconstants"
	"github.com/adampresley/mailslurper/libmailslurper/smtpio"
)

/*
This structure represents a pool of SMTP workers. This will
manage how many workers may respond to SMTP client requests
and allocation of those workers.
*/
type ServerPool struct{
	SmtpWorkers []SmtpWorker
	MaxWorkers  int
}

/*
Create a new server pool with a maximum number of SMTP
workers. An array of workers is initialized with an ID
and an initial state of SMTP_WORKER_IDLE.
*/
func NewServerPool(maxWorkers int) *ServerPool {
	var workers [maxWorkers]SmtpWorker
	result := &ServerPool{MaxWorkers: maxWorkers,}

	for index := 0; index < maxWorkers; index++ {
		workers[index] = SmtpWorker{WorkerId: index, State: smtpconstants.SMTP_WORKER_IDLE,}
	}

	result.SmtpWorkers = workers
	return result
}

/*
Returns the next available SMTP worker from the pool, if
there is a worker available. If there is not a worker
available an error is returned. If a worker is returned
its state is set to SMTP_WORKER_WORKING.
*/
func (this *ServerPool) GetAvailableWorker(connection net.Conn, receiver chan mailitem.MailItem) (*SmtpWorker, error) {
	result := &SmtpWorker{}

	for index := 0; index < this.MaxWorkers; index++ {
		if this.SmtpWorkers[index].State == smtpconstants.SMTP_WORKER_IDLE {
			result = &SmtpWorkers[index]
			result.State = smtpconstants.SMTP_WORKER_WORKING

			result.Connection = connection
			result.Reader = smtpio.Reader{Connection: &result.Connection,}
			result.Receiver = receiver
			result.Writer = smtpio.Writer{Connection: &result.Connection,}
		}
	}

	if result.WorkerId == 0 {
		return result, errors.New("There are no available workers to service your request")
	}

	return result, nil
}
