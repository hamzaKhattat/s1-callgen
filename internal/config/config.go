package config

import (
    "encoding/json"
    "os"
)

type Config struct {
    S2Server struct {
        Host     string `json:"host"`
        Port     int    `json:"port"`
        Protocol string `json:"protocol"`
    } `json:"s2_server"`
    
    Asterisk struct {
        AMI struct {
            Host     string `json:"host"`
            Port     int    `json:"port"`
            Username string `json:"username"`
            Password string `json:"password"`
        } `json:"ami"`
    } `json:"asterisk"`
    
    Database struct {
        Host     string `json:"host"`
        Port     int    `json:"port"`
        Username string `json:"username"`
        Password string `json:"password"`
        Name     string `json:"name"`
    } `json:"database"`
}

func LoadConfig(path string) (*Config, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var config Config
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&config); err != nil {
        return nil, err
    }

    return &config, nil
}
