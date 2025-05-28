package models

import "time"

type Call struct {
    ID          string
    ANI         string
    DNIS        string
    StartTime   time.Time
    EndTime     time.Time
    Duration    int
    Status      string
    SIPCallID   string
    LocalTag    string
    RemoteTag   string
}

type NumberPair struct {
    ANI  string
    DNIS string
}

type Config struct {
    S2Server struct {
        Host string `json:"host"`
        Port int    `json:"port"`
    } `json:"s2_server"`
    
    CallParams struct {
        ACDMin           int     `json:"acd_min"`
        ACDMax           int     `json:"acd_max"`
        ASR              float64 `json:"asr"`
        MaxConcurrent    int     `json:"max_concurrent"`
        CallsPerSecond   float64 `json:"calls_per_second"`
        RampUpTime       int     `json:"ramp_up_time"`
        RampDownTime     int     `json:"ramp_down_time"`
    } `json:"call_params"`
    
    Schedule struct {
        Weekday struct {
            StartHour int `json:"start_hour"`
            EndHour   int `json:"end_hour"`
        } `json:"weekday"`
        Weekend struct {
            StartHour int `json:"start_hour"`
            EndHour   int `json:"end_hour"`
        } `json:"weekend"`
    } `json:"schedule"`
}
