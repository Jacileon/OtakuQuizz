package database

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

type Supabase struct {
	URL    string
	APIKey string
	Client *http.Client
}

func Connect() (*Supabase, error) {
	url := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_ANON_KEY")

	if url == "" {
		return nil, fmt.Errorf("SUPABASE_URL is required")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("SUPABASE_ANON_KEY is required")
	}

	return &Supabase{
		URL:    url,
		APIKey: apiKey,
		Client: &http.Client{},
	}, nil
}

func (s *Supabase) Query(table string, query string) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s?%s", s.URL, table, query)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", s.APIKey)
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (s *Supabase) Insert(table string, data []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s", s.URL, table)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", s.APIKey)
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
