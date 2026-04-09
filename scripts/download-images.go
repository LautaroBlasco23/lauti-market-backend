package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

const (
	apiURL       = "https://api.unsplash.com/photos/random"
	defaultQuery = "product"
	defaultCount = 20
	maxPerReq    = 30 // Unsplash API max per request
)

type unsplashPhoto struct {
	ID   string `json:"id"`
	Urls struct {
		Regular string `json:"regular"`
	} `json:"urls"`
}

func main() { //nolint:gocyclo
	accessKey := os.Getenv("UNSPLASH_ACCESS_KEY")
	if accessKey == "" {
		fmt.Fprintln(os.Stderr, "Error: UNSPLASH_ACCESS_KEY env var is required.")
		fmt.Fprintln(os.Stderr, "Get one at https://unsplash.com/developers")
		os.Exit(1)
	}

	count := defaultCount
	query := defaultQuery

	for _, arg := range os.Args[1:] {
		switch {
		case len(arg) > 8 && arg[:8] == "--count=":
			n, err := strconv.Atoi(arg[8:])
			if err != nil || n < 1 {
				fmt.Fprintln(os.Stderr, "Error: --count must be a positive integer")
				os.Exit(1)
			}
			count = n
		case len(arg) > 8 && arg[:8] == "--query=":
			query = arg[8:]
		case arg == "--help":
			fmt.Println("Usage: go run download-images.go [OPTIONS]")
			fmt.Println()
			fmt.Println("Downloads random images from Unsplash into scripts/fake-data-images/")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  --count=N    Number of images to download (default: 20, max 30)")
			fmt.Println("  --query=STR  Search query (default: product)")
			fmt.Println("  --help       Print this help and exit")
			fmt.Println()
			fmt.Println("Env:")
			fmt.Println("  UNSPLASH_ACCESS_KEY  (required) API key from unsplash.com/developers")
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown argument: %s\n", arg)
			os.Exit(1)
		}
	}

	if count > maxPerReq {
		fmt.Fprintf(os.Stderr, "Warning: Unsplash allows max %d per request, clamping.\n", maxPerReq)
		count = maxPerReq
	}

	// Resolve save directory relative to this script's location
	_, thisFile, _, _ := runtime.Caller(0)
	scriptDir := filepath.Dir(thisFile)
	saveDir := filepath.Join(scriptDir, "fake-data-images")

	if err := os.MkdirAll(saveDir, 0o750); err != nil { //nolint:gosec
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Fetch random photos
	url := fmt.Sprintf("%s?query=%s&count=%d&client_id=%s", apiURL, query, count, accessKey)
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching from Unsplash: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck
		_ = resp.Body.Close()            //nolint:errcheck
		fmt.Fprintf(os.Stderr, "Unsplash API error (HTTP %d): %s\n", resp.StatusCode, string(body))
		if resp.StatusCode == 403 || resp.StatusCode == 429 {
			fmt.Fprintln(os.Stderr, "Hint: free tier allows 50 requests/hour.")
		}
		os.Exit(1)
	}
	defer func() {
		_ = resp.Body.Close() //nolint:errcheck
	}()

	var photos []unsplashPhoto
	if err := json.NewDecoder(resp.Body).Decode(&photos); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1) //nolint:gocritic
	}

	fmt.Printf("Fetched %d photo(s), downloading...\n", len(photos))

	saved := 0
	for i, p := range photos {
		filePath := filepath.Join(saveDir, p.ID+".jpg")

		if err := downloadFile(p.Urls.Regular, filePath); err != nil {
			fmt.Fprintf(os.Stderr, "  [%d/%d] Failed %s: %v\n", i+1, len(photos), p.ID, err)
			continue
		}

		fmt.Printf("  [%d/%d] Saved %s\n", i+1, len(photos), filePath)
		saved++
	}

	fmt.Printf("\nDone: %d/%d images saved to %s\n", saved, len(photos), saveDir)
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close() //nolint:errcheck
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close() //nolint:errcheck
	}()

	_, err = io.Copy(f, resp.Body)
	return err
}
