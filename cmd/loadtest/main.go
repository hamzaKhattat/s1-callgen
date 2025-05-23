package main

import (
    "flag"
    "fmt"
    "log"
    "math/rand"
    "sync"
    "sync/atomic"
    "time"
    
    "github.com/s1-callgen/internal/config"
    "github.com/s1-callgen/internal/generator"
    "github.com/s1-callgen/internal/stats"
)

var (
    configPath   string
    duration     int
    rampUp       int
    maxCalls     int
    targetCPS    int
)

func init() {
    flag.StringVar(&configPath, "config", "configs/config.json", "Path to configuration file")
    flag.IntVar(&duration, "duration", 300, "Test duration in seconds")
    flag.IntVar(&rampUp, "rampup", 60, "Ramp-up period in seconds")
    flag.IntVar(&maxCalls, "max", 1000, "Maximum concurrent calls")
    flag.IntVar(&targetCPS, "cps", 10, "Target calls per second")
}

func main() {
    flag.Parse()
    fmt.Println("starting call gen...")
    
    cfg, err := config.LoadConfig(configPath)
    if err != nil {
        log.Fatalf("Error loading config: %v", err)
    }
    
    statsCollector := stats.NewCollector()
    gen, err := generator.NewGenerator(cfg, statsCollector)
    if err != nil {
        log.Fatalf("Error creating generator: %v", err)
    }
    fmt.Println(gen)
    log.Printf("Starting load test: %d seconds, %d CPS target, %d max calls", 
        duration, targetCPS, maxCalls)
    
    startTime := time.Now()
    endTime := startTime.Add(time.Duration(duration) * time.Second)
    rampEndTime := startTime.Add(time.Duration(rampUp) * time.Second)
    
    var wg sync.WaitGroup
    var activeCalls int32
    
    // Call generation loop
    ticker := time.NewTicker(time.Second / time.Duration(targetCPS))
    defer ticker.Stop()
    
    for time.Now().Before(endTime) {
        select {
        case <-ticker.C:
            current := atomic.LoadInt32(&activeCalls)
            
            // Calculate current target based on ramp-up
            var target int32
            if time.Now().Before(rampEndTime) {
                elapsed := time.Since(startTime).Seconds()
                rampProgress := elapsed / float64(rampUp)
                target = int32(float64(maxCalls) * rampProgress)
            } else {
                target = int32(maxCalls)
            }
            
            if current < target {
                wg.Add(1)
                atomic.AddInt32(&activeCalls, 1)
                
                go func() {
                    defer wg.Done()
                    defer atomic.AddInt32(&activeCalls, -1)
                    
                    // Simulate call with random duration
                    callDuration := 30 + rand.Intn(150)
                    time.Sleep(time.Duration(callDuration) * time.Second)
                }()
            }
            
            // Print progress every 10 seconds
            if int(time.Since(startTime).Seconds())%10 == 0 {
                log.Printf("Progress: %d active calls, target: %d", current, target)
            }
        }
    }
    
    log.Println("Waiting for all calls to complete...")
    wg.Wait()
    
    // Print final statistics
    statsCollector.PrintSummary()
    log.Printf("Load test completed in %v", time.Since(startTime))
}
