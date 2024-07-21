package websocket

import (
	"crypto/tls"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/piyushpatil22/smartapigo"
	"github.com/sirupsen/logrus"
)

const (
	ROOT_URI            = "wss://smartapisocket.angelone.in/smart-stream"
	HEART_BEAT_MESSAGE  = "ping"
	HEART_BEAT_INTERVAL = 10
	RESUBSCRIBE_FLAG    = false

	SUBSCRIBE_ACTION   = 1
	UNSUBSCRIBE_ACTION = 0

	QUOTA_LIMIT = 50

	LTP_MODE   = 1
	QUOTE      = 2
	SNAP_QUOTE = 3
	DEPTH      = 4

	NSE_CM = 1
	NSE_FO = 2
	BSE_CM = 3
	BSE_FO = 4
	MCX_FO = 5
	NCX_FO = 7
	CDE_FO = 13

	DEFAULT_MAX_RETRY_ATTEMPTS = 100
	DEFAULT_RETRY_STRATERGY    = 0
	DEFAULT_RETRY_DELAY        = 10
	DEFAULT_RETRY_MULTPLIER    = 2
	DEFAULT_RETRY_DURATION     = 60
)

var (
	SUBSCRIPTION_MODE_MAP = map[int]string{
		1: "LTP",
		2: "QUOTE",
		3: "SNAP_QUOTE",
		4: "DEPTH",
	}
)

type SocketClientV2 struct {
	Auth_token        string
	Api_key           string
	Client_code       string
	Feed_token        string
	disconnectFlag    bool
	LastPongTimestamp time.Time
	inputRequestMap   map[int]map[int][]string
	retryParams       RetryParams
	resubscribeFlag   bool
	logger            *logrus.Logger
	wsConn            *websocket.Conn
	mutex             sync.Mutex
}

type RetryParams struct {
	MaxRetryAttempt int
	CurrentAttempt  int
	RetryStrategy   int
	RetryDelay      int
	RetryMultiplier int
	RetryDuration   int
}

// Create a new ticker instance with latest supported websocket streaming functionality
func NewSocketConnV2(auth_token, client_code, api_key, feed_token string, retryParam RetryParams) *SocketClientV2 {
	sw := &SocketClientV2{
		Auth_token:      auth_token,
		Client_code:     client_code,
		Api_key:         api_key,
		Feed_token:      feed_token,
		retryParams:     retryParam,
		logger:          logrus.New(),
		inputRequestMap: make(map[int]map[int][]string),
	}
	sw.logger.SetLevel(logrus.InfoLevel)

	sw.logger.SetFormatter(&smartapigo.CustomFormatter{})
	return sw
}

func (s *SocketClientV2) Connect() {

	headers := map[string][]string{
		"Authorization": {s.Auth_token},
		"x-api-key":     {s.Api_key},
		"x-client-code": {s.Client_code},
		"x-feed-token":  {s.Feed_token},
	}

	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, _, err := dialer.Dial(ROOT_URI, headers)
	if err != nil {
		s.logger.Errorf("Error connecting to websocket")
	}

	s.wsConn = conn

	s.logger.Info("Connected to websocket!!")

	s.wsConn.SetPongHandler(func(appData string) error {
		s.onPong(appData)
		return nil
	})

	s.wsConn.SetPingHandler(func(appData string) error {
		s.onPing(appData)
		return nil
	})

	go s.runHeartbeat()
	go s.readMessages()
}

func (sw *SocketClientV2) onPong(appData string) {
	sw.LastPongTimestamp = time.Now()
	sw.logger.Infof("Received pong: %s", appData)
}
func (sw *SocketClientV2) onPing(appData string) {
	sw.LastPongTimestamp = time.Now()
	sw.logger.Infof("Received ping: %s", appData)
}

func (sw *SocketClientV2) runHeartbeat() {
	ticker := time.NewTicker(HEART_BEAT_INTERVAL * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sw.mutex.Lock()
		if sw.disconnectFlag {
			sw.mutex.Unlock()
			return
		}
		sw.wsConn.WriteMessage(websocket.TextMessage, []byte(HEART_BEAT_MESSAGE))
		sw.logger.Info("Sent heartbeat")
		sw.mutex.Unlock()
	}
}

func (sw *SocketClientV2) readMessages() {
	for {
		_, message, err := sw.wsConn.ReadMessage()
		if err != nil {
			sw.logger.Errorf("Read message error: %v", err)
			sw.handleReconnect()
			return
		}
		sw.logger.Infof("Received message: %s", message)
		sw.handleMessage(message)
	}
}

func (sw *SocketClientV2) handleMessage(message []byte) {
	if string(message) == "pong" {
		sw.onPong("pong")
		return
	}

	if string(message) != "pong" {
		data, err := parseBinaryData(message)
		if err != nil {
			sw.logger.Error(err)
		} else {
			sw.logger.Infof("%+v", data)
		}
	}
	// Process other messages (binary data, etc.)
}

func (sw *SocketClientV2) handleReconnect() {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()
	sw.disconnectFlag = true
	sw.wsConn.Close()

	if sw.retryParams.CurrentAttempt < sw.retryParams.MaxRetryAttempt {
		sw.retryParams.CurrentAttempt++
		sw.logger.Warnf("Reconnecting (Attempt %d)...", sw.retryParams.CurrentAttempt)
		time.Sleep(time.Duration(sw.retryParams.RetryDelay) * time.Second)
		sw.Connect()
	} else {
		sw.logger.Warn("Max retry attempts reached, closing connection.")
	}
}

func (sw *SocketClientV2) Subscribe(correlationID string, mode int, tokenList []TokenSet) {
	request := RequestData{
		CorrelationID: correlationID,
		Action:        1,
		Params: Params{
			Mode:      mode,
			TokenList: tokenList,
		},
	}
	data, err := json.Marshal(request)
	if err != nil {
		sw.logger.Errorf("Error marshaling subscribe request: %v", err)
		return
	}

	sw.mutex.Lock()
	if _, exists := sw.inputRequestMap[mode]; !exists {
		sw.inputRequestMap[mode] = make(map[int][]string)
	}

	for _, token := range tokenList {
		sw.inputRequestMap[mode][token.ExchangeType] = append(sw.inputRequestMap[mode][token.ExchangeType], token.Tokens...)
	}
	sw.mutex.Unlock()

	sw.wsConn.WriteMessage(websocket.TextMessage, data)
	sw.resubscribeFlag = true
}

func (sw *SocketClientV2) Unsubscribe(correlationID string, mode int, tokenList []TokenSet) {
	request := RequestData{
		CorrelationID: correlationID,
		Action:        0,
		Params: Params{
			Mode:      mode,
			TokenList: tokenList,
		},
	}
	data, err := json.Marshal(request)
	if err != nil {
		sw.logger.Errorf("Error marshaling unsubscribe request: %v", err)
		return
	}

	sw.mutex.Lock()
	sw.inputRequestMap[mode] = nil
	sw.mutex.Unlock()

	sw.wsConn.WriteMessage(websocket.TextMessage, data)
	sw.resubscribeFlag = true
}

func (sw *SocketClientV2) CloseConnection() {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()
	sw.resubscribeFlag = false
	sw.disconnectFlag = true
	if sw.wsConn != nil {
		sw.wsConn.Close()
	}
}
