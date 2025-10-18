package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type PingResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Timestamp int64  `json:"timestamp"`
}

type CompareRequest struct {
	DatabaseName string `json:"database_name"`
	ManifestData string `json:"manifest_data"`
}

type CompareResponse struct {
	HasDifferences bool     `json:"has_differences"`
	Differences    []string `json:"differences"`
}

type Server struct {
	port string
}

func (s *Server) pingHandler(w http.ResponseWriter, r *http.Request) {
	response := PingResponse{
		Status:    "ok",
		Version:   "1.0.0",
		Timestamp: time.Now().Unix(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) compareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req CompareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	response := CompareResponse{
		HasDifferences: false,
		Differences:    []string{},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", s.pingHandler)
	mux.HandleFunc("/compare", s.compareHandler)
	
	server := &http.Server{
		Addr:    ":" + s.port,
		Handler: mux,
	}
	
	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()
	
	log.Printf("Starting Database Agent on port %s", s.port)
	return server.ListenAndServe()
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}
	
	server := &Server{port: port}
	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
