package gold_price

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devper-gold/gold-shop-api/app/feature/gold_price"
)

type ThaiGoldAPIClient struct {
	url string
}

func NewThaiGoldAPIClient(url string) *ThaiGoldAPIClient {
	return &ThaiGoldAPIClient{
		url: url,
	}
}

type thaiGoldResponse struct {
	Status   string `json:"status"`
	Response struct {
		Date       string `json:"date"`
		UpdateTime string `json:"update_time"`
		Price      struct {
			Gold struct {
				Buy  string `json:"buy"`
				Sell string `json:"sell"`
			} `json:"gold"`
			GoldBar struct {
				Buy  string `json:"buy"`
				Sell string `json:"sell"`
			} `json:"gold_bar"`
		} `json:"price"`
	} `json:"response"`
}

func (c *ThaiGoldAPIClient) GetCurrentPrice(ctx context.Context) (*gold_price.GoldPriceData, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp thaiGoldResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	// Parse strings to float64, removing commas
	parsePrice := func(s string) float64 {
		s = strings.ReplaceAll(s, ",", "")
		val, _ := strconv.ParseFloat(s, 64)
		return val
	}

	return &gold_price.GoldPriceData{
		GoldBarBuy:       parsePrice(apiResp.Response.Price.GoldBar.Buy),
		GoldBarSell:      parsePrice(apiResp.Response.Price.GoldBar.Sell),
		GoldOrnamentBuy:  parsePrice(apiResp.Response.Price.Gold.Buy),
		GoldOrnamentSell: parsePrice(apiResp.Response.Price.Gold.Sell),
	}, nil
}
