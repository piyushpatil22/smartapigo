package websocket

import (
	"encoding/binary"
	"fmt"
	"math"
)

type RequestData struct {
	CorrelationID string `json:"correlationID"`
	Action        int    `json:"action"`
	Params        Params `json:"params"`
}

type Params struct {
	Mode      int        `json:"mode"`
	TokenList []TokenSet `json:"tokenList"`
}

type TokenSet struct {
	ExchangeType int      `json:"exchangeType"`
	Tokens       []string `json:"tokens"`
}

var (
	ErrInvalidExchangeType = fmt.Errorf("invalid exchangeType: Please check the exchange type and try again it support only 1 exchange type")
	ErrQuotaLimitExceeded  = fmt.Errorf("quota limit exceeded: You can subscribe to a maximum of %d tokens only", QUOTA_LIMIT)

	LITTLE_ENDIAN_BYTE_ORDER = binary.LittleEndian
)

const (
	SCALING_FACTOR = 100
)

type SubscriptionMode int

type ParsedData struct {
	SubscriptionMode             SubscriptionMode
	ExchangeType                 byte
	Token                        string
	SequenceNumber               int64
	ExchangeTimestamp            int64
	LastTradedPrice              float64
	LastTradedQuantity           int64
	AverageTradedPrice           float64
	VolumeTradeForTheDay         int64
	TotalBuyQuantity             float64
	TotalSellQuantity            float64
	OpenPriceOfTheDay            float64
	HighPriceOfTheDay            float64
	LowPriceOfTheDay             float64
	ClosedPrice                  float64
	LastTradedTimestamp          int64
	OpenInterest                 int64
	OpenInterestChangePercentage int64
	UpperCircuitLimit            int64
	LowerCircuitLimit            int64
	High52WeekPrice              float64
	Low52WeekPrice               float64
	Best5BuyData                 []OrderData
	Best5SellData                []OrderData
	Depth20BuyData               []DepthData
	Depth20SellData              []DepthData
	PacketReceivedTime           int64
}

type OrderData struct {
	Flag       uint16
	Quantity   int64
	Price      float64
	NoOfOrders uint16
}

type DepthData struct {
	Quantity    int32
	Price       float64
	NumOfOrders int16
}

func ParseBinaryData(binaryData []byte) (ParsedData, error) {
	var parsedData ParsedData

	parsedData.SubscriptionMode = SubscriptionMode(binaryData[0])
	parsedData.ExchangeType = binaryData[1]
	parsedData.Token = parseTokenValue(binaryData[2:27])
	parsedData.SequenceNumber = unpackData(binaryData, 27, 35, "q").(int64)
	parsedData.ExchangeTimestamp = unpackData(binaryData, 35, 43, "q").(int64)
	parsedData.LastTradedPrice = float64(unpackData(binaryData, 43, 51, "q").(int64)) / SCALING_FACTOR

	switch parsedData.SubscriptionMode {
	case QUOTE, SNAP_QUOTE:
		parsedData.LastTradedQuantity = unpackData(binaryData, 51, 59, "q").(int64)
		parsedData.AverageTradedPrice = float64(unpackData(binaryData, 59, 67, "q").(int64)) / SCALING_FACTOR
		parsedData.VolumeTradeForTheDay = unpackData(binaryData, 67, 75, "q").(int64)
		parsedData.TotalBuyQuantity = unpackData(binaryData, 75, 83, "d").(float64)
		parsedData.TotalSellQuantity = unpackData(binaryData, 83, 91, "d").(float64)
		parsedData.OpenPriceOfTheDay = float64(unpackData(binaryData, 91, 99, "q").(int64)) / SCALING_FACTOR
		parsedData.HighPriceOfTheDay = float64(unpackData(binaryData, 99, 107, "q").(int64)) / SCALING_FACTOR
		parsedData.LowPriceOfTheDay = float64(unpackData(binaryData, 107, 115, "q").(int64)) / SCALING_FACTOR
		parsedData.ClosedPrice = float64(unpackData(binaryData, 115, 123, "q").(int64)) / SCALING_FACTOR
	}

	if parsedData.SubscriptionMode == SNAP_QUOTE {
		parsedData.LastTradedTimestamp = unpackData(binaryData, 123, 131, "q").(int64)
		parsedData.OpenInterest = unpackData(binaryData, 131, 139, "q").(int64)
		parsedData.OpenInterestChangePercentage = unpackData(binaryData, 139, 147, "q").(int64)
		parsedData.UpperCircuitLimit = unpackData(binaryData, 347, 355, "q").(int64)
		parsedData.LowerCircuitLimit = unpackData(binaryData, 355, 363, "q").(int64)
		parsedData.High52WeekPrice = float64(unpackData(binaryData, 363, 371, "q").(int64)) / SCALING_FACTOR
		parsedData.Low52WeekPrice = float64(unpackData(binaryData, 371, 379, "q").(int64)) / SCALING_FACTOR
	}

	if parsedData.SubscriptionMode == DEPTH {
		parsedData.PacketReceivedTime = unpackData(binaryData, 35, 43, "q").(int64)
	}

	return parsedData, nil
}

func unpackData(binaryData []byte, start, end int, byteFormat string) interface{} {
	data := binaryData[start:end]
	switch byteFormat {
	case "B":
		return data[0]
	case "H":
		return LITTLE_ENDIAN_BYTE_ORDER.Uint16(data)
	case "I":
		return LITTLE_ENDIAN_BYTE_ORDER.Uint32(data)
	case "q":
		return int64(LITTLE_ENDIAN_BYTE_ORDER.Uint64(data))
	case "d":
		return math.Float64frombits(LITTLE_ENDIAN_BYTE_ORDER.Uint64(data))
	default:
		return nil
	}
}

func parseTokenValue(binaryPacket []byte) string {
	token := ""
	for i := 0; i < len(binaryPacket); i++ {
		if binaryPacket[i] == 0 {
			break
		}
		token += string(binaryPacket[i])
	}
	return token
}
