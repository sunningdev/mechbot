package protocol

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

const morphoEndpoint = "https://api.morpho.org/graphql"

const morphoQuery = `query ($chainId: Int!, $address: String!) {
  userByAddress(chainId: $chainId, address: $address) {
    marketPositions {
      market {
        marketId
        lltv
        collateralAsset { symbol decimals }
        loanAsset { symbol decimals }
        state { borrowApy }
      }
      state { borrowAssets borrowAssetsUsd collateral collateralUsd }
    }
  }
}`

type morphoChain struct {
	chainID int
	chain   string
}

func (m morphoChain) label() string { return "Morpho Blue · " + m.chain }

type morphoAsset struct {
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

// flexNumber accepts big integers sent either as JSON numbers or as strings.
type flexNumber float64

func (n *flexNumber) UnmarshalJSON(raw []byte) error {
	s := strings.Trim(string(raw), `"`)
	if s == "" || s == "null" {
		*n = 0
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*n = flexNumber(f)
	return nil
}

type morphoResponse struct {
	Data struct {
		UserByAddress *struct {
			MarketPositions []struct {
				Market struct {
					MarketID        string       `json:"marketId"`
					LLTV            string       `json:"lltv"`
					CollateralAsset *morphoAsset `json:"collateralAsset"`
					LoanAsset       *morphoAsset `json:"loanAsset"`
					State           *struct {
						BorrowApy *float64 `json:"borrowApy"`
					} `json:"state"`
				} `json:"market"`
				State struct {
					BorrowAssets    flexNumber `json:"borrowAssets"`
					BorrowAssetsUsd *float64   `json:"borrowAssetsUsd"`
					Collateral      flexNumber `json:"collateral"`
					CollateralUsd   *float64   `json:"collateralUsd"`
				} `json:"state"`
			} `json:"marketPositions"`
		} `json:"userByAddress"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func fetchMorpho(ctx context.Context, client *http.Client, chain morphoChain, user string, ethUSD float64) ([]Position, error) {
	body, err := json.Marshal(map[string]any{
		"query": morphoQuery,
		"variables": map[string]any{
			"chainId": chain.chainID,
			"address": user,
		},
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, morphoEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("morpho status %d", resp.StatusCode)
	}
	var parsed morphoResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Errors) > 0 {
		return nil, fmt.Errorf("morpho: %s", parsed.Errors[0].Message)
	}
	if parsed.Data.UserByAddress == nil {
		return nil, nil
	}

	var out []Position
	for _, item := range parsed.Data.UserByAddress.MarketPositions {
		if item.State.BorrowAssetsUsd == nil || *item.State.BorrowAssetsUsd <= 0 {
			continue
		}
		debt := *item.State.BorrowAssetsUsd
		collateral := 0.0
		if item.State.CollateralUsd != nil {
			collateral = *item.State.CollateralUsd
		}
		lltv, err := parseLLTV(item.Market.LLTV)
		if err != nil {
			return nil, err
		}
		hf := 0.0
		if debt > 0 {
			hf = collateral * lltv / debt
		}
		coll := morphoAsset{Symbol: "?"}
		loan := morphoAsset{Symbol: "?"}
		if a := item.Market.CollateralAsset; a != nil && a.Symbol != "" {
			coll = *a
		}
		if a := item.Market.LoanAsset; a != nil && a.Symbol != "" {
			loan = *a
		}
		var exp Exposure
		exp.addColl(coll.Symbol, collateral*lltv)
		exp.addDebt(loan.Symbol, debt)
		pos := Position{
			Key:           fmt.Sprintf("morpho-blue:%d:%s", chain.chainID, item.Market.MarketID),
			Label:         fmt.Sprintf("Morpho Blue · %s %s/%s", chain.chain, coll.Symbol, loan.Symbol),
			CollateralUSD: collateral,
			DebtUSD:       debt,
			HealthFactor:  hf,
			HasDebt:       true,
			ETHPrice:      ethUSD,
			LiqETHPrice:   exp.LiqETHPrice(ethUSD),
			Collateral:    Asset{Symbol: coll.Symbol, PriceUSD: unitPrice(collateral, float64(item.State.Collateral), coll.Decimals)},
			Debt:          Asset{Symbol: loan.Symbol, PriceUSD: unitPrice(debt, float64(item.State.BorrowAssets), loan.Decimals)},
			CollateralLT:  lltv,
			AppURL:        morphoAppURL(chain.chainID, item.Market.MarketID),
			Exposure:      exp,
		}
		if s := item.Market.State; s != nil && s.BorrowApy != nil && *s.BorrowApy > 0 {
			pos.BorrowAPY = *s.BorrowApy
		}
		out = append(out, pos)
	}
	return out, nil
}

// unitPrice derives a token's USD price from a USD value and its raw on-chain amount.
func unitPrice(usd, raw float64, decimals int) float64 {
	if usd <= 0 || raw <= 0 || decimals < 0 || decimals > 36 {
		return 0
	}
	return usd / (raw / math.Pow10(decimals))
}

func parseLLTV(raw string) (float64, error) {
	n, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return 0, fmt.Errorf("lltv 无法解析: %s", raw)
	}
	f := new(big.Float).SetPrec(256).SetInt(n)
	f.Quo(f, new(big.Float).SetPrec(256).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))
	out, _ := f.Float64()
	if out <= 0 || out > 1 {
		return 0, fmt.Errorf("lltv 超出范围: %s", strconv.FormatFloat(out, 'f', -1, 64))
	}
	return out, nil
}
