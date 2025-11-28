// Filename: internal/ws/handler.go
package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
	//"unicode"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 5 * time.Second
	idleReadLimit  = 30 * time.Second
	maxMessageSize = 1024 * 4
)

var (
	// Allow multiple development origins
	allowedOrigins = []string{
		"http://localhost:4000",
		"http://localhost:8080",
		"http://127.0.0.1:4000",
		"http://127.0.0.1:8080",
	}
	messageCounter uint64
)


type CommandRequest struct {
	Command string  `json:"command"`
	A       float64 `json:"a"`
	B       float64 `json:"b"`
}

type CommandResponse struct {
	Result  float64 `json:"result"`
	Command string  `json:"command"`
	Error   string  `json:"error,omitempty"`
}

func originAllowed(o string) bool {
	if o == "" {
		return false
	}
	// Development: allow localhost and 127.0.0.1 on any port
	if strings.Contains(o, "localhost") || strings.Contains(o, "127.0.0.1") {
		return true
	}
	
	// Specific origin check
	for _, a := range allowedOrigins {
		if strings.EqualFold(o, a) {
			return true
		}
	}
	return false
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		ok := originAllowed(origin)
		if !ok {
			log.Printf("blocked cross-origin websocket: Origin=%q Path=%s", origin, r.URL.Path)
		} else {
			log.Printf("allowed cross-origin websocket: Origin=%q", origin)
		}
		return ok
	},
}


func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}


func processCommand(payload []byte) ([]byte, error) {
	var req CommandRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}

	var resp CommandResponse
	resp.Command = req.Command

	switch req.Command {
	case "add":
		resp.Result = req.A + req.B
	case "subtract":
		resp.Result = req.A - req.B
	case "multiply":
		resp.Result = req.A * req.B
	case "divide":
		if req.B == 0 {
			resp.Error = "division by zero"
		} else {
			resp.Result = req.A / req.B
		}
	default:
		resp.Error = fmt.Sprintf("unknown command: %s", req.Command)
	}

	return json.Marshal(resp)
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}
	defer conn.Close()

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(idleReadLimit))

	for {
		msgType, payload, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("read error: %v", err)
			}
			return
		}

		conn.SetReadDeadline(time.Now().Add(idleReadLimit))

		// Echo back text messages
		if msgType == websocket.TextMessage {
			response := payload
			message := string(payload)

			
			if len(payload) > 0 && payload[0] == '{' {
				jsonResponse, err := processCommand(payload)
				if err != nil {
					log.Printf("JSON processing error: %v", err)
					errorMsg := fmt.Sprintf(`{"error": "%v"}`, err)
					response = []byte(errorMsg)
				} else {
					response = jsonResponse
				}
			} else {
				
				if strings.HasPrefix(message, "UPPER:") {
					text := strings.TrimPrefix(message, "UPPER:")
					response = []byte(strings.ToUpper(text))
				}

				
				if strings.HasPrefix(message, "REVERSE:") {
					text := strings.TrimPrefix(message, "REVERSE:")
					response = []byte(reverseString(text))
				}

				
				if len(payload) > 0 && payload[0] != '{' {
					count := atomic.AddUint64(&messageCounter, 1)
					response = []byte(fmt.Sprintf("[Msg #%d] %s", count, response))
				}
			}

			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.TextMessage, response); err != nil {
				log.Printf("write error: %v", err)
				return
			}
		}
	}
}