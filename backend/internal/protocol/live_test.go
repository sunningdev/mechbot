package protocol

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Run with MECHBOT_LIVE=1 to check every configured Aave pool answers and list its first reserves.
func TestLiveAaveMarkets(t *testing.T) {
	if os.Getenv("MECHBOT_LIVE") == "" {
		t.Skip("set MECHBOT_LIVE=1 to hit public RPCs")
	}
	r := NewRegistry("", "", "")
	defer r.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	for _, m := range r.aave {
		err := r.rpc.Do(ctx, m.rpc, func(client *ethclient.Client) error {
			c, err := aaveAddresses(ctx, client, m)
			if err != nil {
				return err
			}
			var symbols []string
			for id := uint16(0); id < 4; id++ {
				out, err := callContract(ctx, client, aaveDetailABI, m.pool, "getReserveAddressById", id)
				if err != nil {
					return err
				}
				meta, err := token(ctx, client, m.chainID, out[0].(common.Address))
				if err != nil {
					return err
				}
				symbols = append(symbols, meta.symbol)
			}
			t.Logf("%s (%s) data provider %s, reserves %v", m.label(), m.key(), c.data.Hex(), symbols)
			return nil
		})
		if err != nil {
			t.Errorf("%s: %v", m.label(), err)
		}
	}
}

// Run with MECHBOT_LIVE_ADDRESS=0x... to read a real wallet through public RPCs.
func TestLiveFetch(t *testing.T) {
	addr := os.Getenv("MECHBOT_LIVE_ADDRESS")
	if addr == "" {
		t.Skip("set MECHBOT_LIVE_ADDRESS to read a real wallet")
	}
	r := NewRegistry("", "", "")
	defer r.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	for _, res := range r.Fetch(ctx, common.HexToAddress(addr)) {
		if res.Failure != nil {
			t.Logf("%s failed: %v", res.Failure.Label, res.Failure.Err)
			continue
		}
		p := res.Position
		if !p.HasDebt {
			continue
		}
		t.Logf("%s HF %.4f coll $%.2f debt $%.2f | coll %s @ $%.4f LT %.4f | debt %s @ $%.4f | APY %.2f%% | ETH liq $%.2f | %s",
			p.Label, p.HealthFactor, p.CollateralUSD, p.DebtUSD, p.Collateral.Symbol, p.Collateral.PriceUSD, p.CollateralLT,
			p.Debt.Symbol, p.Debt.PriceUSD, p.BorrowAPY*100, p.LiqETHPrice, p.AppURL)
		if rescue, ok := p.Rescue(1.2); ok {
			t.Logf("  rescue to %.2f: repay $%.2f (%.4f %s) or add $%.2f (%.4f %s)",
				rescue.Target, rescue.RepayUSD, rescue.RepayAmount, p.Debt.Symbol, rescue.AddUSD, rescue.AddAmount, p.Collateral.Symbol)
		}
	}
}
