package da

import (
	// "encoding/hex"
	"log"

	"gopkg.in/zeromq/goczmq.v4"
)

type Lambda func([][]byte) bool
type Lambda2 func([][]byte) ([]byte, error)

type BlockSubscriber struct {
	channeler *goczmq.Channeler
}

func (subr *BlockSubscriber) Subscribe(endpoint string, filter string) bool {
	log.Println("Subscibe:", endpoint, filter)
	subr.channeler = goczmq.NewSubChanneler(endpoint, filter)
	if subr.channeler == nil {
		log.Fatal("Error creating subscribe channeler: ", endpoint, filter)
		return false
	}
	return true
}

func (subr *BlockSubscriber) Replier(endpoint string) bool {
	log.Println("Replier:", endpoint)
	subr.channeler = goczmq.NewRepChanneler(endpoint)
	if subr.channeler == nil {
		log.Fatal("Error creating reply ehanneler: ", endpoint)
		return false
	}
	return true
}

func (subr *BlockSubscriber) Reset() {
	if subr.channeler != nil {
		subr.channeler.Destroy()
	}
}

func (subr *BlockSubscriber) GetMessage() (bool, [][]byte) {
	msg, ok := <-subr.channeler.RecvChan
	return ok, msg
}

func (subr *BlockSubscriber) Validate(ok bool, msg [][]byte) bool {
	// Validate
	if !ok {
		log.Println("Failed to receive message")
		return false
	}
	if len(msg) != 3 {
		log.Println("Received message with unexpected number of parts")
		return false
	}
	return true
}

func (subr *BlockSubscriber) Process(fn Lambda, msg [][]byte) bool {
	log.Println("Processing message")
	return fn(msg)
}

func (subr *BlockSubscriber) ProcessOutTuple(fn Lambda2, msg [][]byte) ([]byte, error) {
	log.Println("Processing message")
	return fn(msg)
}
