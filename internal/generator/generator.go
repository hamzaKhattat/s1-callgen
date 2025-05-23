package generator

import (
    "context"
    "encoding/csv"
    "fmt"
    "log"
    "math/rand"
    "os"
    "sync"
    "sync/atomic"
    "time"

    "github.com/s1-callgen/internal/config"
    "github.com/s1-callgen/internal/stats"
)

type Generator struct {
    config         *config.Config
    stats          *stats.Collector
    aniDnisPairs   []ANIDNISPair
    activeCalls    int32
    ctx            context.Context
    cancel         context.CancelFunc
    wg             sync.WaitGroup
    callParams     *CallParameters
    autopilotMode  bool
    mu             sync.RWMutex
}

type ANIDNISPair struct {
    ANI     string
    DNIS    string
    Country string
    Carrier string
}

type CallParameters struct {
    ACDMin           int
    ACDMax           int
    ASR              int
    WeekdayStart     string
    WeekdayEnd       string
    WeekendStart     string
    WeekendEnd       string
    IncreaseRate     int
    DecreaseRate     int
    MinSimultaneous  int
    MaxSimultaneous  int
}

func NewGenerator(cfg *config.Config, stats *stats.Collector) (*Generator, error) {
    ctx, cancel := context.WithCancel(context.Background())
    
    gen := &Generator{
        config:        cfg,
        stats:         stats,
        ctx:           ctx,
        cancel:        cancel,
        autopilotMode: true,
        callParams: &CallParameters{
            ACDMin:          30,
            ACDMax:          180,
            ASR:             70,
            WeekdayStart:    "08:00",
            WeekdayEnd:      "20:00",
            WeekendStart:    "10:00",
            WeekendEnd:      "18:00",
            IncreaseRate:    10,
            DecreaseRate:    5,
            MinSimultaneous: 5,
            MaxSimultaneous: 100,
        },
    }

    // Load ANI/DNIS pairs from database or default test data
    if err := gen.loadNumberPairs(); err != nil {
        return nil, fmt.Errorf("error loading number pairs: %v", err)
    }

    return gen, nil
}

func (g *Generator) loadNumberPairs() error {
    // For testing, load some default pairs
    g.aniDnisPairs = []ANIDNISPair{
        {ANI: "19543004835", DNIS: "50764137984", Country: "US", Carrier: "AT&T"},
        {ANI: "19543004836", DNIS: "50764137985", Country: "US", Carrier: "Verizon"},
        {ANI: "19543004837", DNIS: "50764137986", Country: "US", Carrier: "T-Mobile"},
        {ANI: "19543004838", DNIS: "50764137987", Country: "US", Carrier: "Sprint"},
        {ANI: "19543004839", DNIS: "50764137988", Country: "US", Carrier: "AT&T"},
    }

    log.Printf("Loaded %d ANI/DNIS pairs", len(g.aniDnisPairs))
    return nil
}

func (g *Generator) ImportFromCSV(filename string) (int, error) {
    file, err := os.Open(filename)
    if err != nil {
        return 0, err
    }
    defer file.Close()

    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        return 0, err
    }

    g.mu.Lock()
    defer g.mu.Unlock()

    g.aniDnisPairs = []ANIDNISPair{}
    
    for i, record := range records {
        if i == 0 && record[0] == "ANI" {
            continue // Skip header
        }
        
        if len(record) >= 2 {
            pair := ANIDNISPair{
                ANI:     record[0],
                DNIS:    record[1],
                Country: "US",
                Carrier: "Unknown",
            }
            
            if len(record) > 2 {
                pair.Country = record[2]
            }
            if len(record) > 3 {
                pair.Carrier = record[3]
            }
            
            g.aniDnisPairs = append(g.aniDnisPairs, pair)
        }
    }

    return len(g.aniDnisPairs), nil
}

func (g *Generator) Start() {
    log.Println("Starting call generator...")
    
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-g.ctx.Done():
            g.wg.Wait()
            return
        case <-ticker.C:
            g.adjustCallVolume()
            g.generateCalls()
        }
    }
}

func (g *Generator) Stop() {
    g.cancel()
    g.wg.Wait()
}

func (g *Generator) adjustCallVolume() {
    if !g.shouldBeActive() {
        // Decrease calls during off hours
        current := atomic.LoadInt32(&g.activeCalls)
        decrease := int32(g.callParams.DecreaseRate)
        if current > 0 {
            atomic.AddInt32(&g.activeCalls, -decrease)
        }
        return
    }

    current := atomic.LoadInt32(&g.activeCalls)
    
    if g.autopilotMode {
        // Organic traffic pattern
        hour := time.Now().Hour()
        minute := time.Now().Minute()
        
        // Create organic variations
        targetCalls := g.calculateOrganicTarget(hour, minute)
        
        if current < int32(targetCalls) {
            increase := int32(g.callParams.IncreaseRate)
            if current+increase <= int32(g.callParams.MaxSimultaneous) {
                atomic.AddInt32(&g.activeCalls, increase)
            }
        } else if current > int32(targetCalls) {
            decrease := int32(g.callParams.DecreaseRate)
            if current-decrease >= int32(g.callParams.MinSimultaneous) {
                atomic.AddInt32(&g.activeCalls, -decrease)
            }
        }
    }
}

func (g *Generator) calculateOrganicTarget(hour, minute int) int {
    // Create organic traffic pattern based on time of day
    baseLoad := float64(g.callParams.MinSimultaneous)
    maxLoad := float64(g.callParams.MaxSimultaneous)
    
    // Business hours peak (9 AM - 5 PM)
    var multiplier float64
    switch {
    case hour >= 9 && hour < 12:
        multiplier = 0.8 + float64(minute)/60*0.1 // Morning ramp-up
    case hour >= 12 && hour < 14:
        multiplier = 0.6 + rand.Float64()*0.2 // Lunch dip
    case hour >= 14 && hour < 17:
        multiplier = 0.85 + rand.Float64()*0.1 // Afternoon peak
    case hour >= 17 && hour < 20:
        multiplier = 0.7 - float64(hour-17)*0.1 // Evening decline
    default:
        multiplier = 0.1 + rand.Float64()*0.1 // Night/early morning
    }
    
    // Add some randomness for organic feel
    multiplier += (rand.Float64() - 0.5) * 0.1
    
    target := baseLoad + (maxLoad-baseLoad)*multiplier
    return int(target)
}

func (g *Generator) shouldBeActive() bool {
    now := time.Now()
    weekday := now.Weekday()
    
    var startTime, endTime string
    if weekday == time.Saturday || weekday == time.Sunday {
        startTime = g.callParams.WeekendStart
        endTime = g.callParams.WeekendEnd
    } else {
        startTime = g.callParams.WeekdayStart
        endTime = g.callParams.WeekdayEnd
    }
    
    start, _ := time.Parse("15:04", startTime)
    end, _ := time.Parse("15:04", endTime)
    currentTime, _ := time.Parse("15:04", now.Format("15:04"))
    
    return currentTime.After(start) && currentTime.Before(end)
}

func (g *Generator) generateCalls() {
    activeCalls := atomic.LoadInt32(&g.activeCalls)
    
    for i := int32(0); i < activeCalls; i++ {
        if rand.Intn(100) < g.callParams.ASR {
            g.wg.Add(1)
            go g.makeCall()
        } else {
            g.stats.RecordRejectedCall()
        }
    }
}

func (g *Generator) makeCall() {
    defer g.wg.Done()
    
    // Select random ANI/DNIS pair
    g.mu.RLock()
    if len(g.aniDnisPairs) == 0 {
        g.mu.RUnlock()
        return
    }
    pair := g.aniDnisPairs[rand.Intn(len(g.aniDnisPairs))]
    g.mu.RUnlock()
    
    callID := fmt.Sprintf("call_%d_%d", time.Now().Unix(), rand.Intn(10000))
    
    // Record call start
    g.stats.RecordCallStart(callID, pair.ANI, pair.DNIS)
    
    // Send call to S2
    err := g.sendCallToS2(callID, pair.ANI, pair.DNIS)
    if err != nil {
        log.Printf("Error sending call %s: %v", callID, err)
        g.stats.RecordCallFailed(callID)
        return
    }
    
    // Calculate call duration
    duration := g.callParams.ACDMin + rand.Intn(g.callParams.ACDMax-g.callParams.ACDMin)
    
    // Simulate call duration
    select {
    case <-time.After(time.Duration(duration) * time.Second):
        g.stats.RecordCallEnd(callID, duration)
    case <-g.ctx.Done():
        g.stats.RecordCallEnd(callID, 0)
        return
    }
}

func (g *Generator) sendCallToS2(callID, ani, dnis string) error {
    // Use HTTP for testing without Asterisk
    if g.config.S2Server.Protocol == "http" {
        return g.sendCallViaHTTP(callID, ani, dnis)
    }
    
    // Use AMI for production with Asterisk
    return g.sendCallToS2WithAMI(callID, ani, dnis)
}

func (g *Generator) RunTestMode(callCount, concurrent int, duration time.Duration) {
    log.Printf("Running test mode: %d calls, %d concurrent, %v duration", 
        callCount, concurrent, duration)
    
    startTime := time.Now()
    callsMade := int32(0)
    
    // Create worker pool
    semaphore := make(chan struct{}, concurrent)
    
    for i := 0; i < callCount && time.Since(startTime) < duration; i++ {
        semaphore <- struct{}{}
        
        g.wg.Add(1)
        go func(callNum int) {
            defer g.wg.Done()
            defer func() { <-semaphore }()
            
            g.makeCall()
            atomic.AddInt32(&callsMade, 1)
            
            if callsMade%100 == 0 {
                log.Printf("Progress: %d/%d calls made", callsMade, callCount)
            }
        }(i)
    }
    
    g.wg.Wait()
    log.Printf("Test complete: %d calls made in %v", callsMade, time.Since(startTime))
}

func (g *Generator) ShowStats() {
    g.stats.PrintSummary()
}
