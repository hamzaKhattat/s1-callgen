package stats

import (
    "fmt"
    "sync"
    "time"
)

type CallStats struct {
    CallID    string
    ANI       string
    DNIS      string
    StartTime time.Time
    EndTime   time.Time
    Duration  int
    Status    string
}

type Collector struct {
    mu             sync.RWMutex
    totalCalls     int64
    activeCalls    int64
    completedCalls int64
    failedCalls    int64
    rejectedCalls  int64
    totalDuration  int64
    calls          map[string]*CallStats
}

func NewCollector() *Collector {
    return &Collector{
        calls: make(map[string]*CallStats),
    }
}

func (c *Collector) RecordCallStart(callID, ani, dnis string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.totalCalls++
    c.activeCalls++
    
    c.calls[callID] = &CallStats{
        CallID:    callID,
        ANI:       ani,
        DNIS:      dnis,
        StartTime: time.Now(),
        Status:    "active",
    }
}

func (c *Collector) RecordCallEnd(callID string, duration int) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if call, exists := c.calls[callID]; exists {
        call.EndTime = time.Now()
        call.Duration = duration
        call.Status = "completed"
        
        c.activeCalls--
        c.completedCalls++
        c.totalDuration += int64(duration)
    }
}

func (c *Collector) RecordCallFailed(callID string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if call, exists := c.calls[callID]; exists {
        call.Status = "failed"
        c.activeCalls--
        c.failedCalls++
    }
}

func (c *Collector) RecordRejectedCall() {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.rejectedCalls++
}

func (c *Collector) GetStats() map[string]interface{} {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    asr := float64(0)
    if c.totalCalls > 0 {
        asr = float64(c.completedCalls) / float64(c.totalCalls) * 100
    }
    
    acd := float64(0)
    if c.completedCalls > 0 {
        acd = float64(c.totalDuration) / float64(c.completedCalls)
    }
    
    return map[string]interface{}{
        "total_calls":     c.totalCalls,
        "active_calls":    c.activeCalls,
        "completed_calls": c.completedCalls,
        "failed_calls":    c.failedCalls,
        "rejected_calls":  c.rejectedCalls,
        "asr":             fmt.Sprintf("%.2f%%", asr),
        "acd":             fmt.Sprintf("%.2f seconds", acd),
    }
}

func (c *Collector) PrintSummary() {
    stats := c.GetStats()
    
    fmt.Println("\n=== Call Generation Statistics ===")
    fmt.Printf("Total Calls:     %v\n", stats["total_calls"])
    fmt.Printf("Active Calls:    %v\n", stats["active_calls"])
    fmt.Printf("Completed Calls: %v\n", stats["completed_calls"])
    fmt.Printf("Failed Calls:    %v\n", stats["failed_calls"])
    fmt.Printf("Rejected Calls:  %v\n", stats["rejected_calls"])
    fmt.Printf("ASR:             %v\n", stats["asr"])
    fmt.Printf("ACD:             %v\n", stats["acd"])
    fmt.Println("================================")
}
