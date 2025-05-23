package main

import (
    "flag"
//    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/s1-callgen/internal/config"
    "github.com/s1-callgen/internal/generator"
    "github.com/s1-callgen/internal/stats"
)

var (
    configPath   string
    testMode     bool
    callCount    int
    concurrent   int
    duration     int
    showStats    bool
    importCSV    string
)

func init() {
    flag.StringVar(&configPath, "config", "configs/config.json", "Path to configuration file")
    flag.BoolVar(&testMode, "test", false, "Run in test mode")
    flag.IntVar(&callCount, "calls", 1000, "Number of calls to generate in test mode")
    flag.IntVar(&concurrent, "concurrent", 50, "Number of concurrent calls")
    flag.IntVar(&duration, "duration", 300, "Test duration in seconds")
    flag.BoolVar(&showStats, "stats", false, "Show statistics only")
    flag.StringVar(&importCSV, "import", "", "Import ANI/DNIS from CSV file")
}

func main() {
    flag.Parse()

    // Load configuration
    cfg, err := config.LoadConfig(configPath)
    if err != nil {
        log.Fatalf("Error loading config: %v", err)
    }

    // Initialize statistics collector
    statsCollector := stats.NewCollector()

    // Create call generator
    gen, err := generator.NewGenerator(cfg, statsCollector)
    if err != nil {
        log.Fatalf("Error creating generator: %v", err)
    }

    // Handle import if specified
    if importCSV != "" {
        log.Printf("Importing ANI/DNIS pairs from %s", importCSV)
        count, err := gen.ImportFromCSV(importCSV)
        if err != nil {
            log.Fatalf("Error importing CSV: %v", err)
        }
        log.Printf("Successfully imported %d number pairs", count)
        return
    }

    // Show stats if requested
    if showStats {
        gen.ShowStats()
        return
    }

    // Setup signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    // Start generator based on mode
    if testMode {
        log.Printf("Starting test mode: %d calls with %d concurrent", callCount, concurrent)
        gen.RunTestMode(callCount, concurrent, time.Duration(duration)*time.Second)
    } else {
        log.Println("Starting production mode call generator")
        go gen.Start()
        
        // Wait for signal
        <-sigChan
        log.Println("Shutting down call generator...")
        gen.Stop()
    }

    // Show final statistics
    statsCollector.PrintSummary()
}
