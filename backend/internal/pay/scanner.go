package pay

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	scanEvery       = 30 * time.Second
	logStep         = 2000 // blocks per eth_getLogs call
	maxStepsPerScan = 20   // catch-up after downtime spreads over several scans
	addrChunk       = 100  // deposit addresses per topic filter
	tronPause       = 150 * time.Millisecond
	headLag         = 10
)

func (s *Service) Run(ctx context.Context) {
	if !s.Enabled() {
		return
	}
	ticker := time.NewTicker(scanEvery)
	defer ticker.Stop()
	for {
		s.Scan(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// Scan checks every enabled chain once; concurrent callers wait for the running scan instead of doubling RPC load.
func (s *Service) Scan(ctx context.Context) {
	s.scanMu.Lock()
	defer s.scanMu.Unlock()
	for _, c := range s.chains {
		err := s.scanChain(ctx, c)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Printf("扫描充值失败 %s: %v", c.Key, err)
		}
		if s.health != nil {
			s.health.Report(ctx, "充值扫描 · "+c.Name, err)
		}
	}
}

// scanChain keeps one chain's panic from stopping the scan loop for every chain.
func (s *Service) scanChain(ctx context.Context, c Chain) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("内部错误: %v", p)
		}
	}()
	if c.Family == "tron" {
		return s.scanTron(ctx, c)
	}
	return s.rpc.Do(ctx, c.rpc, func(client *ethclient.Client) error {
		return s.scanEVM(ctx, c, client)
	})
}

// ScanUser refreshes the user's watch window and runs a scan right away, for the "我已转账" button.
func (s *Service) ScanUser(ctx context.Context, telegramID int64) error {
	if _, _, err := s.DepositAddress(telegramID); err != nil {
		return err
	}
	s.Scan(ctx)
	return nil
}

func (s *Service) scanEVM(ctx context.Context, c Chain, client *ethclient.Client) error {
	latest, err := client.BlockNumber(ctx)
	if err != nil {
		return err
	}
	// Load-balanced public RPCs can serve getLogs from a node a few blocks behind the one that answered BlockNumber.
	margin := c.minConf + headLag
	if latest < margin {
		return nil
	}
	safe := latest - margin

	cursor, ok, err := s.store.ScanCursor(c.Key)
	if err != nil {
		return err
	}
	from := cursor + 1
	if !ok {
		from = 0
		if safe > c.lookback {
			from = safe - c.lookback
		}
	}

	owners, err := s.store.EVMDepositOwners()
	if err != nil {
		return err
	}
	if len(owners) == 0 {
		return s.store.SetScanCursor(c.Key, safe)
	}
	topics := make([]common.Hash, 0, len(owners))
	for addr := range owners {
		topics = append(topics, common.BytesToHash(common.HexToAddress(addr).Bytes()))
	}

	for step := 0; step < maxStepsPerScan && from <= safe; step++ {
		to := min(from+logStep-1, safe)
		for i := 0; i < len(topics); i += addrChunk {
			chunk := topics[i:min(i+addrChunk, len(topics))]
			logs, err := client.FilterLogs(ctx, ethereum.FilterQuery{
				FromBlock: new(big.Int).SetUint64(from),
				ToBlock:   new(big.Int).SetUint64(to),
				Addresses: []common.Address{c.usdt},
				Topics:    [][]common.Hash{{transferTopic}, nil, chunk},
			})
			if err != nil {
				return fmt.Errorf("getLogs %d-%d: %w", from, to, err)
			}
			for _, lg := range logs {
				if lg.Removed || len(lg.Topics) != 3 {
					continue
				}
				owner, ok := owners[strings.ToLower(common.BytesToAddress(lg.Topics[2].Bytes()).Hex())]
				if !ok {
					continue
				}
				amount := ToMicro(new(big.Int).SetBytes(lg.Data), c.decimals)
				s.credit(ctx, c, strings.ToLower(lg.TxHash.Hex()), int(lg.Index), owner, amount)
			}
		}
		if err := s.store.SetScanCursor(c.Key, to); err != nil {
			return err
		}
		from = to + 1
	}
	return nil
}

func (s *Service) scanTron(ctx context.Context, c Chain) error {
	owners, err := s.store.TronDepositOwners(time.Now())
	if err != nil {
		return err
	}
	for addr, owner := range owners {
		transfers, err := s.tron.incoming(ctx, addr)
		if err != nil {
			return err
		}
		for _, tr := range transfers {
			v, ok := new(big.Int).SetString(tr.Value, 10)
			if !ok {
				continue
			}
			s.credit(ctx, c, strings.ToLower(tr.TxID), 0, owner, ToMicro(v, c.decimals))
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(tronPause):
		}
	}
	return nil
}
