package smartapigo

import "net/http"

const (
	SCRIP_SEARCH_URL = "rest/secure/angelbroking/order/v1/searchScrip"
)

type SearchScripPayload struct {
	Exchange    string `json:"exchange"`
	SearchScrip string `json:"searchscrip"`
}

type ScripResponse struct {
	Data []LTPParams `json:"data"`
}

func (c *Client) SearchScrip(payload SearchScripPayload) ([]LTPParams, error) {
	var list []LTPParams
	params := structToMap(payload, "json")
	err := c.doEnvelope(http.MethodPost, SCRIP_SEARCH_URL, params, nil, &list, true)
	return list, err
}
