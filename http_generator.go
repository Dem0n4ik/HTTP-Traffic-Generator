package main

import (
    "bytes"
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "strings"
    "sync"
    "time"
)

type RequestResult struct {
    Status         string        `json:"status"`
    Duration       time.Duration `json:"duration"`
    ResponseLength int           `json:"response_length"`
    Error          string        `json:"error,omitempty"`
}

func makeRequest(ctx context.Context, client *http.Client, method string, url string, body string, headers map[string]string, wg *sync.WaitGroup, semaphore chan struct{}, results chan<- RequestResult, stats *Statistics, logFile *os.File) {
    defer wg.Done()
    defer func() { <-semaphore }() // Release the semaphore

    var req *http.Request
    var err error
    if method == "POST" || method == "PUT" || method == "PATCH" {
        req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer([]byte(body)))
    } else {
        req, err = http.NewRequestWithContext(ctx, method, url, nil)
    }
    if err != nil {
        logError(logFile, fmt.Sprintf("Error creating request: %v", err))
        results <- RequestResult{Error: err.Error()}
        stats.IncrementFailures()
        return
    }
    for key, value := range headers {
        req.Header.Set(key, value)
    }

    startTime := time.Now()
    resp, err := client.Do(req)
    duration := time.Since(startTime)

    if err != nil {
        logError(logFile, fmt.Sprintf("Error: %v", err))
        results <- RequestResult{Error: err.Error()}
        stats.IncrementFailures()
        return
    }
    defer resp.Body.Close()

    bodyBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        logError(logFile, fmt.Sprintf("Error reading response body: %v", err))
        results <- RequestResult{Error: err.Error()}
        stats.IncrementFailures()
        return
    }

    result := RequestResult{
        Status:         resp.Status,
        Duration:       duration,
        ResponseLength: len(bodyBytes),
    }
    results <- result
    stats.AddDuration(duration)
    stats.IncrementRequests()
}

func logError(logFile *os.File, message string) {
    logFile.WriteString(fmt.Sprintf("%s: %s\n", time.Now().Format(time.RFC3339), message))
}

type Statistics struct {
    sync.Mutex
    TotalDuration time.Duration
    RequestCount  int
    FailureCount  int
}

func (s *Statistics) AddDuration(duration time.Duration) {
    s.Lock()
    defer s.Unlock()
    s.TotalDuration += duration
}

func (s *Statistics) IncrementRequests() {
    s.Lock()
    defer s.Unlock()
    s.RequestCount++
}

func (s *Statistics) IncrementFailures() {
    s.Lock()
    defer s.Unlock()
    s.FailureCount++
}

func (s *Statistics) AverageDuration() time.Duration {
    s.Lock()
    defer s.Unlock()
    if s.RequestCount == 0 {
        return 0
    }
    return s.TotalDuration / time.Duration(s.RequestCount)
}

func writeResults(results <-chan RequestResult, outputFile *os.File, wg *sync.WaitGroup) {
    defer wg.Done()
    encoder := json.NewEncoder(outputFile)
    for result := range results {
        if err := encoder.Encode(result); err != nil {
            fmt.Fprintf(os.Stderr, "Error writing result: %v\n", err)
        }
    }
}

func main() {
    url := flag.String("url", "http://example.com", "URL to send requests to")
    numRequests := flag.Int("n", 10, "Number of requests")
    method := flag.String("method", "GET", "HTTP method (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, TRACE)")
    body := flag.String("body", "", "Request body (for POST, PUT, PATCH method)")
    timeout := flag.Int("timeout", 10, "Request timeout in seconds")
    headersFlag := flag.String("headers", "", "Custom headers (format: key1=value1,key2=value2)")
    interval := flag.Int("interval", 0, "Interval between requests in milliseconds")
    outputFile := flag.String("output", "results.json", "Output file to save results")
    errorFile := flag.String("errorlog", "errors.log", "File to log errors")
    maxConcurrentRequests := flag.Int("maxconcurrent", 5, "Maximum number of concurrent requests")
    flag.Parse()

    headers := make(map[string]string)
    if *headersFlag != "" {
        headersPairs := strings.Split(*headersFlag, ",")
        for _, pair := range headersPairs {
            kv := strings.SplitN(pair, "=", 2)
            if len(kv) == 2) {
                headers[kv[0]] = kv[1]
            }
        }
    }

    tr := &http.Transport{
        MaxIdleConns:       10,
        IdleConnTimeout:    30 * time.Second,
        DisableCompression: true,
    }
    client := &http.Client{
        Transport: tr,
        Timeout:   time.Duration(*timeout) * time.Second,
    }

    var wg sync.WaitGroup
    results := make(chan RequestResult, *numRequests)
    stats := &Statistics{}
    semaphore := make(chan struct{}, *maxConcurrentRequests)

    outputFileHandle, err := os.OpenFile(*outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        fmt.Println("Error opening output file:", err)
        return
    }
    defer outputFileHandle.Close()

    errorFileHandle, err := os.OpenFile(*errorFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        fmt.Println("Error opening error log file:", err)
        return
    }
    defer errorFileHandle.Close()

    wg.Add(*numRequests)
    for i := 0; i < *numRequests; i++ {
        semaphore <- struct{}{} // Acquire semaphore
        ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
        defer cancel()
        go makeRequest(ctx, client, *method, *url, *body, headers, &wg, semaphore, results, stats, errorFileHandle)
        if *interval > 0 {
            time.Sleep(time.Duration(*interval) * time.Millisecond)
        }
    }

    var writeWg sync.WaitGroup
    writeWg.Add(1)
    go writeResults(results, outputFileHandle, &writeWg)

    wg.Wait()
    close(results)
    writeWg.Wait()

    fmt.Println("All requests completed")
    fmt.Printf("Total requests: %d\n", stats.RequestCount)
    fmt.Printf("Failed requests: %d\n", stats.FailureCount)
    fmt.Printf("Average response time: %v\n", stats.AverageDuration())
}
