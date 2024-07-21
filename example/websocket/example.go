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


	tokenList := []websocket.TokenSet{
		{ExchangeType: 1, Tokens: []string{"5900"}},
	}
	newSocket.Subscribe("correlationID1", 3, tokenList)

	//TODO make connection stay longer (infinite time)
	time.Sleep(60 * time.Minute)
	newSocket.CloseConnection()

}

func GenerateTOTP(utf8string string) string {
	passcode, err := totp.GenerateCode(utf8string, time.Now())
	if err != nil {
		return ""
	}
	return passcode
}
