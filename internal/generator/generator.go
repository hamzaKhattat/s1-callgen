package generator

import (
   "encoding/csv"
   "fmt"
   "log"
   "math/rand"
   "os"
   "sync"
   "time"
   
   "github.com/s1-callgen/internal/models"
   "github.com/s1-callgen/internal/sip"
)

type Generator struct {
   config      *models.Config
   sipClient   *sip.Client
   numberPairs []models.NumberPair
   stats       *Statistics
   mu          sync.RWMutex
   stopChan    chan bool
   wg          sync.WaitGroup
}

type Statistics struct {
   TotalCalls      int64
   SuccessfulCalls int64
   FailedCalls     int64
   ActiveCalls     int64
   StartTime       time.Time
   mu              sync.Mutex
}

func NewGenerator(config *models.Config) (*Generator, error) {
   // Get local IP
   localIP := getLocalIP()
   
   // Create SIP client
   sipClient, err := sip.NewClient(localIP, 5070, config.S2Server.Host, config.S2Server.Port)
   if err != nil {
       return nil, err
   }
   
   return &Generator{
       config:    config,
       sipClient: sipClient,
       stats: &Statistics{
           StartTime: time.Now(),
       },
       stopChan: make(chan bool),
   }, nil
}

func (g *Generator) LoadNumbersFromCSV(filename string) error {
   file, err := os.Open(filename)
   if err != nil {
       return err
   }
   defer file.Close()
   
   reader := csv.NewReader(file)
   records, err := reader.ReadAll()
   if err != nil {
       return err
   }
   
   g.mu.Lock()
   defer g.mu.Unlock()
   
   g.numberPairs = make([]models.NumberPair, 0, len(records))
   for _, record := range records {
       if len(record) >= 2 {
           g.numberPairs = append(g.numberPairs, models.NumberPair{
               ANI:  record[0],
               DNIS: record[1],
           })
       }
   }
   
   log.Printf("[GENERATOR] Loaded %d number pairs", len(g.numberPairs))
   return nil
}

func (g *Generator) Start() error {
   if err := g.sipClient.Connect(); err != nil {
       return err
   }
   
   log.Printf("[GENERATOR] Starting call generation")
   log.Printf("Parameters: ACD=%d-%ds, ASR=%.0f%%, Max Concurrent=%d, CPS=%.2f",
       g.config.CallParams.ACDMin, g.config.CallParams.ACDMax,
       g.config.CallParams.ASR, g.config.CallParams.MaxConcurrent,
       g.config.CallParams.CallsPerSecond)
   
   // Start call generation loop
   g.wg.Add(1)
   go g.generateCalls()
   
   // Start statistics reporter
   g.wg.Add(1)
   go g.reportStatistics()
   
   return nil
}

func (g *Generator) generateCalls() {
   defer g.wg.Done()
   
   // Calculate interval between calls
   interval := time.Duration(float64(time.Second) / g.config.CallParams.CallsPerSecond)
   ticker := time.NewTicker(interval)
   defer ticker.Stop()
   
   for {
       select {
       case <-ticker.C:
           // Check if we should make a call based on schedule
           if !g.isWithinSchedule() {
               continue
           }
           
           // Check concurrent call limit
           if g.sipClient.GetActiveCallCount() >= g.config.CallParams.MaxConcurrent {
               continue
           }
           
           // Make a call
           g.wg.Add(1)
           go g.makeCall()
           
       case <-g.stopChan:
           return
       }
   }
}

func (g *Generator) makeCall() {
   defer g.wg.Done()
   
   // Get random number pair
   g.mu.RLock()
   if len(g.numberPairs) == 0 {
       g.mu.RUnlock()
       return
   }
   pair := g.numberPairs[rand.Intn(len(g.numberPairs))]
   g.mu.RUnlock()
   
   // Determine if call should be answered based on ASR
   shouldAnswer := rand.Float64()*100 < g.config.CallParams.ASR
   
   g.stats.mu.Lock()
   g.stats.TotalCalls++
   g.stats.ActiveCalls++
   g.stats.mu.Unlock()
   
   if shouldAnswer {
       // Random duration between ACDMin and ACDMax
       duration := time.Duration(g.config.CallParams.ACDMin+rand.Intn(g.config.CallParams.ACDMax-g.config.CallParams.ACDMin)) * time.Second
       
       err := g.sipClient.MakeCall(pair.ANI, pair.DNIS, duration)
       
       g.stats.mu.Lock()
       if err == nil {
           g.stats.SuccessfulCalls++
       } else {
           g.stats.FailedCalls++
           log.Printf("[GENERATOR] Call failed: %v", err)
       }
       g.stats.ActiveCalls--
       g.stats.mu.Unlock()
   } else {
       // Simulate rejected call
       time.Sleep(5 * time.Second)
       
       g.stats.mu.Lock()
       g.stats.FailedCalls++
       g.stats.ActiveCalls--
       g.stats.mu.Unlock()
   }
}

func (g *Generator) isWithinSchedule() bool {
   now := time.Now()
   hour := now.Hour()
   
   var schedule struct {
       StartHour int
       EndHour   int
   }
   
   if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
       schedule = g.config.Schedule.Weekend
   } else {
       schedule = g.config.Schedule.Weekday
   }
   
   return hour >= schedule.StartHour && hour < schedule.EndHour
}

func (g *Generator) reportStatistics() {
   defer g.wg.Done()
   
   ticker := time.NewTicker(10 * time.Second)
   defer ticker.Stop()
   
   for {
       select {
       case <-ticker.C:
           g.stats.mu.Lock()
           elapsed := time.Since(g.stats.StartTime)
           cps := float64(g.stats.TotalCalls) / elapsed.Seconds()
           asr := float64(g.stats.SuccessfulCalls) / float64(g.stats.TotalCalls) * 100
           
           log.Printf("[STATS] Total: %d, Success: %d, Failed: %d, Active: %d, CPS: %.2f, ASR: %.1f%%",
               g.stats.TotalCalls, g.stats.SuccessfulCalls, g.stats.FailedCalls,
               g.stats.ActiveCalls, cps, asr)
           g.stats.mu.Unlock()
           
       case <-g.stopChan:
           return
       }
   }
}

func (g *Generator) Stop() {
   close(g.stopChan)
   g.wg.Wait()
   g.sipClient.Close()
}

func getLocalIP() string {
   // Try to get the primary network interface IP
   interfaces, err := net.Interfaces()
   if err != nil {
       return "127.0.0.1"
   }
   
   for _, iface := range interfaces {
       if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
           continue
       }
       
       addrs, err := iface.Addrs()
       if err != nil {
           continue
       }
       
       for _, addr := range addrs {
           if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
               return ipnet.IP.String()
           }
       }
   }
   
   return "127.0.0.1"
}

func (g *Generator) LoadTestNumbers() {
   g.mu.Lock()
   defer g.mu.Unlock()
   
   g.numberPairs = []models.NumberPair{
       {ANI: "19543004835", DNIS: "50764137984"},
       {ANI: "19543004836", DNIS: "50764137985"},
       {ANI: "19543004837", DNIS: "50764137986"},
       {ANI: "19543004838", DNIS: "50764137987"},
       {ANI: "19543004839", DNIS: "50764137988"},
   }
}
