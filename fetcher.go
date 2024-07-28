package smartapigo

import "net/http"

const (
	SCRIP_SEARCH_URL     = "rest/secure/angelbroking/order/v1/searchScrip"
	DAIL_INSTRUMENTS_URL = "https://margincalculator.angelbroking.com/OpenAPI_File/files/OpenAPIScripMaster.json"
)

type SearchScripPayload struct {
	Exchange    string `json:"exchange"`
	SearchScrip string `json:"searchscrip"`
}

type ScripResponse struct {
	Data []LTPParams `json:"data"`
}

type Instrument struct {
	Token          string `json:"token"`
	Symbol         string `json:"symbol"`
	Name           string `json:"name"`
	Expiry         string `json:"expiry"`
	Strike         string `json:"strike"`
	Lotsize        string `json:"lotsize"`
	InstrumentType string `json:"instrumenttype"`
	ExchangeSeg    string `json:"exch_seg"`
	TickSize       string `json:"tick_size"`
}

func (c *Client) SearchScrip(payload SearchScripPayload) ([]LTPParams, error) {
	var list []LTPParams
	params := structToMap(payload, "json")
	err := c.doEnvelope(http.MethodPost, SCRIP_SEARCH_URL, params, nil, &list, true)
	return list, err
}

func (c *Client) FetchDailyInstrumentsList() ([]Instrument, error) {
	var list []Instrument
	err := c.do(http.MethodGet, DAIL_INSTRUMENTS_URL, nil, nil, &list)

	return list, err
}
