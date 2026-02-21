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

	// Mapping:
	// api.gold_bar.buy (77,700) -> GoldBarSell (Shop sells to customer)
	// api.gold_bar.sell (77,600) -> GoldBarBuy (Shop buys from customer)
	// api.gold.buy (78,500) -> GoldOrnamentSell (Shop sells to customer)
	// api.gold.sell (76,042.56) -> GoldOrnamentBuy (Shop buys from customer)

	return &gold_price.GoldPriceData{
		GoldBarBuy:       parsePrice(apiResp.Response.Price.GoldBar.Sell),
		GoldBarSell:      parsePrice(apiResp.Response.Price.GoldBar.Buy),
		GoldOrnamentBuy:  parsePrice(apiResp.Response.Price.Gold.Sell),
		GoldOrnamentSell: parsePrice(apiResp.Response.Price.Gold.Buy),
	}, nil
}
