package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"encoding/json"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

const (
	requestLimit = 5           // Maximum number of requests
	limitWindow  = time.Minute // Time window for the limit
	dbFile       = "keys.db"
)

var db *sql.DB

// Initialize the database and API keys table
func initDatabase() {
	var err error
	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	} else {
		fmt.Println("valjda otvorena i kreirana baza")
	}

	// Create the api_keys table if it doesn't exist
	query := `
		CREATE TABLE IF NOT EXISTS api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT UNIQUE NOT NULL
		);
	`
	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

// Insert an API key into the database
func insertAPIKey(key string) error {
	query := `INSERT INTO api_keys (key) VALUES (?);`
	_, err := db.Exec(query, key)
	return err
}

// Validate API key against the database
func validateAPIKey(key string) bool {
	query := `SELECT COUNT(*) FROM api_keys WHERE key = ?;`
	var count int
	err := db.QueryRow(query, key).Scan(&count)
	if err != nil {
		log.Printf("Error validating API key: %v", err)
		return false
	}
	return count > 0
}

type Response struct {
	Message string `json:"message"`
}

type RateLimiter struct {
	mu         sync.Mutex
	requests   map[string]int
	timestamps map[string]time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requests:   make(map[string]int),
		timestamps: make(map[string]time.Time),
	}
}

func (rl *RateLimiter) Allow(apiKey string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// If the timestamp is outside the window, reset the count and timestamp
	if ts, exists := rl.timestamps[apiKey]; !exists || now.Sub(ts) > limitWindow {
		rl.requests[apiKey] = 0
		rl.timestamps[apiKey] = now
	}

	// Check the current request count
	if rl.requests[apiKey] >= requestLimit {
		return false
	}

	// Increment the request count
	rl.requests[apiKey]++
	return true
}

var rateLimiter = NewRateLimiter()

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("request from base route")
	// Set the response header to JSON
	w.Header().Set("Content-Type", "application/json")

	// Create a response struct
	response := Response{Message: "Hello, base route!"}

	// Encode the response to JSON and write it to the response writer
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Unable to encode response", http.StatusInternalServerError)
	}
}

func apiroute(w http.ResponseWriter, r *http.Request) {
	// Check for API key in the query parameters
	key := r.URL.Query().Get("api_key")
	if !validateAPIKey(key) {
		http.Error(w, "Unauthorized: Invalid API key", http.StatusUnauthorized)
		return
	}

	// Enforce rate limiting
	if !rateLimiter.Allow(key) {
		http.Error(w, "Too Many Requests: Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Set the response header to plain text
	w.Header().Set("Content-Type", "text/plain")

	// Write the plain text response directly
	w.Write([]byte("Hello, world!"))
}

func name(w http.ResponseWriter, r *http.Request) {
	response := "Evo me\n"
	w.Write([]byte(response))
}

func main() {
	fmt.Println("REST API Server")
	initDatabase()
	// insertAPIKey("apikey123")
	// validateAPIKey("apikey123")
	// Register the handler endpoints
	http.HandleFunc("/", handler)
	http.HandleFunc("/name", name)
	http.HandleFunc("/apiroute", apiroute)

	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using default values")
	}

	// Get the port from the .env file, with a default of 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	// Start the server on the configured port
	address := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on %s", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
		panic(err)
	}
	defer db.Close()
}
