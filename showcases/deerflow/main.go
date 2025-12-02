package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Check for API key
	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Fatal("Please set OPENAI_API_KEY environment variable")
	}
	// Check for API Base if using DeepSeek (optional but recommended for non-OpenAI)
	if os.Getenv("OPENAI_API_BASE") == "" {
		fmt.Println("Warning: OPENAI_API_BASE not set. Defaulting to OpenAI. If using DeepSeek, set this to their API URL.")
	}

	// If arguments are provided, run in CLI mode
	if len(os.Args) > 1 {
		runCLI(os.Args[1])
		return
	}

	// Otherwise, run in Web Server mode
	runServer()
}

func runCLI(query string) {
	fmt.Printf("Starting Deer-Flow Research Agent with query: %s\n", query)

	graph, err := NewGraph()
	if err != nil {
		log.Fatalf("Failed to create graph: %v", err)
	}

	initialState := &State{
		Request: Request{
			Query: query,
		},
	}

	result, err := graph.Invoke(context.Background(), initialState)
	if err != nil {
		log.Fatalf("Graph execution failed: %v", err)
	}

	finalState := result.(*State)
	fmt.Println("\n=== Final Report ===")
	fmt.Println(finalState.FinalReport)
}

func runServer() {
	fs := http.FileServer(http.Dir("showcases/deerflow/web"))
	http.Handle("/", fs)

	http.HandleFunc("/api/run", handleRun)

	fmt.Println("🚀 DeerFlow Web Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Refactored handleRun to support concurrent logging and result retrieval
func handleRun(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	sendSSE(w, flusher, "update", map[string]string{"step": "Initializing..."})

	g, err := NewGraph()
	if err != nil {
		sendSSE(w, flusher, "error", map[string]string{"message": err.Error()})
		return
	}

	initialState := &State{
		Request: Request{
			Query: query,
		},
	}

	logChan := make(chan string, 100)
	resultChan := make(chan *State, 1)
	errChan := make(chan error, 1)

	ctx := context.WithValue(context.Background(), logKey{}, logChan)

	go func() {
		defer close(logChan)
		defer close(resultChan)
		defer close(errChan)

		res, err := g.Invoke(ctx, initialState)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- res.(*State)
	}()

	// Loop to handle logs and result
	for {
		select {
		case msg, ok := <-logChan:
			if !ok {
				logChan = nil // Channel closed
			} else {
				sendSSE(w, flusher, "log", map[string]string{"message": msg})
			}
		case res, ok := <-resultChan:
			if !ok {
				resultChan = nil
			} else {
				sendSSE(w, flusher, "result", map[string]string{"report": res.FinalReport})
				return // Done
			}
		case err, ok := <-errChan:
			if !ok {
				errChan = nil
			} else {
				sendSSE(w, flusher, "error", map[string]string{"message": err.Error()})
				return
			}
		}

		if logChan == nil && resultChan == nil && errChan == nil {
			break
		}
	}
}

func sendSSE(w http.ResponseWriter, flusher http.Flusher, eventType string, data interface{}) {
	payload := map[string]interface{}{
		"type": eventType,
	}

	// Merge data into payload
	if m, ok := data.(map[string]string); ok {
		for k, v := range m {
			payload[k] = v
		}
	}

	jsonPayload, _ := json.Marshal(payload)
	fmt.Fprintf(w, "data: %s\n\n", jsonPayload)
	flusher.Flush()
}
