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
	callbacks         callbacksV2
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

type callbacksV2 struct {
	onMessage     func([]byte)
	onNoReconnect func(int)
	onReconnect   func(int, time.Duration)
	onConnect     func()
	onClose       func(int, string)
	onError       func(error)
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
		s.triggerError(err)
	}

	s.wsConn = conn

	defer s.wsConn.Close()

	s.triggerConnect()

	s.logger.Info("Connected to websocket!!")

	s.wsConn.SetCloseHandler(s.handleClose)

	s.wsConn.SetPongHandler(func(appData string) error {
		s.onPong(appData)
		return nil
	})

	s.wsConn.SetPingHandler(func(appData string) error {
		s.onPing(appData)
		return nil
	})

	go s.runHeartbeat()
}

func (sw *SocketClientV2) onPong(appData string) {
	sw.LastPongTimestamp = time.Now()
	sw.logger.Infof("Received pong: %s", appData)
}
func (sw *SocketClientV2) onPing(appData string) {
	sw.LastPongTimestamp = time.Now()
	sw.logger.Infof("Received ping: %s", appData)
}

func (sw *SocketClientV2) Serve() {
	for {
		//reconnect logic
		sw.mutex.Lock()

		if sw.disconnectFlag {
			sw.mutex.Unlock()
			return
		}
		//need to handle reconnect

		sw.readMessages()
	}
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
		go sw.triggerMessage(message)
		sw.handleMessage(message)
	}
}

func (sw *SocketClientV2) handleMessage(message []byte) {
	if string(message) == "pong" {
		sw.onPong("pong")
		return
	}

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

// Trigger callback methods
func (s *SocketClientV2) triggerError(err error) {
	if s.callbacks.onError != nil {
		s.callbacks.onError(err)
	}
}

func (s *SocketClientV2) triggerClose(code int, reason string) {
	if s.callbacks.onClose != nil {
		s.callbacks.onClose(code, reason)
	}
}

func (s *SocketClientV2) triggerConnect() {
	if s.callbacks.onConnect != nil {
		s.callbacks.onConnect()
	}
}

func (s *SocketClientV2) triggerReconnect(attempt int, delay time.Duration) {
	if s.callbacks.onReconnect != nil {
		s.callbacks.onReconnect(attempt, delay)
	}
}

func (s *SocketClientV2) triggerNoReconnect(attempt int) {
	if s.callbacks.onNoReconnect != nil {
		s.callbacks.onNoReconnect(attempt)
	}
}

func (s *SocketClientV2) triggerMessage(message []byte) {
	if s.callbacks.onMessage != nil {
		s.callbacks.onMessage(message)
	}
}

func (s *SocketClientV2) handleClose(code int, reason string) error {
	s.triggerClose(code, reason)
	return nil
}

// OnConnect callback.
func (s *SocketClientV2) OnConnect(f func()) {
	s.callbacks.onConnect = f
}

// OnError callback.
func (s *SocketClientV2) OnError(f func(err error)) {
	s.callbacks.onError = f
}

// OnClose callback.
func (s *SocketClientV2) OnClose(f func(code int, reason string)) {
	s.callbacks.onClose = f
}

// OnMessage callback.
func (s *SocketClientV2) OnMessage(f func(message []byte)) {
	s.callbacks.onMessage = f
}

// OnReconnect callback.
func (s *SocketClientV2) OnReconnect(f func(attempt int, delay time.Duration)) {
	s.callbacks.onReconnect = f
}

// OnNoReconnect callback.
func (s *SocketClientV2) OnNoReconnect(f func(attempt int)) {
	s.callbacks.onNoReconnect = f
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
