package main

import (
    "encoding/csv"
    "fmt"
    "log"
    "math/rand"
    "os"
    "time"
)

func main() {
    rand.Seed(time.Now().UnixNano())
    
    // Generate test ANI/DNIS pairs
    file, err := os.Create("test_numbers.csv")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    writer := csv.NewWriter(file)
    defer writer.Flush()
    
    // Write header
    writer.Write([]string{"ANI", "DNIS", "Country", "Carrier"})
    
    carriers := []string{"AT&T", "Verizon", "T-Mobile", "Sprint"}
    
    // Generate 1000 test number pairs
    for i := 0; i < 1000; i++ {
        ani := fmt.Sprintf("1954300%04d", rand.Intn(10000))
        dnis := fmt.Sprintf("1507641%04d", rand.Intn(10000))
        country := "US"
        carrier := carriers[rand.Intn(len(carriers))]
        
        writer.Write([]string{ani, dnis, country, carrier})
    }
    
    log.Println("Generated 1000 test number pairs in test_numbers.csv")
}
