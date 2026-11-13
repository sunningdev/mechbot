# Sweeping deposits

User payments sit on each user's personal deposit address. **You** sweep them to your main wallet from time to time by running a command **on your own computer**. The server does not sweep and has no way to.

## Why it is manual

- The server `.env` only holds xpubs (public keys): it can derive deposit addresses and see payments, but **cannot sign transfers**.
- Transfers need private keys derived from the mnemonic. The mnemonic is typed in locally only while the sweep command runs (input is hidden) and is never saved.
- Even if the server is compromised, no funds can be moved.

## Where funds go

| Chain | `-chain` | Main wallet (`-to`) |
|---|---|---|
| BNB Chain | `bsc` | `<your EVM wallet>` |
| Arbitrum One | `arbitrum` | `<your EVM wallet>` |
| Base | `base` | `<your EVM wallet>` |
| Optimism | `optimism` | `<your EVM wallet>` |
| Ethereum | `ethereum` | `<your EVM wallet>` |
| TRON | `tron` | `<your TRON wallet>` |

The main wallet must be an address whose keys you control. A wrong address cannot be undone.

## Preparation: the fee wallet

Deposit addresses only hold USDT / USDC and nothing to pay gas with. The sweep first sends gas from the **fee wallet** to each deposit address, then moves the stablecoins.

Fee wallet = the **index 0** address of your mnemonic (`go run ./cmd/mechwallet xpub` prints it; it is also the first account when you import the mnemonic into MetaMask / TronLink). Fund it in advance:

| Chain | What to hold | Cost per deposit address |
|---|---|---|
| BNB Chain | BNB | under 1 cent |
| Arbitrum | ETH (on Arbitrum) | under 1 cent |
| Base | ETH (on Base) | under 1 cent |
| Optimism | ETH (on Optimism) | under 1 cent |
| Ethereum | ETH (mainnet) | about 1 cent normally, 0.2–2 USD when congested |
| TRON | TRX | about 6.5 TRX of energy per transfer, about 13 TRX if the receiving address has never held USDT |

On TRON each deposit address is topped up to **30 TRX** by default (`-trx 30`), so keep "addresses to sweep × 30" TRX in the fee wallet. TRX that isn't burned stays on the deposit address and is used next time.

## Database

The command reads the list of deposit addresses (read-only, nothing is changed):

- Production: copy the server's `mechbot.db` into `backend` on your computer. **Never run the sweep on the server** and never type the mnemonic there.
- Local development: just run it in `backend`; it reads `mechbot.db` by default, or pass `-db <path>`.

## Steps

Run everything below in `backend` (PowerShell).

### 0. List deposit addresses (optional)

```powershell
go run ./cmd/mechwallet addresses -db mechbot.db
```

Lists each user's index, Telegram ID, username, EVM address (shared by all EVM chains) and TRON address. No mnemonic needed.

### 1. Check balances

```powershell
go run ./cmd/mechwallet balances -db mechbot.db
```

Lists deposit addresses holding stablecoins, per chain, with totals. No mnemonic needed.

### 2. Dry run (no transactions)

```powershell
go run ./cmd/mechwallet sweep -chain bsc -to <your EVM wallet> -db mechbot.db
```

Asks for the mnemonic, then shows which addresses would be swept, how much gas each gets and the fee wallet address. **Nothing is sent without `-yes`.**

### 3. Sweep for real

Once the dry run looks right, add `-yes`:

```powershell
go run ./cmd/mechwallet sweep -chain bsc      -to <your EVM wallet> -db mechbot.db -yes
go run ./cmd/mechwallet sweep -chain arbitrum -to <your EVM wallet> -db mechbot.db -yes
go run ./cmd/mechwallet sweep -chain base     -to <your EVM wallet> -db mechbot.db -yes
go run ./cmd/mechwallet sweep -chain optimism -to <your EVM wallet> -db mechbot.db -yes
go run ./cmd/mechwallet sweep -chain ethereum -to <your EVM wallet> -db mechbot.db -yes
go run ./cmd/mechwallet sweep -chain tron     -to <your TRON wallet> -min 50 -db mechbot.db -yes
```

Run each chain separately.

### 4. Verify

Check that the main wallet received the funds on a block explorer:

- BNB Chain: https://bscscan.com
- Arbitrum: https://arbiscan.io
- Base: https://basescan.org
- Optimism: https://optimistic.etherscan.io
- Ethereum: https://etherscan.io
- TRON: https://tronscan.org

Run `balances` again; swept addresses should no longer appear.

## Flags

| Flag | Default | Meaning |
|---|---|---|
| `-chain` | required | `tron` / `bsc` / `arbitrum` / `base` / `optimism` / `ethereum` |
| `-to` | required | Main wallet: 0x address on EVM chains, T address on TRON |
| `-min` | `20` | Skip addresses holding less than this many USD, so fees don't exceed the amount |
| `-trx` | `30` | TRON only: top each deposit address up to this many TRX for energy |
| `-db` | `mechbot.db` | Database path (also reads `DB_PATH` from `.env`) |
| `-yes` | off | Actually send transactions; without it the command only does a dry run |

Public RPC nodes are used by default; override them in `.env` with `BSC_RPC_URL`, `ARB_RPC_URL`, `BASE_RPC_URL`, `OP_RPC_URL` and `ETH_RPC_URL`. On TRON, `TRON_API_KEY` raises the query limit.

## Suggested routine

- No need to sweep every payment; let it build up and sweep in batches.
- BNB Chain, Arbitrum, Base and Optimism are very cheap, `-min 20` is fine.
- TRON costs about 2 USD per address, so use `-min 50` or more; rent energy first when volumes grow.
- Ethereum mainnet fees swing a lot; sweep when the network is quiet.

## Common errors

The tool prints its errors in Chinese:

| Error | Cause / fix |
|---|---|
| `助记词无效，检查拼写和词数` | Invalid mnemonic: a word is misspelled or the count is wrong (should be 24 words) |
| `index N 推出的地址 … 与数据库 … 不一致，助记词不对` | The address derived at index N doesn't match the database: this is not the mnemonic that generated these deposit addresses. Stop and check the mnemonic |
| `补 gas 失败（手续费钱包 … 余额够吗？）` / `补 TRX 失败` | Topping up gas / TRX failed: the fee wallet lacks BNB / ETH / TRX. Fund it and rerun |
| `找不到数据库 mechbot.db` | Database not found: run from `backend`, or pass the right path with `-db` |
| `-to 需要 T 开头的 TRON 地址` / `-to 需要 0x 开头的地址` | `-to` doesn't match the chain: TRON needs a T address, EVM chains need a 0x address |

A failed run can simply be rerun: addresses already swept have a zero balance and are skipped.

## Security

- Keep the mnemonic on paper, offline: no screenshots, no cloud storage, no chat apps, never on the server.
- Only run the sweep on a computer you trust.
- Lose the mnemonic and the deposits are gone; leak it and they will be stolen.
