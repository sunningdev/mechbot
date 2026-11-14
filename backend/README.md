# MechBot

Telegram liquidation alerts plus a web console. Every 5 minutes the bot reads Aave V3, Compound V3 and Morpho Blue positions and sends a DM when the health factor drops below the alert line.

- Starter (free forever): 3 addresses, Aave V3 alerts
- Pro: 20 addresses, alerts for every protocol. 8 USD / month, 20 USD / quarter (save 17%), 60 USD / year (save 38%) in USDT or USDC, activated automatically when the transfer reaches the user's deposit address
- Referrals: on a new user's first purchase, both the referrer and the new user get Pro days (7 each for monthly, 15 for quarterly, 30 for yearly). Rewards stack across referrals; renewals are not rewarded again

## Local development

Two terminals:

```powershell
# Terminal 1: backend (bot + API, listens on 127.0.0.1:8080)
cd backend
go run .

# Terminal 2: frontend (http://localhost:5173, /api is proxied to the backend)
cd frontend
npm install
npm run dev
```

`.env` needs at least `TELEGRAM_BOT_TOKEN`. Use a separate test bot locally: two processes polling the same token get Telegram 409 conflicts. Data lives in `backend/mechbot.db` (SQLite) and tables are created on first start; `schema/mechbot.sql` is there for a future move to MySQL. Billing is enabled once at least one of `PAY_XPUB_EVM` or `PAY_XPUB_TRON` is set (see "Payments" below).

## Build and deploy

```powershell
cd frontend
npm run build
```

The backend serves the pages in `frontend/dist`; open `http://127.0.0.1:8080`. In production, put Nginx or Caddy in front with HTTPS, proxy to `127.0.0.1:8080`, and set `SITE_URL=https://your-domain` and `COOKIE_SECURE=true` in `.env`.

Linux build from Windows:

```powershell
cd backend
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -o mechbot .
Remove-Item Env:GOOS,Env:GOARCH,Env:CGO_ENABLED
```

## Login

On the website, "Log in with Telegram" shows a 4-digit code and opens @MechFi_bot; the user checks the code and taps "Confirm login", and the page enters the console by itself. Logging in also gives the bot DM permission, so alerts can always be delivered.

## Payments

Every user has a permanent deposit address derived from an HD wallet: one EVM address shared by Ethereum, BNB Chain, Arbitrum, Base and Optimism, plus one TRON address. The user sends USDT or USDC to their own address; the backend scans every 30 seconds, activates the matching plan and notifies the user and the referrer on Telegram.

There are two ways to pay, and both end up as a plain transfer to the same deposit address:

- Send manually: copy the address or scan the QR code. Default for Chinese, Russian and Vietnamese.
- Connect a wallet (RainbowKit): the backend reads the wallet's balances, suggests a network and coin with enough funds and gas, and the wallet sends a single ERC-20 `transfer`. No approvals, no signatures. Default for English, French, Spanish and Portuguese. Set `WALLETCONNECT_PROJECT_ID` (free at cloud.reown.com) to enable mobile wallets; without it only browser-extension wallets work.

| Amount received | Pro |
|---------|------|
| ≥ 8 USD | 30 days |
| ≥ 20 USD | 90 days |
| ≥ 60 USD | 365 days |

A shortfall of up to 0.3 is tolerated, so exchange withdrawal fees don't drop a payment to the lower tier. Anything above a tier is not carried over. Less than that is recorded as "amount too low" and the user is notified.

TRON is offered only for quarterly and yearly plans: sweeping one TRON address burns about 6.5 TRX of energy (around 2 USD), which doesn't pay off for a monthly plan. If someone sends 8 USDT on TRON anyway, they still get 30 days.

### One-time setup (on your own computer, never on the server)

```powershell
cd backend
go run ./cmd/mechwallet init   # generates a 24-word mnemonic; write it on paper and keep it offline
go run ./cmd/mechwallet xpub   # asks for the mnemonic, prints PAY_XPUB_EVM / PAY_XPUB_TRON
```

Put the two xpub lines in the server's `.env` and restart. The server only has the xpubs: it can derive addresses but cannot move funds. Lose the mnemonic and the funds are gone; leak it and they will be stolen.

`xpub` also prints the index 0 address. That is the **fee wallet**, used to send gas to deposit addresses when sweeping; it is never assigned to a user. Import the mnemonic into MetaMask / TronLink and it is the first account; keep a little BNB / ETH / TRX in it.

### Sweeping

Funds sit on each user's deposit address. Sweep them to your main wallet from your own computer from time to time (full steps and common errors in [SWEEP.md](SWEEP.md)):

```powershell
go run ./cmd/mechwallet balances -db mechbot.db                              # balances per address
go run ./cmd/mechwallet sweep -chain bsc -to <your EVM wallet> -db mechbot.db       # dry run, shows the plan only
go run ./cmd/mechwallet sweep -chain bsc -to <your EVM wallet> -db mechbot.db -yes
go run ./cmd/mechwallet sweep -chain tron -to <your TRON wallet> -min 50 -db mechbot.db -yes
```

`-db` points to a copy of the server database (only the address list is read). By default only addresses holding at least 20 USDT are swept. Rough costs: under 1 cent per address on BSC and Arbitrum; about 1 cent on Ethereum normally, 0.2–2 USD when congested; about 6.5 TRX of energy per transfer on TRON (about 13 TRX if the receiving address has never held USDT), so wait until you have 50 USDT or more, or rent energy first.

## Bot commands

```text
/start  /add 0x...  /list  /remove 0x...  /status
```

Referral links: `https://t.me/MechFi_bot?start=ref_<telegram_id>` or `SITE_URL/?ref=<telegram_id>`.
