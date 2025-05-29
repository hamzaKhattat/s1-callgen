package models

import "time"

type Call struct {
    ID          string    `json:"id"`
    ANI         string    `json:"ani"`
    DNIS        string    `json:"dnis"`
    StartTime   time.Time `json:"start_time"`
    EndTime     time.Time `json:"end_time"`
    Duration    int       `json:"duration"`
    Status      string    `json:"status"`
    SIPCallID   string    `json:"sip_call_id"`
    LocalTag    string    `json:"local_tag"`
    RemoteTag   string    `json:"remote_tag"`
    Country     string    `json:"country"`
    Carrier     string    `json:"carrier"`
}

type NumberPair struct {
    ANI     string `json:"ani"`
    DNIS    string `json:"dnis"`
    Country string `json:"country"`
    Carrier string `json:"carrier"`
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
        MinConcurrent    int     `json:"min_concurrent"`
        CallsPerSecond   float64 `json:"calls_per_second"`
        RampUpTime       int     `json:"ramp_up_time"`
        RampDownTime     int     `json:"ramp_down_time"`
        RampUpRate       int     `json:"ramp_up_rate"`    // calls per minute
        RampDownRate     int     `json:"ramp_down_rate"`  // calls per minute
    } `json:"call_params"`
    
    Schedule struct {
        Enabled bool `json:"enabled"`
        Weekday struct {
            StartHour int `json:"start_hour"`
            EndHour   int `json:"end_hour"`
        } `json:"weekday"`
        Weekend struct {
            StartHour int `json:"start_hour"`
            EndHour   int `json:"end_hour"`
        } `json:"weekend"`
    } `json:"schedule"`
    
    Autopilot struct {
        Enabled            bool    `json:"enabled"`
        TargetASR          float64 `json:"target_asr"`
        AdjustmentInterval int     `json:"adjustment_interval"` // seconds
        MaxCPSAdjustment   float64 `json:"max_cps_adjustment"`
    } `json:"autopilot"`
    
    WebInterface struct {
        Enabled bool   `json:"enabled"`
        Port    int    `json:"port"`
        Auth    struct {
            Username string `json:"username"`
            Password string `json:"password"`
        } `json:"auth"`
    } `json:"web_interface"`
}

type Statistics struct {
    TotalCalls          int64     `json:"total_calls"`
    SuccessfulCalls     int64     `json:"successful_calls"`
    FailedCalls         int64     `json:"failed_calls"`
    ActiveCalls         int64     `json:"active_calls"`
    CurrentCPS          float64   `json:"current_cps"`
    AverageCallDuration float64   `json:"average_call_duration"`
    CurrentASR          float64   `json:"current_asr"`
    StartTime           time.Time `json:"start_time"`
    LastUpdate          time.Time `json:"last_update"`
    HourlyStats         map[int]*HourlyStats `json:"hourly_stats"`
}

type HourlyStats struct {
    Hour            int   `json:"hour"`
    TotalCalls      int64 `json:"total_calls"`
    SuccessfulCalls int64 `json:"successful_calls"`
    FailedCalls     int64 `json:"failed_calls"`
    PeakConcurrent  int64 `json:"peak_concurrent"`
}
