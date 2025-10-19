package da

import (
	"encoding/hex"
	"log"
	"strings"
	"time"

	"github.com/Layer-Edge/bitcoin-da/clients"
	"github.com/Layer-Edge/bitcoin-da/config"
	"github.com/Layer-Edge/bitcoin-da/models"
)

func ProcessBTCMsg(msg []byte, protocolId string) ([]byte, error) {
	data := append([]byte(protocolId), msg...)
	hash := CreateOPReturnTransaction(hex.EncodeToString(data))
	return []byte(hash), nil
}

func SuperProofCronJob(cfg *config.Config) {
	// Initialize database with retry mechanism
	err := models.InitDB(cfg.PostgresConnectionURI)
	if err != nil {
		log.Fatalf("Error initializing DB Connection: %v", err)
	}
	defer func() {
		if err := models.CloseDB(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	InitOPReturnRPC(cfg.BtcEndpoint, cfg.Auth, cfg.WalletPassphrase)

	log.Println("Starting Super Proof Cron Job")
	log.Println("Super proof will run 6 times daily at 12:00 AM, 4:00 AM, 8:00 AM, 12:00 PM, 4:00 PM, and 8:00 PM UTC")

	// Define the scheduled hours (0, 4, 8, 12, 16, 20 in 24-hour format)
	scheduledHours := []int{0, 4, 8, 12, 16, 20}

	for {
		now := time.Now().UTC()

		// Find the next scheduled time
		nextScheduledTime := findNextScheduledTime(now, scheduledHours)

		// Calculate duration until next scheduled time
		duration := nextScheduledTime.Sub(now)
		log.Printf("Next super proof scheduled for: %s (in %v)", nextScheduledTime.Format("2006-01-02 15:04:05 UTC"), duration)

		// Wait until next scheduled time
		time.Sleep(duration)

		// Run the super proof process
		log.Printf("Running super proof at scheduled time: %s", time.Now().UTC().Format("2006-01-02 15:04:05 UTC"))
		processSuperProof(cfg)
	}
}

// findNextScheduledTime calculates the next scheduled time based on current time and scheduled hours
func findNextScheduledTime(now time.Time, scheduledHours []int) time.Time {
	currentHour := now.Hour()

	// Find the next hour in today's schedule
	for _, hour := range scheduledHours {
		if hour > currentHour {
			// Found a time later today
			return time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, time.UTC)
		}
	}

	// No more times today, schedule for the first time tomorrow
	nextDay := now.AddDate(0, 0, 1)
	return time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), scheduledHours[0], 0, 0, 0, time.UTC)
}

func processSuperProof(cfg *config.Config) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in processSuperProof: %v", r)
		}
	}()

	log.Println("Processing super proof...")

	// Get the last processed timestamp
	lastProcessedTimestamp, err := models.GetLastSuperProofTimestamp()
	if err != nil {
		log.Printf("Error getting last super proof timestamp: %v", err)
		return
	}

	// Fetch all unprocessed merkle roots
	merkleRoots, err := models.GetUnprocessedMerkleRoots(lastProcessedTimestamp)
	if err != nil {
		log.Printf("Error fetching unprocessed merkle roots: %v", err)
		return
	}

	if len(merkleRoots) == 0 {
		log.Println("No new merkle roots to process for super proof")
		return
	}

	log.Printf("Found %d merkle roots to process in super proof", len(merkleRoots))

	// Initialize data reader for BTC processing
	dataReader := NewBlockSubscriber()
	defer func() {
		if err := dataReader.Close(); err != nil {
			log.Printf("Error closing BlockSubscriber: %v", err)
		}
	}()

	// Generate super proof (merkle tree of all merkle roots)
	prf := ZKProof{}
	superMerkleRoot := prf.GenerateAggregatedProof(strings.Join(merkleRoots, ""))

	if superMerkleRoot == "" {
		log.Println("Failed to generate super proof, skipping write")
		return
	}

	log.Printf("Generated super proof: %s", superMerkleRoot)

	// Process BTC transaction for the super proof
	fnBtc := func(msg [][]byte) ([]byte, error) {
		hash, err := ProcessBTCMsg(msg[1], cfg.ProtocolId)
		return hash, err
	}

	hash, err := dataReader.ProcessOutTuple(fnBtc, [][]byte{nil, []byte(superMerkleRoot)})
	if err != nil {
		log.Printf("Error writing super proof to BTC: %v", err)
		return
	}

	btcTxHash := strings.ReplaceAll(string(hash[:]), "\n", "")
	log.Printf("Super proof BTC transaction hash: %s", btcTxHash)

	// Get transaction details including block number
	_, btcBlockNumber := GetTransactionInfo(btcTxHash)
	if btcBlockNumber != nil {
		log.Printf("Super proof BTC transaction confirmed in block: %d", *btcBlockNumber)
	} else {
		log.Printf("Super proof BTC transaction block information not available yet")
	}

	// Store super proof merkle tree on LayerEdge
	txData, err := clients.StoreMerkleTree(cfg, cfg.LayerEdgeRPC.SuperProofContract, superMerkleRoot, merkleRoots)
	if err != nil {
		log.Printf("Error storing super proof merkle tree: %v", err)
		// Continue with database storage even if contract call fails
	}

	// Store super proof in database
	if txData != nil {
		// Create a super proof entry with BTC information
		// Super proofs are distinguished by having BTCTxHash set (not null)
		aggProof, err := models.CreateAggregatedProofWithBTC(
			superMerkleRoot,
			merkleRoots,    // The individual merkle roots are the "proofs" for the super proof
			&btcTxHash,     // BTC transaction hash for super proof
			btcBlockNumber, // BTC block number
			*txData,
		)
		if err != nil {
			log.Printf("Failed to store super proof in DB: %v", err)
		} else {
			log.Printf("Stored super proof successfully: %v", aggProof)
		}
	} else {
		log.Println("No transaction data available for super proof, skipping database storage")
	}
}
