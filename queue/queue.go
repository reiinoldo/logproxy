package queue

import (
	"github.com/philips-software/go-hsdp-api/logging"
)

// Queue implements a logproxy queuer
// Internally the queue should transform the raw payload to a logging.Resource
type Queue interface {
	// Start initializes the and returns a stop channel
	Start() (chan bool, error)
	// Output should return a channel fed by the queue raw data
	Output() <-chan logging.Resource
	// Push should queue the raw payload
	Push([]byte) error
	// DeadLetter should store a rejected logging.Resource for later processing
	DeadLetter(msg logging.Resource) error
}
