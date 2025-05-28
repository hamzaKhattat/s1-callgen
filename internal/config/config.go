package config

import (
   "encoding/json"
   "os"
   
   "github.com/s1-callgen/internal/models"
)

func LoadConfig(filename string) (*models.Config, error) {
   file, err := os.Open(filename)
   if err != nil {
       return nil, err
   }
   defer file.Close()
   
   config := &models.Config{}
   decoder := json.NewDecoder(file)
   if err := decoder.Decode(config); err != nil {
       return nil, err
   }
   
   // Set defaults
   if config.CallParams.ACDMin == 0 {
       config.CallParams.ACDMin = 30
   }
   if config.CallParams.ACDMax == 0 {
       config.CallParams.ACDMax = 180
   }
   if config.CallParams.ASR == 0 {
       config.CallParams.ASR = 70
   }
   if config.CallParams.MaxConcurrent == 0 {
       config.CallParams.MaxConcurrent = 100
   }
   if config.CallParams.CallsPerSecond == 0 {
       config.CallParams.CallsPerSecond = 1
   }
   
   return config, nil
}
