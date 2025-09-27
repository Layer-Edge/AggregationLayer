package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Layer-Edge/bitcoin-da/config"
	"github.com/Layer-Edge/bitcoin-da/da"

	"github.com/Layer-Edge/bitcoin-da/utils"
)

var cfg = config.GetConfig()

func main() {
	// Initialize monitoring and error handling
	monitor := utils.InitializeMonitoring()
	defer monitor.Stop()

	// Set up error rate monitoring
	go utils.GetErrorHandler().MonitorErrorRate(10.0, 5*time.Minute)

	// Create a context that can be cancelled
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a channel for super proof messages
	superProofChan := make(chan [][]byte, 1000)

	// Create separate error channels for each service
	hashBlockDone := make(chan error, 1)
	superProofDone := make(chan error, 1)

	log.Println("Starting Bitcoin DA services...")
	utils.LogSystemError("main", "Services starting", nil, map[string]interface{}{
		"config": cfg,
	})

	// Start HashBlockSubscriber service
	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.RecoverFromPanic("HashBlockSubscriber")
				hashBlockDone <- fmt.Errorf("HashBlockSubscriber panic: %v", r)
			}
		}()

		log.Println("Starting HashBlockSubscriber...")
		da.HashBlockSubscriber(superProofChan, &cfg)
		hashBlockDone <- nil
	}()

	// Start SuperProofSubscriber service
	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.RecoverFromPanic("SuperProofSubscriber")
				superProofDone <- fmt.Errorf("SuperProofSubscriber panic: %v", r)
			}
		}()

		log.Println("Starting SuperProofSubscriber...")
		da.SuperProofSubscriber(superProofChan, &cfg)
		superProofDone <- nil
	}()

	// Wait for either shutdown signal or service completion
	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
		utils.LogSystemError("main", "Graceful shutdown initiated", nil, map[string]interface{}{
			"signal": sig.String(),
		})
		cancel()

		// Give the services time to shut down gracefully
		shutdownTimeout := time.NewTimer(30 * time.Second)
		defer shutdownTimeout.Stop()

		// Wait for both services to complete or timeout
		servicesShutdown := make(chan bool, 1)
		go func() {
			// Wait for both services to complete
			<-hashBlockDone
			<-superProofDone
			servicesShutdown <- true
		}()

		select {
		case <-servicesShutdown:
			log.Println("Services shut down gracefully")
			utils.LogSystemError("main", "Services shut down gracefully", nil, nil)
		case <-shutdownTimeout.C:
			log.Println("Service shutdown timeout reached, forcing exit")
			utils.LogCriticalError("main", "Service shutdown timeout", fmt.Errorf("shutdown timeout"), nil)
		}

	case err := <-hashBlockDone:
		if err != nil {
			utils.LogCriticalError("main", "HashBlockSubscriber failed", err, nil)
			log.Fatalf("HashBlockSubscriber failed: %v", err)
		}
		log.Println("HashBlockSubscriber completed normally")

	case err := <-superProofDone:
		if err != nil {
			utils.LogCriticalError("main", "SuperProofSubscriber failed", err, nil)
			log.Fatalf("SuperProofSubscriber failed: %v", err)
		}
		log.Println("SuperProofSubscriber completed normally")
	}
}
