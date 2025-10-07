package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Layer-Edge/bitcoin-da/config"
)

// CMCResponse models the small slice of the CMC response we need.
type CMCResponse struct {
	Data map[string]struct {
		Quote map[string]struct {
			Price float64 `json:"price"`
		} `json:"quote"`
	} `json:"data"`
}

func GetPrice(cfg *config.Config, symbol string) float64 {
	apiKey := cfg.CMCAPIKey

	symbols := []string{"ETH", "BTC", "EDGEN"}
	endpoint := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest"

	// Build URL with query
	q := url.Values{}
	q.Set("symbol", strings.Join(symbols, ","))

	req, err := http.NewRequest("GET", endpoint+"?"+q.Encode(), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		log.Fatalf("CMC API error: status=%d body=%s", res.StatusCode, string(b))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var cmc CMCResponse
	if err := json.Unmarshal(body, &cmc); err != nil {
		log.Fatal(err)
	}

	// Log the raw EDGEN object like console.log(response.data.data.EDGEN)
	if edgen, ok := cmc.Data["EDGEN"]; ok {
		j, _ := json.MarshalIndent(edgen, "", "  ")
		fmt.Printf("EDGEN object:\n%s\n", string(j))
	} else {
		fmt.Println("EDGEN object: <missing>")
	}

	// Extract prices (default to 0 if missing)
	return getUSDPrice(cmc, symbol)
}

func getUSDPrice(resp CMCResponse, symbol string) float64 {
	s, ok := resp.Data[symbol]
	if !ok {
		return 0
	}
	usd, ok := s.Quote["USD"]
	if !ok {
		return 0
	}
	return usd.Price
}
