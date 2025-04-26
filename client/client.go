package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatal("Failed to create request to server:", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Failed to contact server:", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Failed to read response body:", err)
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatal("Failed to parse JSON response:", err)
	}

	value := result["bid"]
	content := fmt.Sprintf("Dollar: %s\n", value)

	err = os.WriteFile("cotacao.txt", []byte(content), 0644)
	if err != nil {
		log.Fatal("Failed to write exchange rate to file:", err)
	}

	fmt.Println("Exchange rate saved to cotacao.txt:", value)
}
