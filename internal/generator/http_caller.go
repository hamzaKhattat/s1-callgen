package generator

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type HTTPCaller struct {
    baseURL string
    client  *http.Client
}

type CallRequest struct {
    ID   string `json:"id"`
    ANI  string `json:"ani"`
    DNIS string `json:"dnis"`
}

func NewHTTPCaller(s2Host string, s2Port int) *HTTPCaller {
    return &HTTPCaller{
        baseURL: fmt.Sprintf("http://%s:%d", s2Host, s2Port),
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (h *HTTPCaller) SendCall(callID, ani, dnis string) error {
    req := CallRequest{
        ID:   callID,
        ANI:  ani,
        DNIS: dnis,
    }

    jsonData, err := json.Marshal(req)
    if err != nil {
        return err
    }

    // Send to S2's incoming endpoint
    resp, err := h.client.Post(
        h.baseURL+"/process-incoming",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return fmt.Errorf("failed to send call: %v", err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("S2 returned error: %d - %s", resp.StatusCode, string(body))
    }

    return nil
}

// sendCallViaHTTP sends a call using HTTP protocol (for testing without Asterisk)
func (g *Generator) sendCallViaHTTP(callID, ani, dnis string) error {
    httpCaller := NewHTTPCaller(g.config.S2Server.Host, g.config.S2Server.Port)
    return httpCaller.SendCall(callID, ani, dnis)
}

