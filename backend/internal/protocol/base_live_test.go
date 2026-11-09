package protocol

import (
	"context"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Run with MECHBOT_LIVE=1: finds recent borrowers on each Base Comet and reads their positions,
// checking the market config (and the WETH market's ETH quoting) against real accounts.
func TestLiveBaseComets(t *testing.T) {
	if os.Getenv("MECHBOT_LIVE") == "" {
		t.Skip("set MECHBOT_LIVE=1 to hit public RPCs")
	}
	r := NewRegistry("", "", "", "")
	defer r.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	withdraw := crypto.Keccak256Hash([]byte("Withdraw(address,address,uint256)"))
	parsed := mustABI(cometABIJSON)
	ethUSD := r.ethPrice(ctx)
	for _, m := range r.comets {
		if m.chainID != 8453 {
			continue
		}
		found := 0
		err := r.rpc.Do(ctx, m.rpc, func(client *ethclient.Client) error {
			head, err := client.BlockNumber(ctx)
			if err != nil {
				return err
			}
			seen := map[common.Address]bool{}
			for to := head; to > head-40_000 && found < 2; to -= 5_000 {
				logs, err := client.FilterLogs(ctx, ethereum.FilterQuery{
					FromBlock: new(big.Int).SetUint64(to - 4_999), ToBlock: new(big.Int).SetUint64(to),
					Addresses: []common.Address{m.comet}, Topics: [][]common.Hash{{withdraw}},
				})
				if err != nil {
					return err
				}
				for _, lg := range logs {
					user := common.BytesToAddress(lg.Topics[1].Bytes())
					if seen[user] || found >= 2 {
						continue
					}
					seen[user] = true
					pos, err := fetchCompound(ctx, client, parsed, m, user, ethUSD)
					if err != nil || !pos.HasDebt {
						continue
					}
					found++
					t.Logf("%s %s HF %.4f coll $%.2f debt $%.2f debt %s @ $%.2f | %s", m.label(), user.Hex(),
						pos.HealthFactor, pos.CollateralUSD, pos.DebtUSD, pos.Debt.Symbol, pos.Debt.PriceUSD, pos.AppURL)
					if pos.HealthFactor < 1 || pos.HealthFactor > 1000 || pos.Debt.PriceUSD <= 0 {
						t.Errorf("%s implausible position %+v", m.label(), pos)
					}
					if strings.EqualFold(m.market, "WETH") && (pos.Debt.PriceUSD < ethUSD*0.9 || pos.Debt.PriceUSD > ethUSD*1.1) {
						t.Errorf("WETH debt priced $%.2f, ETH is $%.2f", pos.Debt.PriceUSD, ethUSD)
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Errorf("%s: %v", m.label(), err)
		}
		if found == 0 {
			t.Logf("%s: no borrower in recent blocks", m.label())
		}
	}
}
