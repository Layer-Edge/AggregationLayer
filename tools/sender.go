package main

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/zeromq/goczmq.v4"
)

func main() {
	ep := "tcp://0.0.0.0:40006"
	sender := goczmq.NewReqChanneler(ep)
	if sender == nil {
		log.Fatal("Failed to subscribe to endpoint: ", ep)
	}
	defer sender.Destroy()
	// Let the socket connect
	time.Sleep(5 * time.Second)
	// Send some data
	data := [][]byte{[]byte("datablock"), []byte("Test Data " + time.Now().String()), []byte("!!!!!")}
	sender.SendChan <- data
	fmt.Printf("Data sent %s\n", data)
	resp := <-sender.RecvChan
	fmt.Printf("Response received %s\n", resp)
}
