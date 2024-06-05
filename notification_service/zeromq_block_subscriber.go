package notification_service

import (
	"fmt"
	"log"

	"gopkg.in/zeromq/goczmq.v4"
)

func ZeromqBlockSubscriber() {
	channeler := goczmq.NewSubChanneler("tcp://127.0.0.1:28332", "hashblock")
	// Print the channeler
	fmt.Println(channeler)

	if channeler == nil {
		// Print error
		log.Fatal("Error creating channeler", channeler)
	}
	defer channeler.Destroy()

	// Listen for messages
	fmt.Println("Listening for messages...")
	for {
		select {
		case msg, ok := <-channeler.RecvChan:
			if !ok {
				log.Fatal("Failed to receive message")
			}
			if len(msg) != 3 {
				log.Println("Received message with unexpected number of parts")
				continue
			}

			// Split the message into topic, serialized transaction, and sequence number
			topic := string(msg[0])
			serializedTx := msg[1]

			// Print out the parts
			fmt.Printf("Topic: %s\n", topic)
			fmt.Printf("Serialized Transaction: %x\n", serializedTx) // Print as hex
		}
	}
}
