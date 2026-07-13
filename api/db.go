package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Supabase struct {
	URL        string
	AnonKey    string
	ServiceKey string
	Client     *http.Client
}

func Connect() *Supabase {
	url := getEnv("SUPABASE_URL", "NEXT_PUBLIC_SUPABASE_URL")
	anonKey := getEnv("SUPABASE_ANON_KEY", "NEXT_PUBLIC_SUPABASE_ANON_KEY")
	serviceKey := getEnv("SUPABASE_SERVICE_ROLE_KEY", "")
	return &Supabase{URL: url, AnonKey: anonKey, ServiceKey: serviceKey, Client: &http.Client{}}
}

func getEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func (s *Supabase) request(method, path string, body []byte, useService bool) (*http.Response, error) {
	u := s.URL + path
	req, err := http.NewRequest(method, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	key := s.AnonKey
	if useService && s.ServiceKey != "" {
		key = s.ServiceKey
	}
	req.Header.Set("apikey", key)
	req.Header.Set("Content-Type", "application/json")
	if useService && s.ServiceKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.ServiceKey)
	}
	if method == "POST" || method == "PATCH" {
		req.Header.Set("Prefer", "return=representation")
	}
	return s.Client.Do(req)
}

func (s *Supabase) Select(table, query string, useService bool) ([]byte, error) {
	path := fmt.Sprintf("/rest/v1/%s?%s", table, query)
	resp, err := s.request("GET", path, nil, useService)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("select %s failed (%d): %s", table, resp.StatusCode, string(body))
	}
	return body, nil
}

func (s *Supabase) Insert(table string, data []byte, useService bool) ([]byte, error) {
	path := fmt.Sprintf("/rest/v1/%s", table)
	resp, err := s.request("POST", path, data, useService)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return body, fmt.Errorf("insert %s status=%d: %s", table, resp.StatusCode, string(body))
	}
	return body, nil
}

func (s *Supabase) Update(table, query string, data []byte, useService bool) ([]byte, error) {
	path := fmt.Sprintf("/rest/v1/%s?%s", table, query)
	resp, err := s.request("PATCH", path, data, useService)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (s *Supabase) RPC(rpcName string, params map[string]interface{}) ([]byte, error) {
	path := fmt.Sprintf("/rest/v1/rpc/%s", rpcName)
	body, _ := json.Marshal(params)
	resp, err := s.request("POST", path, body, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (s *Supabase) GetProfile(userID, token string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/rest/v1/user_profiles?id=eq.%s&select=*", userID)
	req, _ := http.NewRequest("GET", s.URL+path, nil)
	req.Header.Set("apikey", s.AnonKey)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var profiles []map[string]interface{}
	json.Unmarshal(body, &profiles)
	if len(profiles) == 0 {
		return nil, fmt.Errorf("profile not found")
	}
	return profiles[0], nil
}
