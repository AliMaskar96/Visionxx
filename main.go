package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const openaiEndpoint = "https://api.openai.com/v1/images/generations"

type OpenAIRequest struct {
	Model         string `json:"model"`
	Prompt        string `json:"prompt"`
	N             int    `json:"n"`
	Size          string `json:"size"`
	ResponseFormat string `json:"response_format"`
}

type OpenAIResponse struct {
	Data []struct {
		URL string `json:"url"`
	} `json:"data"`
}

// In-memory session store
var sessionStore = struct {
	sync.RWMutex
	store map[string]string
}{store: make(map[string]string)}

func generateSessionID() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%x", rand.Int63())
}

func generateDalleImage(prompt string) (string, error) {

	requestBody := OpenAIRequest{
		Model:         "dall-e-2",
		Prompt:        prompt,
		N:             1,
		Size:          "256x256",
		ResponseFormat: "url",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", openaiEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s", resp.Status)
	}

	var openAIResponse OpenAIResponse
	err = json.NewDecoder(resp.Body).Decode(&openAIResponse)
	if err != nil {
		return "", fmt.Errorf("failed to decode API response: %v", err)
	}

	if len(openAIResponse.Data) > 0 {
		return openAIResponse.Data[0].URL, nil
	}

	return "", fmt.Errorf("no image URL returned by the API")
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Retrieve or create session
	cookie, err := r.Cookie("session_id")
	if err != nil || cookie.Value == "" {
		sessionID := generateSessionID()
		http.SetCookie(w, &http.Cookie{
			Name:  "session_id",
			Value: sessionID,
			Path:  "/",
		})
		sessionStore.Lock()
		sessionStore.store[sessionID] = ""
		sessionStore.Unlock()
		fmt.Println("New session created:", sessionID)
	}

	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "./index.html")
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form data", http.StatusBadRequest)
			return
		}
		prompt := r.FormValue("prompt")

		// Generate image using DALL-E 2
		imageURL, err := generateDalleImage(prompt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"imagePath": imageURL})
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", handler)
	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
