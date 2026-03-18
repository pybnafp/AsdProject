package services

import (
	"asd/app/vo"
	"asd/conf"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

var ErrRagWorkerBusy = errors.New("all rag workers are busy")

// RagWorker represents an individual worker server
type RagWorker struct {
	ID             int
	URL            string
	ActiveRequests int
	mutex          sync.Mutex // Protects ActiveRequests for this specific worker
}

// IncrementActiveRequests safely increments the active request count
func (w *RagWorker) IncrementActiveRequests() {
	w.mutex.Lock()
	w.ActiveRequests++
	w.mutex.Unlock()
}

// DecrementActiveRequests safely decrements the active request count
func (w *RagWorker) DecrementActiveRequests() {
	w.mutex.Lock()
	if w.ActiveRequests > 0 {
		w.ActiveRequests--
	}
	w.mutex.Unlock()
}

// GetActiveRequests safely gets the active request count
func (w *RagWorker) GetActiveRequests() int {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.ActiveRequests
}

// ragService manages the pool of workers and their capacities
type ragService struct {
	maxWorkerCapacity int
	workers           []*RagWorker
	selectionMutex    sync.Mutex // Protects the worker selection logic and status reporting for consistency
}

// NewRagService creates and initializes a new ragService
func NewRagService(maxCapacity int, workerURLs []string) *ragService {
	service := &ragService{
		maxWorkerCapacity: maxCapacity,
		workers:           make([]*RagWorker, len(workerURLs)),
	}
	for i, url := range workerURLs {
		service.workers[i] = &RagWorker{
			ID:  i + 1,
			URL: url + "/chat/messages",
		}
	}
	return service
}

var RagService = NewRagService(10000, conf.CONFIG.ApiConfig.RagUrls)

// selectAndAssignWorker finds the least busy worker and atomically increments its request count.
// It returns nil if all workers are at capacity.
func (s *ragService) selectAndAssignWorker() *RagWorker {
	s.selectionMutex.Lock()
	defer s.selectionMutex.Unlock()

	availableWorkers := make([]*RagWorker, 0, len(s.workers))
	for _, w := range s.workers {
		// GetActiveRequests is individually mutexed, safe to call
		if w.GetActiveRequests() < s.maxWorkerCapacity {
			availableWorkers = append(availableWorkers, w)
		}
	}

	if len(availableWorkers) == 0 {
		// All workers are at capacity, randomly choose one
		// TODO: Return system busy error
		chosenWorker := s.workers[rand.Intn(len(s.workers))]
		chosenWorker.IncrementActiveRequests()
		return chosenWorker
	}

	// Sort by active requests to pick the least busy
	sort.Slice(availableWorkers, func(i, j int) bool {
		// Reading GetActiveRequests again for sorting is fine as we hold the selectionMutex
		return availableWorkers[i].GetActiveRequests() < availableWorkers[j].GetActiveRequests()
	})

	chosenWorker := availableWorkers[0]
	chosenWorker.IncrementActiveRequests() // ATOMICALLY INCREMENT INSIDE THE LOCK

	return chosenWorker
}

// GetMessagesFromRag is the Gin handler function for processing new requests
func (s *ragService) GetMessagesFromRag(
	ctx context.Context,
	prompt string,
	messages []map[string]string,
	enableGuidelines bool,
	enableResearches bool) ([]openai.ChatCompletionMessage, error) {

	selectedWorker := s.selectAndAssignWorker()

	defer selectedWorker.DecrementActiveRequests()
	requestBody := map[string]any{
		"prompt":            prompt,
		"enable_guidelines": enableGuidelines,
		"enable_researches": enableResearches,
		"chat_history":      messages,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// 使用带上下文的请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", selectedWorker.URL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer fTrytyIiFuPqDXYWFc4GwFbu2nH7Wh9K")

	// 创建一个带有超时的客户端，确保请求不会无限期挂起
	client := &http.Client{
		Timeout: 1800 * time.Second,
	}

	httpResponse, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("请求失败，状态码: %d, 响应: %s", httpResponse.StatusCode, string(body))
	}

	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	var responseData struct {
		Messages []openai.ChatCompletionMessage `json:"messages"`
	}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return nil, err
	}
	var retval []openai.ChatCompletionMessage
	for _, msg := range responseData.Messages {
		if msg.Content != "" {
			retval = append(retval, msg)
		}
	}
	return retval, nil
}

// GetRagWorkersStatus is the Gin handler function for returning worker statuses
func (s *ragService) GetRagWorkersStatus() []vo.RagWorkerStatus {
	s.selectionMutex.Lock() // Lock to get a consistent snapshot of all worker statuses
	defer s.selectionMutex.Unlock()

	statuses := make([]vo.RagWorkerStatus, len(s.workers))
	totalActiveRequests := 0

	for i, w := range s.workers {
		activeRequests := w.GetActiveRequests()
		isAtMax := activeRequests >= s.maxWorkerCapacity
		var loadPercent float64
		if s.maxWorkerCapacity > 0 {
			loadPercent = (float64(activeRequests) / float64(s.maxWorkerCapacity)) * 100
		}

		statuses[i] = vo.RagWorkerStatus{
			ID:                 w.ID,
			URL:                w.URL,
			ActiveRequests:     activeRequests,
			MaxCapacity:        s.maxWorkerCapacity,
			IsAtMaxCapacity:    isAtMax,
			CurrentLoadPercent: loadPercent,
		}
		totalActiveRequests += activeRequests
	}
	return statuses
}
