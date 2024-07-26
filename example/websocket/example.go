package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	SmartApi "github.com/piyushpatil22/smartapigo"
	"github.com/piyushpatil22/smartapigo/websocket"
	"github.com/pquerna/otp/totp"
)

var dataChannel = make(chan []byte, 100)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	// You will need a .env file with below fields with appropriate values
	QRCodeSecret := os.Getenv("QR_CODE")
	apiKey := os.Getenv("API_KEY")
	PIN := os.Getenv("PIN")
	clientCode := os.Getenv("CLIENT_CODE")

	// Generate TOTP using the secret
	opt := GenerateTOTP(QRCodeSecret)

	ABClient := SmartApi.New(clientCode, PIN, apiKey)

	// User Login and Generate User Session
	session, err := ABClient.GenerateSession(opt)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//Get User Profile
	session.UserProfile, err = ABClient.GetUserProfile()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	retryParams := websocket.RetryParams{
		CurrentAttempt:  0,
		MaxRetryAttempt: 3,
		RetryDelay:      10,
		RetryDuration:   60,
	}

	newSocket := websocket.NewSocketConnV2(session.AccessToken, session.ClientCode, apiKey, session.FeedToken, retryParams)

	newSocket.Connect()
	newSocket.OnMessage(onMessage)

	tokenList := []websocket.TokenSet{
		{ExchangeType: 1, Tokens: []string{"5900"}},
	}
	newSocket.Subscribe("correlationID1", 1, tokenList)
	
	go processData()

	newSocket.Serve()
}

func GenerateTOTP(utf8string string) string {
	passcode, err := totp.GenerateCode(utf8string, time.Now())
	if err != nil {
		return ""
	}
	return passcode
}

func onMessage(message []byte) {
	dataChannel <- message
}

func processData() {
	token_106298 := 0
	token_106299 := 0
	token_106300 := 0
	token_5900 := 0
	for data := range dataChannel {
		parsDT, err := websocket.ParseBinaryData(data)
		if err != nil {
			log.Println(err)
		}
		if parsDT.Token == "106298" {
			token_106298++
		}
		if parsDT.Token == "5900" {
			token_5900++
		}
		if parsDT.Token == "106299" {
			token_106299++
		}
		if parsDT.Token == "106300" {
			token_106300++
		}
		log.Printf("Counts 5900: %d 298: %d 299: %d 300: %d", token_5900, token_106298, token_106299, token_106300)
	}
}
