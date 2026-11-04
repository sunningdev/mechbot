package liqmap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"mechbot/internal/protocol"
)

const (
	morphoEndpoint = "https://api.morpho.org/graphql"
	morphoPage     = 1000
	morphoMaxPages = 30
	morphoGap      = time.Second
)

// Only Ethereum positions between liquidation and maxHealth; the rest can't land on the map.
const morphoMapQuery = `query ($skip: Int!, $hfMax: Float!) {
  marketPositions(first: 1000, skip: $skip, orderBy: BorrowShares, orderDirection: Desc,
    where: { chainId_in: [1], borrowShares_gte: "1", healthFactor_gte: 1, healthFactor_lte: $hfMax }) {
    items {
      user { address }
      market { lltv collateralAsset { symbol } loanAsset { symbol } }
      state { borrowAssetsUsd collateralUsd }
    }
    pageInfo { countTotal }
  }
}`

type morphoMapResponse struct {
	Data struct {
		MarketPositions struct {
			Items []struct {
				User struct {
					Address string `json:"address"`
				} `json:"user"`
				Market struct {
					LLTV            json.Number `json:"lltv"`
					CollateralAsset *struct {
						Symbol string `json:"symbol"`
					} `json:"collateralAsset"`
					LoanAsset struct {
						Symbol string `json:"symbol"`
					} `json:"loanAsset"`
				} `json:"market"`
				State *struct {
					BorrowAssetsUSD *float64 `json:"borrowAssetsUsd"`
					CollateralUSD   *float64 `json:"collateralUsd"`
				} `json:"state"`
			} `json:"items"`
			PageInfo struct {
				CountTotal int `json:"countTotal"`
			} `json:"pageInfo"`
		} `json:"marketPositions"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (s *Service) refreshMorpho(ctx context.Context, ethUSD float64) ([]Item, error) {
	var items []Item
	for page := range morphoMaxPages {
		if page > 0 {
			if err := sleep(ctx, morphoGap); err != nil {
				return nil, err
			}
		}
		resp, err := s.morphoPage(ctx, page*morphoPage)
		if err != nil {
			return nil, err
		}
		mp := resp.Data.MarketPositions
		for _, p := range mp.Items {
			if p.State == nil || p.State.BorrowAssetsUSD == nil || p.State.CollateralUSD == nil || p.Market.CollateralAsset == nil {
				continue
			}
			debt, coll := *p.State.BorrowAssetsUSD, *p.State.CollateralUSD
			if debt < minDebtUSD {
				continue
			}
			lltv, err := strconv.ParseFloat(p.Market.LLTV.String(), 64)
			if err != nil || lltv <= 0 {
				continue
			}
			lltv /= 1e18
			collSym, loanSym := p.Market.CollateralAsset.Symbol, p.Market.LoanAsset.Symbol
			var exp protocol.Exposure
			if protocol.IsETHSymbol(collSym) {
				exp.ETHColl = coll * lltv
			} else {
				exp.OtherColl = coll * lltv
			}
			if protocol.IsETHSymbol(loanSym) {
				exp.ETHDebt = debt
			} else {
				exp.OtherDebt = debt
			}
			liq := exp.LiqETHPrice(ethUSD)
			if liq <= 0 || liq >= ethUSD {
				continue
			}
			items = append(items, Item{
				Protocol: "morpho",
				Market:   "Morpho " + collSym + "/" + loanSym,
				Address:  p.User.Address,
				Liq:      liq,
				CollUSD:  coll * liq / ethUSD,
				DebtUSD:  debt,
			})
		}
		if len(mp.Items) < morphoPage || (page+1)*morphoPage >= mp.PageInfo.CountTotal {
			break
		}
	}
	return items, nil
}

func (s *Service) morphoPage(ctx context.Context, skip int) (morphoMapResponse, error) {
	var out morphoMapResponse
	body, _ := json.Marshal(map[string]any{
		"query":     morphoMapQuery,
		"variables": map[string]any{"skip": skip, "hfMax": maxHealth},
	})
	var lastErr error
	for _, wait := range retryWaits {
		if err := sleep(ctx, wait); err != nil {
			return out, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, morphoEndpoint, bytes.NewReader(body))
		if err != nil {
			return out, err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := s.http.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		out = morphoMapResponse{}
		err = json.NewDecoder(resp.Body).Decode(&out)
		resp.Body.Close()
		switch {
		case resp.StatusCode != http.StatusOK:
			lastErr = fmt.Errorf("Morpho API HTTP %d", resp.StatusCode)
		case err != nil:
			lastErr = err
		case len(out.Errors) > 0:
			return out, fmt.Errorf("Morpho API: %s", out.Errors[0].Message)
		default:
			return out, nil
		}
	}
	return out, lastErr
}
