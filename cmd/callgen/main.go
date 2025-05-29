package main

import (
   "flag"
   "log"
   "os"
   "os/signal"
   "syscall"
   
   "github.com/s1-callgen/internal/config"
   "github.com/s1-callgen/internal/generator"
   "github.com/s1-callgen/internal/web"
)

func main() {
   var (
       configFile = flag.String("config", "configs/config.json", "Configuration file")
       csvFile    = flag.String("csv", "", "CSV file with number pairs")
       webOnly    = flag.Bool("web", false, "Start only web interface")
   )
   flag.Parse()
   
   log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
   log.Println("S1 Call Generator starting...")
   
   // Load configuration
   cfg, err := config.LoadConfig(*configFile)
   if err != nil {
       log.Fatalf("Failed to load config: %v", err)
   }
   
   // Create generator
   gen, err := generator.NewGenerator(cfg)
   if err != nil {
       log.Fatalf("Failed to create generator: %v", err)
   }
   
   // Load numbers
   if *csvFile != "" {
       if err := gen.LoadNumbersFromCSV(*csvFile); err != nil {
           log.Fatalf("Failed to load CSV: %v", err)
       }
   } else {
       log.Println("No CSV provided, using test numbers")
       gen.LoadTestNumbers()
   }
   
   // Start web interface if enabled
   if cfg.WebInterface.Enabled {
       webServer := web.NewWebServer(cfg, gen)
       go func() {
           if err := webServer.Start(); err != nil {
               log.Printf("Web server error: %v", err)
           }
       }()
   }
   
   // Start generator if not web-only mode
   if !*webOnly {
       if err := gen.Start(); err != nil {
           log.Fatalf("Failed to start generator: %v", err)
       }
   }
   
   // Wait for interrupt
   sigChan := make(chan os.Signal, 1)
   signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
   <-sigChan
   
   log.Println("Shutting down...")
   if !*webOnly {
       gen.Stop()
   }
}
