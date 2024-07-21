# The Smart API Go client

The official Go client for communicating with the Angel Broking Smart APIs.

SmartAPI is a set of REST-like APIs that expose many capabilities required to build a complete investment and trading platform. Execute orders in real time, manage user portfolio, stream live market data (WebSockets), and more, with the simple HTTP API collection.

## Installation

```
go get github.com/piyushpatil22/smartapigo
```

## API usage

```golang
package main

import (
	"fmt"
	SmartApi "github.com/piyushpatil22/smartapigo"
)

func main() {

	apiKey := "{your-api-key}"
	clientCode := "{your-client-code}"
	PIN := "{your-pin}"
	// Create New Angel Broking Client
	ABClient := SmartApi.New(clientCode, PIN, apiKey)

	QRCodeSecret := "{your-totp-qr-code-key}"

	// Generate TOTP using the QR Key
	opt := GeneratePassCode(QRCodeSecret)

	// User Login and Generate User Session
	session, err := ABClient.GenerateSession(otp)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//Renew User Tokens using refresh token
	session.UserSessionTokens, err = ABClient.RenewAccessToken(session.RefreshToken)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("User Session Tokens :- ", session.UserSessionTokens)

	//Get User Profile
	session.UserProfile, err = ABClient.GetUserProfile()

	if err != nil {
		fmt.Println(err.Error())
		return
	}


	//Place Order
	order, err := ABClient.PlaceOrder(SmartApi.OrderParams{Variety: "NORMAL", TradingSymbol: "SBIN-EQ", SymbolToken: "3045", TransactionType: "BUY", Exchange: "NSE", OrderType: "LIMIT", ProductType: "INTRADAY", Duration: "DAY", Price: "19500", SquareOff: "0", StopLoss: "0", Quantity: "1"})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Placed Order ID and Script :- ", order)
}
```

## Websocket Data Streaming

```golang
package main

import (
	"fmt"
	SmartApi "github.com/piyushpatil22/smartapigo"
	"github.com/piyushpatil22/smartapigo/websocket"
	"time"
)

var socketClient *websocket.SocketClient


func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	// You will need a .env file with below fields with appropriate values
	QRCodeSecret := os.Getenv("QR_CODE")
	apiKey := os.Getenv("API_KEY")
	PIN := os.Getenv("PIN")
	clientCode := os.Getenv("CLIENT_CODE")

	// Generate TOTP using the QR Key
	opt := GenerateTOTP(QRCodeSecret)

	// User Login and Generate User Session
	session, err := ABClient.GenerateSession(otp)

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

	// Don't worry about this, this will be abstracted in coming changes, as this is not required at this
	// level
	retryParams := websocket.RetryParams{
		CurrentAttempt:  0,
		MaxRetryAttempt: 3,
		RetryDelay:      10,
		RetryDuration:   60,
	}

	newSocket := websocket.NewSocketConnV2(session.AccessToken, session.ClientCode, apiKey, session.FeedToken, retryParams)

	// Mouting user defined funcs is currently not supported. Work on that is in progress
	newSocket.Connect()

	// For subscribing you need ExchangeType and the token ID of the instrument you want to receive data
	// of. You can send different types of Exchanges and multiple tokens at once. The current token limit
	// is 50. A max of 50 tokens can be subscribed at once
	tokenList := []websocket.TokenSet{
		{ExchangeType: 1, Tokens: []string{"5900"}},
	}
	newSocket.Subscribe("correlationID1", 3, tokenList)

	// As of now this will insta close the connection, will be kepping this infinite
	// work in progress on that
	newSocket.CloseConnection()

}
```

## Examples

Check example folder for more examples.

You can run the following after updating the Credentials in the examples:

```
go run example/example.go
```

For websocket example

```
go run example/websocket/example.go
```

## Run unit tests

```
go test -v

#running tests might fail
```
