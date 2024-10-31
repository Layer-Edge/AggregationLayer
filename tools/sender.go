package main

import (
    "log"
    "fmt"
    "gopkg.in/zeromq/goczmq.v4"
    "time"
)

func main() {
	ep := "tcp://34.71.52.251:40000"
    sender := goczmq.NewReqChanneler(ep)
    if sender == nil {
        log.Fatal("Failed to subscribe to endpoint: ", ep)
    }
    defer sender.Destroy()
    // Let the socket connect
    time.Sleep(5 * time.Second)
    // Send some data
    data := [][]byte{[]byte("datablock"), []byte("Hello world!! " + time.Now().String()), []byte("!!!!!")}
    sender.SendChan <- data
    fmt.Printf("Data sent %s\n", data)
    resp := <- sender.RecvChan
    fmt.Printf("Response received %s\n", resp)
}
