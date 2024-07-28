package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	SmartApi "github.com/piyushpatil22/smartapigo"
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
	opt := SmartApi.GenerateTOTP(QRCodeSecret)

	ABClient := SmartApi.New(clientCode, PIN, apiKey)

	// User Login and Generate User Session
	_, err := ABClient.GenerateSession(opt)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	list, err := ABClient.FetchDailyInstrumentsList()
	if err != nil {
		log.Println(err)
		return
	}
	for _, item := range list {
		log.Println(item)
	}
	// searchScrip := SmartApi.SearchScripPayload{
	// 	Exchange:    SmartApi.NFO,
	// 	SearchScrip: "SUNPHARMA",
	// }
	// res, err := ABClient.SearchScrip(searchScrip)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// for _, item := range res {
	// 	log.Println(item)
	// }
	// //Place Order
	// order, err := ABClient.PlaceOrder(SmartApi.OrderParams{Variety: "NORMAL", TradingSymbol: "SBIN-EQ", SymbolToken: "3045", TransactionType: "BUY", Exchange: "NSE", OrderType: "LIMIT", ProductType: "INTRADAY", Duration: "DAY", Price: "19500", SquareOff: "0", StopLoss: "0", Quantity: "1"})

	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	// //Modify Order
	// modifiedOrder, err := ABClient.ModifyOrder(SmartApi.ModifyOrderParams{Variety: "NORMAL", OrderID: order.OrderID, OrderType: "LIMIT", ProductType: "INTRADAY", Duration: "DAY", Price: "19400", Quantity: "1", TradingSymbol: "SBI-EQ", SymbolToken: "3045", Exchange: "NSE"})

	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	// fmt.Println("Modified Order ID :- ", modifiedOrder)

	// //Cancel Order
	// cancelledOrder, err := ABClient.CancelOrder("NORMAL", modifiedOrder.OrderID)

	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	// fmt.Println("Cancelled Order ID :- ", cancelledOrder)

	// //Get Holdings
	// holdings, err := ABClient.GetHoldings()

	// if err != nil {
	// 	fmt.Println(err.Error())
	// } else {

	// 	fmt.Println("Holdings :- ", holdings)
	// }

	// //Get Positions
	// positions, err := ABClient.GetPositions()

	// if err != nil {
	// 	fmt.Println(err.Error())
	// } else {

	// 	fmt.Println("Positions :- ", positions)
	// }

	// //Get TradeBook
	// trades, err := ABClient.GetTradeBook()

	// if err != nil {
	// 	fmt.Println(err.Error())
	// } else {

	// 	fmt.Println("All Trades :- ", trades)
	// }

	// //Get Last Traded Price
	// ltp, err := ABClient.GetLTP(SmartApi.LTPParams{Exchange: "NSE", TradingSymbol: "SBIN-EQ", SymbolToken: "3045"})

	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	// fmt.Println("Last Traded Price :- ", ltp)

	// //Get Risk Management System
	// rms, err := ABClient.GetRMS()

	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	// fmt.Println("Risk Managemanet System :- ", rms)

	// //Position Conversion
	// err = ABClient.ConvertPosition(SmartApi.ConvertPositionParams{"NSE", "SBIN-EQ", "INTRADAY", "MARGIN", "BUY", 1, "DAY"})
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	// fmt.Println("Position Conversion Successful")
}
