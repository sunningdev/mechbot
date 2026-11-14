# MechBot

Liquidation alerts for DeFi lending. MechBot watches positions with real debt and sends a Telegram DM when the health factor drops below your alert line. The web console handles login, addresses, positions and Pro billing.

Telegram: [@MechFi_bot](https://t.me/MechFi_bot). Slogan: Stay calm, stay ahead.

## Product

MechBot is not a general wallet tracker. It serves people who borrow on Aave, Compound and Morpho: as long as the position is open, the alerts stay useful.

Coverage on Ethereum, Arbitrum, Base and Optimism:

- Aave V3 (Ethereum Core, plus the Lido and EtherFi markets)
- Compound V3 (USDC, USDT, WETH)
- Morpho Blue

An alert is more than a health factor. It includes collateral, debt, how far the position is from liquidation, and roughly how much to repay or add to get back above the line. A position that crosses the threshold is alerted once, then again with "position is safe" when it recovers, so nobody gets spammed every 5 minutes.

| | Starter (free) | Pro |
|---|---|---|
| Addresses | 3 | 20 |
| Protocol alerts | Aave V3 only | Aave, Compound, Morpho |
| Early alerts, borrow rate, stablecoin / LST depeg | No | Yes |
| Liquidation map | 1 hour delay, coarse bands | Live, finer price bands |

New users get a 7-day Pro trial on sign-up; it falls back to Starter afterwards with no charge. Pro costs 8 USD / 30 days, 20 USD / 90 days (save 17%) or 60 USD / 365 days (save 38%), paid in USDT or USDC. Every user has a permanent personal deposit address; Pro is activated as soon as the transfer lands. Users can also connect a wallet and send the transfer in one click. On a new user's first purchase, both the referrer and the new user get extra Pro days (7 for monthly, 15 for quarterly, 30 for yearly); renewals are not rewarded again.

Web login goes through Telegram: the page shows a code, the user checks it in the bot and confirms. Logging in also grants DM permission, so alerts can always be delivered.

## Tech stack

| Layer | Choice |
|---|---|
| Backend | Go 1.26, standard library `net/http`, one process runs both the bot and the web API |
| On-chain reads | go-ethereum; Aave / Compound contracts are called directly, Morpho goes through its official GraphQL API |
| Nodes | Several RPCs per chain, automatic failover to the next one, plus built-in public fallbacks |
| Prices | Binance public WebSocket (ETH, BTC) for early alerts |
| Depeg | CoinGecko prices for USDC, USDT, DAI, USDe, GHO, stETH, eETH |
| Database | SQLite (`modernc.org/sqlite`), tables are created on first start |
| Payments | BIP-39 / HD derivation. The server only stores xpubs, derives addresses and scans for USDT / USDC transfers; private keys stay in the local `mechwallet` tool |
| Frontend | React 19, TypeScript, Vite, React Router, RainbowKit (loaded only for wallet payment). UI in English, Russian, Spanish, Simplified and Traditional Chinese, Vietnamese, French and Portuguese |
| Sessions | HttpOnly cookie, 30 days |

## Architecture

One Go process runs several long-lived loops that share one SQLite database and one set of on-chain reads. The website and the bot are just two front doors.

```text
Telegram user ──long poll──┐
                           │
Web (React) ──── /api ─────┤
                           ▼
                   main (one process)
                   ├── telegram   commands, login confirmation, alert delivery
                   ├── web        JSON API, also serves the built frontend in production
                   ├── monitor    reads positions every 5 minutes, decides whether to alert
                   ├── price      Binance prices, re-reads Pro positions near their line early
                   ├── depeg      stablecoin / LST peg deviation
                   ├── pay        scans deposit addresses for USDT / USDC every 30 seconds
                   ├── liqmap     slow scan of Ethereum borrowers, builds liquidation walls
                   ├── renewal    reminders before Pro / trial expiry
                   └── watchdog   tells the operator when nodes or scans keep failing
                           │
                           ▼
              SQLite (users, addresses, alert state, deposits, liquidation map)
                           │
                           ▼
         Aave / Compound contracts    Morpho API    Binance / CoinGecko
```

How data flows:

1. A user adds a `0x` address on the web or in the bot and sets a health-factor threshold (default 1.2).
2. On every poll, `monitor` loads all addresses, queries each protocol concurrently (at most 4 at a time) and stores a snapshot.
3. A Telegram alert is sent only when the health factor is below the address's threshold and outside the cooldown (1 hour; 15 minutes when the health factor is under 1.05). A recovery message follows once it is back above the threshold. Changing the threshold triggers an immediate check.
4. Pro users get a fast path: prices update every 5 seconds, and positions within 0.15 of their line are re-read about once a minute, so an early alert can go out before the oracle catches up. Borrow-rate spikes and collateral depegs are Pro-only too.
5. When the deposit scanner sees USDT / USDC arrive on a deposit address, it extends `expire_at` by the matching plan and notifies the payer and the referrer. Pro status depends only on the expiry time, so downgrades need no separate job.

The liquidation map runs separately from alerts. It has its own RPC set, slowly scans Aave and Morpho borrowers, estimates the ETH price that would liquidate each one and buckets the results by price. It never competes with the 5-minute alert loop for nodes.

### Code layout

```text
backend/
  main.go                 startup, wires the loops together
  cmd/mechwallet/         local tool: mnemonic, xpub, balances, sweeping (never on the server)
  internal/
    config/               .env
    store/                SQLite and schema
    protocol/             Aave, Compound, Morpho position readers
    evmrpc/               RPC pool and failover
    monitor/              polling, early alerts, rate alerts
    notify/               alert messages
    depeg/                depeg detection
    price/                Binance prices
    pay/                  HD addresses, chain scanning, crediting, wallet balances
    liqmap/               liquidation map
    renewal/              expiry reminders
    telegram/             Bot API and commands
    web/                  HTTP API and static files
    watchdog/             failure alerts
  schema/mechbot.sql      schema for a future move to MySQL
frontend/
  src/pages/              landing, console, liquidation map, billing, referral, blog
  src/components/         positions, billing, wallet payment, login
  src/locales/            UI translations
```

Web routes: `/` landing page, `/blog` posts, `/app` console (addresses and positions, `/app/map` liquidation map, `/app/pro` billing, `/app/referral` referrals). In development Vite proxies `/api` to `127.0.0.1:8080`; after `npm run build` Go serves `frontend/dist` directly.

Bot commands: `/start`, `/add`, `/list`, `/remove`, `/set_threshold`, `/status`, `/help`.

## Running locally

```powershell
# Terminal 1
cd backend
go run .

# Terminal 2
cd frontend
npm install
npm run dev
```

The site runs at http://localhost:5173 and the backend at http://127.0.0.1:8080. `.env` needs at least `TELEGRAM_BOT_TOKEN`; use a separate test bot locally so you don't fight the production bot for updates. Deposit addresses, sweeping and deployment are covered in [backend/README.md](backend/README.md).
