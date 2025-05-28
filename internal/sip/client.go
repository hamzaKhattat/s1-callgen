package sip

import (
    "fmt"
    "log"
    "math/rand"
    "net"
    "strings"
    "sync"
    "time"
    
    "github.com/s1-callgen/internal/models"
)

type Client struct {
    localIP    string
    localPort  int
    remoteIP   string
    remotePort int
    transport  string
    conn       net.Conn
    mu         sync.Mutex
    activeCalls map[string]*models.Call
    rtpPorts   chan int
}

func NewClient(localIP string, localPort int, remoteIP string, remotePort int) (*Client, error) {
    // Initialize RTP port pool (even ports from 10000-20000)
    rtpPorts := make(chan int, 1000)
    for i := 10000; i < 20000; i += 2 {
        rtpPorts <- i
    }
    
    return &Client{
        localIP:     localIP,
        localPort:   localPort,
        remoteIP:    remoteIP,
        remotePort:  remotePort,
        transport:   "UDP",
        activeCalls: make(map[string]*models.Call),
        rtpPorts:    rtpPorts,
    }, nil
}

func (c *Client) Connect() error {
    addr := fmt.Sprintf("%s:%d", c.remoteIP, c.remotePort)
    conn, err := net.Dial("udp", addr)
    if err != nil {
        return fmt.Errorf("failed to connect: %v", err)
    }
    c.conn = conn
    
    // Start listening for responses
    go c.listenResponses()
    
    log.Printf("[SIP] Connected to %s", addr)
    return nil
}

func (c *Client) MakeCall(ani, dnis string, duration time.Duration) error {
    call := &models.Call{
        ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
        ANI:       ani,
        DNIS:      dnis,
        StartTime: time.Now(),
        Status:    "INITIATING",
        SIPCallID: c.generateCallID(),
        LocalTag:  c.generateTag(),
    }
    
    // Get RTP port
    rtpPort := <-c.rtpPorts
    defer func() { c.rtpPorts <- rtpPort }()
    
    // Send INVITE
    invite := c.buildINVITE(call, rtpPort)
    if err := c.sendMessage(invite); err != nil {
        return err
    }
    
    c.mu.Lock()
    c.activeCalls[call.SIPCallID] = call
    c.mu.Unlock()
    
    log.Printf("[SIP] Call initiated: %s -> %s (CallID: %s)", ani, dnis, call.SIPCallID)
    
    // Simulate call duration
    time.Sleep(duration)
    
    // Send BYE
    bye := c.buildBYE(call)
    c.sendMessage(bye)
    
    c.mu.Lock()
    delete(c.activeCalls, call.SIPCallID)
    c.mu.Unlock()
    
    return nil
}

func (c *Client) buildINVITE(call *models.Call, rtpPort int) string {
    branch := c.generateBranch()
    
    sdp := fmt.Sprintf(
        "v=0\r\n" +
        "o=- %d %d IN IP4 %s\r\n" +
        "s=S1 Call Generator\r\n" +
        "c=IN IP4 %s\r\n" +
        "t=0 0\r\n" +
        "m=audio %d RTP/AVP 0 8 101\r\n" +
        "a=rtpmap:0 PCMU/8000\r\n" +
        "a=rtpmap:8 PCMA/8000\r\n" +
        "a=rtpmap:101 telephone-event/8000\r\n" +
        "a=fmtp:101 0-16\r\n" +
        "a=sendrecv\r\n",
        time.Now().Unix(), time.Now().Unix(), c.localIP, c.localIP, rtpPort,
    )
    
    invite := fmt.Sprintf(
        "INVITE sip:%s@%s:%d SIP/2.0\r\n" +
        "Via: SIP/2.0/%s %s:%d;branch=%s;rport\r\n" +
        "Max-Forwards: 70\r\n" +
        "From: <sip:%s@%s>;tag=%s\r\n" +
        "To: <sip:%s@%s>\r\n" +
        "Call-ID: %s\r\n" +
        "CSeq: 1 INVITE\r\n" +
        "Contact: <sip:%s@%s:%d>\r\n" +
        "Content-Type: application/sdp\r\n" +
        "Content-Length: %d\r\n" +
        "User-Agent: S1-CallGenerator/1.0\r\n" +
        "\r\n%s",
        call.DNIS, c.remoteIP, c.remotePort,
        c.transport, c.localIP, c.localPort, branch,
        call.ANI, c.localIP, call.LocalTag,
        call.DNIS, c.remoteIP,
        call.SIPCallID,
        call.ANI, c.localIP, c.localPort,
        len(sdp), sdp,
    )
    
    return invite
}

func (c *Client) buildBYE(call *models.Call) string {
    branch := c.generateBranch()
    
    bye := fmt.Sprintf(
        "BYE sip:%s@%s:%d SIP/2.0\r\n" +
        "Via: SIP/2.0/%s %s:%d;branch=%s;rport\r\n" +
       "Max-Forwards: 70\r\n" +
       "From: <sip:%s@%s>;tag=%s\r\n" +
       "To: <sip:%s@%s>;tag=%s\r\n" +
       "Call-ID: %s\r\n" +
       "CSeq: 2 BYE\r\n" +
       "Content-Length: 0\r\n" +
       "\r\n",
       c.transport, c.localIP, c.localPort, branch,
       call.ANI, c.localIP, call.LocalTag,
       call.DNIS, c.remoteIP, call.RemoteTag,
       call.SIPCallID,
   )
   
   return bye
}

func (c *Client) sendMessage(message string) error {
   c.mu.Lock()
   defer c.mu.Unlock()
   
   _, err := c.conn.Write([]byte(message))
   return err
}

func (c *Client) listenResponses() {
   buffer := make([]byte, 4096)
   for {
       n, err := c.conn.Read(buffer)
       if err != nil {
           log.Printf("[SIP] Error reading response: %v", err)
           continue
       }
       
       response := string(buffer[:n])
       c.handleResponse(response)
   }
}

func (c *Client) handleResponse(response string) {
   lines := strings.Split(response, "\r\n")
   if len(lines) < 1 {
       return
   }
   
   // Parse status line
   parts := strings.Split(lines[0], " ")
   if len(parts) < 3 {
       return
   }
   
   if parts[0] != "SIP/2.0" {
       return
   }
   
   statusCode := parts[1]
   
   // Extract Call-ID
   var callID string
   for _, line := range lines {
       if strings.HasPrefix(line, "Call-ID:") {
           callID = strings.TrimSpace(strings.TrimPrefix(line, "Call-ID:"))
           break
       }
   }
   
   c.mu.Lock()
   call, exists := c.activeCalls[callID]
   c.mu.Unlock()
   
   if !exists {
       return
   }
   
   switch statusCode {
   case "100":
       log.Printf("[SIP] Call %s: Trying", callID)
       call.Status = "TRYING"
   case "180":
       log.Printf("[SIP] Call %s: Ringing", callID)
       call.Status = "RINGING"
   case "200":
       log.Printf("[SIP] Call %s: Answered", callID)
       call.Status = "ANSWERED"
       // Extract remote tag
       for _, line := range lines {
           if strings.HasPrefix(line, "To:") && strings.Contains(line, "tag=") {
               tagStart := strings.Index(line, "tag=") + 4
               tagEnd := strings.IndexAny(line[tagStart:], ";>")
               if tagEnd == -1 {
                   call.RemoteTag = line[tagStart:]
               } else {
                   call.RemoteTag = line[tagStart:tagStart+tagEnd]
               }
               break
           }
       }
   default:
       log.Printf("[SIP] Call %s: Status %s", callID, statusCode)
   }
}

func (c *Client) generateCallID() string {
   return fmt.Sprintf("%d@%s", time.Now().UnixNano(), c.localIP)
}

func (c *Client) generateTag() string {
   return fmt.Sprintf("%d", rand.Int63())
}

func (c *Client) generateBranch() string {
   return fmt.Sprintf("z9hG4bK%d", rand.Int63())
}

func (c *Client) GetActiveCallCount() int {
   c.mu.Lock()
   defer c.mu.Unlock()
   return len(c.activeCalls)
}

func (c *Client) Close() {
   if c.conn != nil {
       c.conn.Close()
   }
}
