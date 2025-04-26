package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ExchangeRate struct {
	Bid string `json:"bid"`
}

type ApiResponse struct {
	USDBRL ExchangeRate `json:"USDBRL"`
}

func main() {
	http.HandleFunc("/cotacao", handleExchangeRate)
	log.Println("Server running on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleExchangeRate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Context for external API call (200ms timeout)
	ctxAPI, cancelAPI := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancelAPI()

	req, err := http.NewRequestWithContext(ctxAPI, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		log.Println("Failed to create request to external API:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error calling external API:", err)
		http.Error(w, "Failed to retrieve exchange rate", http.StatusGatewayTimeout)
		return
	}
	defer resp.Body.Close()

	var apiResp ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Println("Failed to decode API response:", err)
		http.Error(w, "Failed to parse exchange rate", http.StatusInternalServerError)
		return
	}

	// Context for database write (10ms timeout)
	ctxDB, cancelDB := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancelDB()

	if err := saveExchangeRate(ctxDB, apiResp.USDBRL.Bid); err != nil {
		log.Println("Error saving exchange rate to database:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"bid": apiResp.USDBRL.Bid})
}

func saveExchangeRate(ctx context.Context, bid string) error {
	db, err := sql.Open("sqlite3", "./exchange_rates.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS exchange_rates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}

	stmt, err := db.PrepareContext(ctx, `INSERT INTO exchange_rates(bid) VALUES(?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, bid)
	return err
}
