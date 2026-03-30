package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type URL struct {
	ID           string    `json:"id"`
	OriginalURL  string    `json:"original_url"`
	ShortURL     string    `json:"short_url"`
	CreationDate time.Time `json:"creation_date"`
}

var urlDB = make(map[string]URL) //this is our in-memory database to store the URL mappings

func RootPageURL(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!")
}

func ShortURLHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		URL string `json:"url"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	shortURL := createURL(data.URL)
	// fmt.Fprintf(w, shortURL)
	response := struct {
		ShortURL string `json:"short_url"`
	}{ShortURL: shortURL}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func redirectURLHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/redirect/"):]
	url, err := getURL(id)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}

/*
	d9736711 --> {
					ID: "d9736711",
					OriginalURL: "https://github.com/sujal12344/",
					ShortURL: "d9736711",
					CreationDate: time.Now()
				}
*/

func generateShortURL(originalUrl string) string {
	hasher := md5.New()
	hasher.Write([]byte(originalUrl))
	// fmt.Println("hasher: ", hasher)
	data := hasher.Sum(nil)
	// fmt.Println("data: ", string(data))
	hash := hex.EncodeToString(data)
	// fmt.Println("EncodeToString: ", hash)
	urlID := hash[:8] //taking the first 8 characters of the hash as the short URL ID

	//before storing the URL mapping, we should check if the generated short URL ID already exists in our database to avoid collisions. If it does, we can generate a new short URL ID by appending a random string or using a different hashing algorithm.
	if _, exists := urlDB[urlID]; exists {
		urlID = generateShortURL(originalUrl + time.Now().String()) //append current time to original URL to generate a new hash, for example "https://github.com/sujal12344/2023-10-01 12:00:00"
	}

	return urlID
}

func createURL(originalURL string) string {
	shortURL := generateShortURL(originalURL)
	id := shortURL // Use the short URL as the ID for simplicity
	urlDB[id] = URL{
		ID:           id,
		OriginalURL:  originalURL,
		ShortURL:     shortURL,
		CreationDate: time.Now(),
	}
	return shortURL
}

func getURL(id string) (URL, error) {
	url, ok := urlDB[id]
	if !ok {
		return URL{}, errors.New("URL not found")
	}

	if time.Since(url.CreationDate) > 30*24*time.Hour { // Check if the URL is older than 30 days
		delete(urlDB, id)
		return URL{}, errors.New("URL has expired")
	}

	if url.OriginalURL == "" {
		return URL{}, errors.New("Original URL is empty")
	}

	if url.ShortURL == "" {
		return URL{}, errors.New("Short URL is empty")
	}
	return url, nil
}

func main() {
	fmt.Println("This is URL shortener")
	originalUrl := "https://github.com/sujal12344/"
	shortUrl := generateShortURL(originalUrl)
	fmt.Println("Short URL: ", shortUrl)

	// Register the handler function to handle all requests to the root URL ("/")
	http.HandleFunc("/", RootPageURL)
	http.HandleFunc("/shorten", ShortURLHandler)
	http.HandleFunc("/redirect/", redirectURLHandler)

	// Start the HTTP server on port 3000
	fmt.Println("Starting server on port 3000...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Println("Error on starting server:", err)
	}
}
