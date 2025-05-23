package generator

import (
    "fmt"
    "log"
    "net"
    "strings"
    "time"
)

type AMIClient struct {
    host     string
    port     int
    username string
    password string
    conn     net.Conn
}

func NewAMIClient(host string, port int, username, password string) *AMIClient {
    return &AMIClient{
        host:     host,
        port:     port,
        username: username,
        password: password,
    }
}

func (a *AMIClient) Connect() error {
    addr := fmt.Sprintf("%s:%d", a.host, a.port)
    conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
    if err != nil {
        return fmt.Errorf("failed to connect to AMI: %v", err)
    }
    
    a.conn = conn
    
    // Read welcome message
    buf := make([]byte, 1024)
    _, err = a.conn.Read(buf)
    if err != nil {
        return fmt.Errorf("failed to read welcome message: %v", err)
    }
    
    // Login
    loginCmd := fmt.Sprintf("Action: Login\r\nUsername: %s\r\nSecret: %s\r\n\r\n", 
        a.username, a.password)
    
    _, err = a.conn.Write([]byte(loginCmd))
    if err != nil {
        return fmt.Errorf("failed to send login: %v", err)
    }
    
    // Read login response
    _, err = a.conn.Read(buf)
    if err != nil {
        return fmt.Errorf("failed to read login response: %v", err)
    }
    
    response := string(buf)
    if !strings.Contains(response, "Success") {
        return fmt.Errorf("login failed: %s", response)
    }
    
    log.Println("Connected to Asterisk AMI successfully")
    return nil
}

func (a *AMIClient) Originate(channel, context, exten, priority, callerID string) (string, error) {
    actionID := fmt.Sprintf("call_%d", time.Now().UnixNano())
    
    originateCmd := fmt.Sprintf(
        "Action: Originate\r\n"+
        "ActionID: %s\r\n"+
        "Channel: %s\r\n"+
        "Context: %s\r\n"+
        "Exten: %s\r\n"+
        "Priority: %s\r\n"+
        "CallerID: %s\r\n"+
        "Async: true\r\n"+
        "Variable: CALL_TYPE=test\r\n"+
        "\r\n",
        actionID, channel, context, exten, priority, callerID)
    
    _, err := a.conn.Write([]byte(originateCmd))
    if err != nil {
        return "", fmt.Errorf("failed to send originate: %v", err)
    }
    
    return actionID, nil
}

func (a *AMIClient) Disconnect() {
    if a.conn != nil {
        logoffCmd := "Action: Logoff\r\n\r\n"
        a.conn.Write([]byte(logoffCmd))
        a.conn.Close()
    }
}

// Enhanced generator method to use AMI
func (g *Generator) sendCallToS2WithAMI(callID, ani, dnis string) error {
    // Create AMI client
    amiClient := NewAMIClient(
        g.config.Asterisk.AMI.Host,
        g.config.Asterisk.AMI.Port,
        g.config.Asterisk.AMI.Username,
        g.config.Asterisk.AMI.Password,
    )
    
    // Connect to AMI
    if err := amiClient.Connect(); err != nil {
        return err
    }
    defer amiClient.Disconnect()
    
    // Originate call to S2
    channel := fmt.Sprintf("SIP/%s@TRUNK_TO_S2", dnis)
    context := "outgoing-to-s2"
    exten := dnis
    priority := "1"
    callerID := fmt.Sprintf("\"%s\" <%s>", ani, ani)
    
    actionID, err := amiClient.Originate(channel, context, exten, priority, callerID)
    if err != nil {
        return err
    }
    
    log.Printf("Call originated: ID=%s, ActionID=%s, ANI=%s, DNIS=%s", 
        callID, actionID, ani, dnis)
    
    return nil
}
