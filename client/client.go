package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	serverURL  = "http://localhost:8080/cotacao"
	timeout    = 300 * time.Millisecond
	outputFile = "cotacao.txt"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
	if err != nil {
		logErrorAndExit("Error creating request:", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logErrorAndExit("Error making request:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logErrorAndExit("Error: received non-200 response code", fmt.Errorf("status code: %d", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logErrorAndExit("Error reading response body:", err)
	}

	value := parseExchangeRate(body)
	saveToFile(value)
	fmt.Println("Cotação salva com sucesso:", value)
}

func parseExchangeRate(data []byte) string {
	type Cotacao struct {
		Bid string `json:"bid"`
	}

	var cotacao Cotacao
	if err := json.Unmarshal(data, &cotacao); err != nil {
		logErrorAndExit("Error parsing JSON:", err)
	}

	return cotacao.Bid
}

func saveToFile(value string) {
	content := fmt.Sprintf("Dólar: %s", value)
	if err := ioutil.WriteFile(outputFile, []byte(content), 0644); err != nil {
		logErrorAndExit("Error writing to file:", err)
	}
}

func logErrorAndExit(msg string, err error) {
	fmt.Println(msg, err)
	os.Exit(1)
}
