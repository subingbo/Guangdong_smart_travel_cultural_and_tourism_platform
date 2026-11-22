// agent-bench is a small, dependency-free harness for the agent executor
// comparison described in docs/agent/benchmark.md.
//
//	go run ./cmd/agent-bench -mode fake-llm -addr :9099
//	go run ./cmd/agent-bench -mode bench -arm spawn4 -concurrency 4 -runs 3 -out spawn4.jsonl
//	go run ./cmd/agent-bench -mode analyze -in spawn4.jsonl
package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	mode := flag.String("mode", "bench", "fake-llm | bench | analyze")
	addr := flag.String("addr", ":9099", "fake LLM listen address")
	baseURL := flag.String("base-url", "http://127.0.0.1:8081", "API base URL")
	arm := flag.String("arm", "spawn4", "arm label (fixed4 | fixed32 | spawn32 ...)")
	concurrency := flag.Int("concurrency", 4, "concurrent runs")
	runs := flag.Int("runs", 3, "runs per worker")
	users := flag.Int("users", -1, "users to create (default = concurrency)")
	password := flag.String("password", "benchpass123", "bench user password")
	out := flag.String("out", "agent-bench.jsonl", "output JSONL")
	in := flag.String("in", "agent-bench.jsonl", "input JSONL")
	timeout := flag.Duration("timeout", 5*time.Minute, "per-run timeout")
	delay := flag.Duration("delay", 50*time.Millisecond, "fake LLM chunk delay")
	flag.Parse()

	switch *mode {
	case "fake-llm":
		runFakeLLM(*addr, *delay)
	case "bench":
		if err := runBench(*baseURL, *arm, *concurrency, *runs, *users, *password, *out, *timeout); err != nil {
			log.Fatal(err)
		}
	case "analyze":
		if err := analyze(*in); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown mode %q", *mode)
	}
}

// ---------------------------------------------------------------------------
// Fake OpenAI-compatible LLM
// ---------------------------------------------------------------------------

type chatRequest struct {
	Model    string        `json:"model"`
	Stream   bool          `json:"stream"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name"`
	ToolCallID string     `json:"tool_call_id"`
	ToolCalls  []toolCall `json:"tool_calls"`
}

type toolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type decision struct {
	Content  string
	ToolName string
	ToolArgs string
}

var fakeChunkDelay = 50 * time.Millisecond

func runFakeLLM(addr string, delay time.Duration) {
	fakeChunkDelay = delay
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", handleChat)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	log.Printf("fake LLM listening on %s", addr)
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req chatRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 2<<20)).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	dec := decide(&req)
	if !req.Stream {
		writeNonStream(w, dec)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	if dec.ToolName != "" {
		args := strings.ReplaceAll(dec.ToolArgs, "\n", "")
		first := map[string]any{
			"choices": []any{map[string]any{
				"delta": map[string]any{
					"tool_calls": []any{map[string]any{
						"index": 0, "id": "call_" + randomID()[:8], "type": "function",
						"function": map[string]any{"name": dec.ToolName, "arguments": args},
					}},
				},
				"finish_reason": nil,
			}},
		}
		writeSSE(w, flusher, first)
		writeSSE(w, flusher, map[string]any{
			"choices": []any{map[string]any{"delta": map[string]any{}, "finish_reason": "tool_calls"}},
		})
		writeSSE(w, flusher, usageChunk())
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
		return
	}

	content := dec.Content
	chunks := splitChunks(content, 4)
	for _, chunk := range chunks {
		writeSSE(w, flusher, map[string]any{
			"choices": []any{map[string]any{
				"delta":         map[string]any{"content": chunk},
				"finish_reason": nil,
			}},
		})
		time.Sleep(fakeChunkDelay)
	}
	writeSSE(w, flusher, map[string]any{
		"choices": []any{map[string]any{"delta": map[string]any{}, "finish_reason": "stop"}},
	})
	writeSSE(w, flusher, usageChunk())
	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func decide(req *chatRequest) decision {
	var lastTool *chatMessage
	for i := range req.Messages {
		if req.Messages[i].Role == "tool" {
			lastTool = &req.Messages[i]
		}
	}
	if lastTool == nil {
		return decision{
			ToolName: "search_places",
			ToolArgs: `{"query":"广州","limit":3}`,
		}
	}
	if strings.Contains(lastTool.Content, "tool_error") ||
		strings.Contains(lastTool.Content, "duplicate_call") ||
		strings.Contains(lastTool.Content, "budget_exhausted") {
		return decision{Content: "工具没有返回可用数据，本次先到这里。"}
	}
	switch {
	case strings.Contains(lastTool.Content, `"places"`):
		var result struct {
			Places []struct {
				PlaceID int64 `json:"placeId"`
			} `json:"places"`
		}
		if err := json.Unmarshal([]byte(lastTool.Content), &result); err != nil || len(result.Places) == 0 {
			return decision{Content: "没有找到合适的景点。"}
		}
		ids := make([]int64, 0, 3)
		for _, p := range result.Places {
			ids = append(ids, p.PlaceID)
			if len(ids) == 3 {
				break
			}
		}
		args, _ := json.Marshal(map[string]any{
			"placeIds": ids, "when": "明天下午", "mode": "DRIVE",
		})
		return decision{ToolName: "plan_itinerary", ToolArgs: string(args)}
	case strings.Contains(lastTool.Content, `"solutions"`):
		var result struct {
			Solutions []struct {
				SolutionID string `json:"solutionId"`
			} `json:"solutions"`
		}
		if err := json.Unmarshal([]byte(lastTool.Content), &result); err != nil || len(result.Solutions) == 0 {
			return decision{Content: "这个时间窗内排不出可行方案，可以换一天或减少地点。"}
		}
		args, _ := json.Marshal(map[string]any{
			"solutionId": result.Solutions[0].SolutionID,
			"title":      "Bench plan",
			"summary":    "synthetic benchmark plan",
		})
		return decision{ToolName: "submit_plan", ToolArgs: string(args)}
	default:
		return decision{Content: "行程草稿已生成，可以确认保存。"}
	}
}

func writeSSE(w io.Writer, flusher http.Flusher, payload any) {
	raw, _ := json.Marshal(payload)
	fmt.Fprintf(w, "data: %s\n\n", raw)
	flusher.Flush()
}

func usageChunk() map[string]any {
	return map[string]any{
		"choices": []any{},
		"usage":   map[string]any{"prompt_tokens": 120, "completion_tokens": 40, "total_tokens": 160},
	}
}

func writeNonStream(w http.ResponseWriter, dec decision) {
	message := map[string]any{"role": "assistant", "content": dec.Content}
	finish := "stop"
	if dec.ToolName != "" {
		message["content"] = ""
		message["tool_calls"] = []any{map[string]any{
			"id": "call_" + randomID()[:8], "type": "function",
			"function": map[string]any{"name": dec.ToolName, "arguments": dec.ToolArgs},
		}}
		finish = "tool_calls"
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"choices": []any{map[string]any{"message": message, "finish_reason": finish}},
		"usage":   map[string]any{"prompt_tokens": 120, "completion_tokens": 40, "total_tokens": 160},
	})
}

func splitChunks(text string, n int) []string {
	if n < 1 {
		n = 1
	}
	runes := []rune(text)
	if len(runes) == 0 {
		return []string{"..."}
	}
	size := int(math.Ceil(float64(len(runes)) / float64(n)))
	out := make([]string, 0, n)
	for i := 0; i < len(runes); i += size {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[i:end]))
	}
	return out
}

// ---------------------------------------------------------------------------
// Bench client
// ---------------------------------------------------------------------------

type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

type runRecord struct {
	Type        string `json:"type"`
	Arm         string `json:"arm"`
	Concurrency int    `json:"concurrency"`
	Worker      int    `json:"worker"`
	Index       int    `json:"index"`
	RunID       int64  `json:"runId"`
	Status      string `json:"status"`
	LatencyMS   int64  `json:"latencyMs"`
	QueueMS     int64  `json:"queueMs"`
	Steps       int    `json:"steps"`
	HasPlan     bool   `json:"hasPlan"`
	ErrorCode   string `json:"errorCode,omitempty"`
	SubmittedAt string `json:"submittedAt"`
	FinishedAt  string `json:"finishedAt"`
}

type metricsRecord struct {
	Type    string         `json:"type"`
	Arm     string         `json:"arm"`
	At      string         `json:"at"`
	Metrics map[string]any `json:"metrics"`
}

type snapshot struct {
	RunID     int64  `json:"runId"`
	Status    string `json:"status"`
	StepCount int    `json:"stepCount"`
	Plan      any    `json:"plan"`
	Error     *struct {
		Code string `json:"code"`
	} `json:"error"`
	CreatedAt  string `json:"createdAt"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt"`
}

func runBench(baseURL, arm string, concurrency, runs, users int, password, out string, timeout time.Duration) error {
	baseURL = strings.TrimRight(baseURL, "/")
	if users <= 0 {
		users = concurrency
	}
	client := &http.Client{Timeout: 30 * time.Second}
	tokens := make([]string, 0, users)
	for i := 0; i < users; i++ {
		token, err := ensureUser(client, baseURL, arm, i, password)
		if err != nil {
			return fmt.Errorf("user %d: %w", i, err)
		}
		tokens = append(tokens, token)
	}

	file, err := os.Create(out)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()

	var mu sync.Mutex
	writeRecord := func(v any) {
		raw, _ := json.Marshal(v)
		mu.Lock()
		fmt.Fprintln(writer, string(raw))
		mu.Unlock()
	}

	stopMetrics := make(chan struct{})
	var metricsWG sync.WaitGroup
	if len(tokens) > 0 {
		metricsWG.Add(1)
		go func() {
			defer metricsWG.Done()
			sampleMetrics(client, baseURL, arm, tokens[0], stopMetrics, writeRecord)
		}()
	}

	start := time.Now()
	var wg sync.WaitGroup
	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			token := tokens[worker%len(tokens)]
			for i := 0; i < runs; i++ {
				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				record := runOnce(ctx, client, baseURL, token, arm, concurrency, worker, i)
				cancel()
				writeRecord(record)
			}
		}(worker)
	}
	wg.Wait()
	close(stopMetrics)
	metricsWG.Wait()
	log.Printf("arm=%s concurrency=%d runs=%d wall=%s", arm, concurrency, concurrency*runs, time.Since(start).Round(time.Millisecond))
	return nil
}

func ensureUser(client *http.Client, baseURL, arm string, i int, password string) (string, error) {
	username := fmt.Sprintf("bench_%s_%d", sanitize(arm), i)
	body, _ := json.Marshal(map[string]string{
		"username": username, "password": password, "nickname": username,
	})
	token, err := authCall(client, baseURL+"/api/auth/register", body)
	if err == nil {
		return token, nil
	}
	token, loginErr := authCall(client, baseURL+"/api/auth/login", body)
	if loginErr != nil {
		return "", fmt.Errorf("register: %v; login: %v", err, loginErr)
	}
	return token, nil
}

func authCall(client *http.Client, url string, body []byte) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var env envelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || !env.Success {
		return "", fmt.Errorf("%s: %s", resp.Status, env.Message)
	}
	var auth struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(env.Data, &auth); err != nil {
		return "", err
	}
	if auth.Token == "" {
		return "", fmt.Errorf("no token in response")
	}
	return auth.Token, nil
}

func runOnce(ctx context.Context, client *http.Client, baseURL, token, arm string, concurrency, worker, index int) runRecord {
	submitted := time.Now()
	record := runRecord{
		Type: "run", Arm: arm, Concurrency: concurrency, Worker: worker, Index: index,
		SubmittedAt: submitted.Format(time.RFC3339Nano),
	}
	createBody, _ := json.Marshal(map[string]any{
		"message":         "请帮我规划明天下午在广州的三个景点",
		"clientRequestId": randomUUID(),
	})
	var snap snapshot
	if err := callJSON(ctx, client, http.MethodPost, baseURL+"/api/agent/runs", token, createBody, &snap); err != nil {
		record.Status = "request_failed"
		record.ErrorCode = err.Error()
		record.LatencyMS = time.Since(submitted).Milliseconds()
		return record
	}
	record.RunID = snap.RunID
	if snap.StartedAt != "" && snap.CreatedAt != "" {
		if created, err1 := time.Parse(time.RFC3339Nano, snap.CreatedAt); err1 == nil {
			if started, err2 := time.Parse(time.RFC3339Nano, snap.StartedAt); err2 == nil {
				record.QueueMS = started.Sub(created).Milliseconds()
			}
		}
	}
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		snap = snapshot{}
		if err := callJSON(ctx, client, http.MethodGet, fmt.Sprintf("%s/api/agent/runs/%d", baseURL, record.RunID), token, nil, &snap); err != nil {
			record.Status = "poll_failed"
			record.ErrorCode = err.Error()
			break
		}
		if terminal(snap.Status) {
			record.Status = snap.Status
			record.Steps = snap.StepCount
			record.HasPlan = snap.Plan != nil
			if snap.Error != nil {
				record.ErrorCode = snap.Error.Code
			}
			break
		}
		select {
		case <-ctx.Done():
			record.Status = "client_timeout"
			record.ErrorCode = ctx.Err().Error()
		case <-ticker.C:
			continue
		}
		if record.Status == "client_timeout" {
			break
		}
	}
	record.LatencyMS = time.Since(submitted).Milliseconds()
	record.FinishedAt = time.Now().Format(time.RFC3339Nano)
	return record
}

func callJSON(ctx context.Context, client *http.Client, method, url, token string, body []byte, out any) error {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var env envelope
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(&env); err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || !env.Success {
		return fmt.Errorf("%s: %s", resp.Status, env.Message)
	}
	return json.Unmarshal(env.Data, out)
}

func sampleMetrics(client *http.Client, baseURL, arm, token string, stop <-chan struct{}, write func(any)) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			var metrics map[string]any
			err := callJSON(ctx, client, http.MethodGet, baseURL+"/api/agent/debug/metrics", token, nil, &metrics)
			cancel()
			if err != nil {
				continue
			}
			write(metricsRecord{Type: "metrics", Arm: arm, At: time.Now().Format(time.RFC3339Nano), Metrics: metrics})
		}
	}
}

func terminal(status string) bool {
	switch status {
	case "succeeded", "failed", "canceled", "interrupted":
		return true
	default:
		return false
	}
}

// ---------------------------------------------------------------------------
// Analyze
// ---------------------------------------------------------------------------

func analyze(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	type key struct {
		arm         string
		concurrency int
	}
	groups := make(map[key][]runRecord)
	maxGoroutines := make(map[string]float64)
	maxHeap := make(map[string]float64)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 8<<20)
	for scanner.Scan() {
		var probe struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &probe); err != nil {
			continue
		}
		switch probe.Type {
		case "run":
			var rec runRecord
			if err := json.Unmarshal(scanner.Bytes(), &rec); err == nil {
				groups[key{rec.Arm, rec.Concurrency}] = append(groups[key{rec.Arm, rec.Concurrency}], rec)
			}
		case "metrics":
			var rec metricsRecord
			if err := json.Unmarshal(scanner.Bytes(), &rec); err == nil {
				if v, ok := rec.Metrics["goroutines"].(float64); ok && v > maxGoroutines[rec.Arm] {
					maxGoroutines[rec.Arm] = v
				}
				if v, ok := rec.Metrics["heapAlloc"].(float64); ok && v > maxHeap[rec.Arm] {
					maxHeap[rec.Arm] = v
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	fmt.Printf("%-10s %5s %6s %7s %9s %9s %9s %10s %10s\n",
		"arm", "conc", "runs", "ok%", "p50ms", "p95ms", "p99ms", "queueP50", "req/s")
	keys := make([]key, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].arm != keys[j].arm {
			return keys[i].arm < keys[j].arm
		}
		return keys[i].concurrency < keys[j].concurrency
	})
	for _, k := range keys {
		records := groups[k]
		latencies := make([]int64, 0, len(records))
		queues := make([]int64, 0, len(records))
		ok := 0
		var minSubmit, maxFinish time.Time
		for _, rec := range records {
			latencies = append(latencies, rec.LatencyMS)
			queues = append(queues, rec.QueueMS)
			if rec.Status == "succeeded" {
				ok++
			}
			if t, err := time.Parse(time.RFC3339Nano, rec.SubmittedAt); err == nil {
				if minSubmit.IsZero() || t.Before(minSubmit) {
					minSubmit = t
				}
			}
			if t, err := time.Parse(time.RFC3339Nano, rec.FinishedAt); err == nil {
				if t.After(maxFinish) {
					maxFinish = t
				}
			}
		}
		sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
		sort.Slice(queues, func(i, j int) bool { return queues[i] < queues[j] })
		wall := maxFinish.Sub(minSubmit).Seconds()
		rate := 0.0
		if wall > 0 {
			rate = float64(len(records)) / wall
		}
		fmt.Printf("%-10s %5d %6d %6.1f %9d %9d %9d %10d %10.2f\n",
			k.arm, k.concurrency, len(records), 100*float64(ok)/float64(len(records)),
			percentile(latencies, 50), percentile(latencies, 95), percentile(latencies, 99),
			percentile(queues, 50), rate)
	}
	if len(maxGoroutines) > 0 {
		fmt.Println()
		for arm, v := range maxGoroutines {
			fmt.Printf("arm=%s max goroutines=%.0f max heap=%.1f MiB\n", arm, v, maxHeap[arm]/1024/1024)
		}
	}
	return nil
}

func percentile(sorted []int64, p int) int64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(float64(p)/100*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// ---------------------------------------------------------------------------
// Small helpers
// ---------------------------------------------------------------------------

func randomUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func randomID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func sanitize(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return "arm"
	}
	if len(out) > 20 {
		out = out[:20]
	}
	return string(out)
}
