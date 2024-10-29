package main

import (
    "log"
    "fmt"
    "gopkg.in/zeromq/goczmq.v4"
    "time"
)

func main() {
    ep := "tcp://0.0.0.0:19000"
    sender := goczmq.NewPubChanneler(ep)
    if sender == nil {
        log.Println("Failed to subscribe to endpoint: ", ep)
        return
    }
    time.Sleep(5 * time.Second)
    data := [][]byte{[]byte("datablock"), []byte("Hello world!! " + time.Now().String()), []byte("!!!!!")}
    sender.SendChan <- data
    time.Sleep(5 * time.Second)
    fmt.Printf("Data sent %s\n", data)
}
