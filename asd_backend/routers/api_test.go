package routers_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"testing"
)

func TestSimultaneousChatCreateStream(t *testing.T) {
	// Step 1: Login and get cookies
	loginURL := "http://localhost:9081/api/login/mobile-pwd"
	loginPayload := `{
		"mobile": "18888888888",
		"password": "123456"
	}`

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(loginPayload))
	if err != nil {
		t.Fatalf("Failed to create login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("Login failed: %s", string(body))
	}

	// Step 2: Prepare to send 12 simultaneous requests
	var wg sync.WaitGroup
	numRequests := 12
	results := make([]string, numRequests)
	errors := make([]error, numRequests)

	for i := range numRequests {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			chatURL := "http://localhost:9081/api/chat/create_stream"
			chatPayload := `{
				"prompt": "自闭症的成因是什么",
				"enable_guidelines": true,
				"enable_researches": true
			}`
			req, err := http.NewRequest("POST", chatURL, bytes.NewBufferString(chatPayload))
			if err != nil {
				errors[idx] = fmt.Errorf("Failed to create chat request: %v", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				errors[idx] = fmt.Errorf("Chat request failed: %v", err)
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			results[idx] = string(body)
		}(i)
	}
	wg.Wait()

	// Step 3: Print results and errors
	for i, res := range results {
		if errors[i] != nil {
			t.Errorf("Request %d error: %v", i, errors[i])
		} else {
			t.Logf("Request %d response: %s", i, res)
		}
	}
}
