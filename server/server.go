package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	apiURL         = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	httpPort       = ":8080"
	dbFile         = "cotacoes.db"
	apiTimeout     = 200 * time.Millisecond
	dbTimeout      = 10 * time.Millisecond
	createTableSQL = `CREATE TABLE IF NOT EXISTS cotacoes (id INTEGER PRIMARY KEY AUTOINCREMENT, bid TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP);`
	insertSQL      = `INSERT INTO cotacoes (bid) VALUES (?)`
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	db := initDB()
	defer db.Close()

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		bid, err := getDollarExchangeRate(ctx)
		if err != nil {
			http.Error(w, "Error fetching exchange rate", http.StatusInternalServerError)
			log.Println("Error fetching exchange rate:", err)
			return
		}

		if err := saveToDatabase(ctx, db, bid); err != nil {
			http.Error(w, "Error saving to database", http.StatusInternalServerError)
			log.Println("Error saving to database:", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Cotacao{Bid: bid})
	})

	log.Println("Server is running on port", httpPort)
	log.Fatal(http.ListenAndServe(httpPort, nil))
}

func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func getDollarExchangeRate(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response code")
	}

	var result map[string]Cotacao
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result["USDBRL"].Bid, nil
}

func saveToDatabase(ctx context.Context, db *sql.DB, bid string) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	_, err := db.ExecContext(ctx, insertSQL, bid)
	return err
}
