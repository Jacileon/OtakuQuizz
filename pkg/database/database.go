package database

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

func Connect() (*Supabase, error) {
	url := os.Getenv("SUPABASE_URL")
	if url == "" {
		url = os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	}
	anonKey := os.Getenv("SUPABASE_ANON_KEY")
	if anonKey == "" {
		anonKey = os.Getenv("NEXT_PUBLIC_SUPABASE_ANON_KEY")
	}
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if url == "" {
		return nil, fmt.Errorf("SUPABASE_URL is required")
	}
	if anonKey == "" {
		return nil, fmt.Errorf("SUPABASE_ANON_KEY is required")
	}

	return &Supabase{
		URL:        url,
		AnonKey:    anonKey,
		ServiceKey: serviceKey,
		Client:     &http.Client{},
	}, nil
}

func (s *Supabase) request(method, path string, body []byte, useService bool) (*http.Response, error) {
	url := s.URL + path
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
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

func (s *Supabase) SelectOne(table, query string, useService bool) ([]byte, error) {
	path := fmt.Sprintf("/rest/v1/%s?%s&limit=1", table, query)
	resp, err := s.request("GET", path, nil, useService)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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
		return body, fmt.Errorf("Insert %s status=%d body=%s", table, resp.StatusCode, string(body))
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

func (s *Supabase) Delete(table, query string, useService bool) error {
	path := fmt.Sprintf("/rest/v1/%s?%s", table, query)
	resp, err := s.request("DELETE", path, nil, useService)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
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

func (s *Supabase) GetUser(token string) (map[string]interface{}, error) {
	path := "/auth/v1/user"
	req, err := http.NewRequest("GET", s.URL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", s.AnonKey)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return result, nil
}

func (s *Supabase) GetProfile(userID string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/rest/v1/user_profiles?id=eq.%s&select=*", userID)
	resp, err := s.request("GET", path, nil, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var profiles []map[string]interface{}
	if err := json.Unmarshal(body, &profiles); err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return nil, fmt.Errorf("profile not found")
	}
	return profiles[0], nil
}

func (s *Supabase) GetProfiles(userIDs []string) ([]map[string]interface{}, error) {
	if len(userIDs) == 0 {
		return []map[string]interface{}{}, nil
	}
	ids := make([]string, len(userIDs))
	for i, id := range userIDs {
		ids[i] = "id.eq." + id
	}
	query := "or=(" + strings.Join(ids, ",") + ")"
	path := fmt.Sprintf("/rest/v1/user_profiles?%s", query)
	resp, err := s.request("GET", path, nil, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var profiles []map[string]interface{}
	json.Unmarshal(body, &profiles)
	return profiles, nil
}

func (s *Supabase) GetQuizzes(query string) ([]map[string]interface{}, error) {
	path := fmt.Sprintf("/rest/v1/quizzes?%s", query)
	resp, err := s.request("GET", path, nil, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("get quizzes failed (%d): %s", resp.StatusCode, string(body))
	}
	var quizzes []map[string]interface{}
	json.Unmarshal(body, &quizzes)
	return quizzes, nil
}

func (s *Supabase) GetQuiz(id string) (map[string]interface{}, error) {
	body, err := s.Select("quizzes", fmt.Sprintf("id=eq.%s&select=*", id), true)
	if err != nil {
		return nil, err
	}
	var quizzes []map[string]interface{}
	json.Unmarshal(body, &quizzes)
	if len(quizzes) == 0 {
		return nil, fmt.Errorf("quiz not found")
	}
	return quizzes[0], nil
}

func DBValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case map[string]interface{}:
		b, _ := json.Marshal(val)
		return string(b)
	case []interface{}:
		b, _ := json.Marshal(val)
		return string(b)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func DBInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case string:
		n := 0
		fmt.Sscanf(val, "%d", &n)
		return n
	}
	return 0
}

func DBFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case string:
		n := 0.0
		fmt.Sscanf(val, "%f", &n)
		return n
	}
	return 0
}

func DBBool(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val == "true"
	}
	return false
}
