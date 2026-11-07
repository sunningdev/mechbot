# MechBot

DeFi 借贷清算预警。盯住有真实借款的仓位，健康因子跌破预警线就通过 Telegram 私聊提醒。网页控制台负责登录、管地址、看仓位和开通 Pro。

Telegram：[@MechFi_bot](https://t.me/MechFi_bot)。口号：Stay calm, stay ahead.

## 业务

不做通用钱包追踪。只服务在 Aave、Compound、Morpho 上有借款的人：仓位还在，通知就一直有用。

监控范围是 Ethereum、Arbitrum、Optimism 上的：

- Aave V3（以太坊主池，以及 Lido、EtherFi 市场）
- Compound V3（USDC、USDT、WETH）
- Morpho Blue

预警不只报一个健康因子。消息里带抵押、债务、距离清算线还有多远，以及大概再还多少或补多少抵押能回到安全线。同一仓位跌破阈值只发一次，回到安全线再发「仓位已安全」，避免每 5 分钟刷屏。

| | Starter（免费） | Pro |
|---|---|---|
| 地址数 | 3 | 20 |
| 协议预警 | 仅 Aave V3 | Aave、Compound、Morpho |
| 提前预警、借款利率、稳定币/LST 脱锚 | 无 | 有 |
| 清算地图 | 延迟 1 小时，粗分档 | 实时，更细的价格档 |

新用户注册即送 7 天 Pro 体验，到期自动回到 Starter，不扣费。Pro 价格：5 USDT / 30 天，12 USDT / 90 天，36 USDT / 365 天。每个用户一个永久专属充值地址，转到账即开通。新用户第一次购买时，推荐人和新用户各得 Pro 天数（月付 7 天、季付 15 天、年付 30 天），续费不再奖励。

网页用 Telegram 登录：页面给出验证码，用户在机器人里核对并确认。登录同时拿到私聊权限，预警才能发得出去。

## 技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go 1.26，标准库 `net/http`，一个进程同时跑机器人和网页 |
| 链上读取 | go-ethereum；Aave / Compound 直接调合约，Morpho 走官方 GraphQL |
| 节点 | 每条链可配多个 RPC，失败自动换下一个，并带公共备用节点 |
| 行情 | Binance 公开 WebSocket（ETH、BTC），用于提前预警 |
| 脱锚 | CoinGecko 价格，对照 USDC、USDT、DAI、USDe、GHO、stETH、eETH |
| 数据库 | SQLite（`modernc.org/sqlite`），首次启动自动建表 |
| 充值 | BIP-39 / HD 派生。服务器只存 xpub，派生地址、扫 USDT 转账；私钥留在本机的 `mechwallet` 工具里 |
| 前端 | React 19、TypeScript、Vite、React Router。界面文案覆盖英、俄、西、简中、繁中、越、法、葡 |
| 登录态 | HttpOnly Cookie，会话 30 天 |

## 架构

一个 Go 进程里有几条常驻循环，共用同一份 SQLite 和同一份链上读取结果。网页和机器人只是两扇门。

```text
Telegram 用户 ──长轮询──┐
                        │
网页 (React) ── /api ───┤
                        ▼
                   main（一个进程）
                   ├── telegram   命令、登录确认、发预警
                   ├── web        JSON API，生产环境同时托管前端静态文件
                   ├── monitor    每 5 分钟读仓位，按阈值决定发不发
                   ├── price      Binance 行情，Pro 仓位接近阈值时提前重读
                   ├── depeg      稳定币 / LST 偏离锚定价
                   ├── pay        每 30 秒扫专属地址上的 USDT
                   ├── liqmap     慢速扫描以太坊借款人，汇总清算价格墙
                   ├── renewal    Pro / 体验到期前提醒
                   └── watchdog   节点或扫描连续失败时通知运营者
                        │
                        ▼
              SQLite（用户、地址、预警状态、充值、清算地图）
                        │
                        ▼
         Aave / Compound 合约    Morpho API    Binance / CoinGecko
```

数据怎么走：

1. 用户在网页或机器人里添加 `0x` 地址，并设自己的健康因子阈值（默认 1.2）。
2. `monitor` 按轮询间隔拉出全部地址，并发查询各协议仓位（同时最多 4 路），把快照写入数据库。
3. 健康因子跌破该地址的阈值，且冷却期（1 小时）外，才发 Telegram。回到阈值以上发一条恢复消息。
4. Pro 用户额外走一条快路径：行情每 5 秒更新，仓位离预警线 0.15 以内时约每分钟重读一次链上数据，在预言机跟上之前先发提前预警。借款利率跳变和抵押品脱锚也只对 Pro 发送。
5. 充值扫描看到专属地址收到 USDT 后，按金额档位延长 `expire_at`，并通知付款人和推荐人。Pro 是否有效只看到期时间，到期降级不需要单独的任务。

清算地图和预警分开跑。地图自己用一组 RPC，慢慢扫 Aave 与 Morpho 的借款人，估出 ETH 跌到什么价会触发清算，再按价格分档。它不跟 5 分钟预警抢节点。

### 代码布局

```text
backend/
  main.go                 启动、装配各循环
  cmd/mechwallet/         本机工具：助记词、xpub、余额、归集（不要放服务器）
  internal/
    config/               .env
    store/                SQLite 与表结构
    protocol/             Aave、Compound、Morpho 读仓
    evmrpc/               RPC 连接池与故障切换
    monitor/              轮询、提前预警、利率提醒
    notify/               预警文案
    depeg/                脱锚
    price/                Binance 行情
    pay/                  HD 地址、扫链、入账
    liqmap/               清算地图
    renewal/              到期提醒
    telegram/             Bot API 与命令
    web/                  HTTP API 与静态文件
    watchdog/             故障告警
  schema/mechbot.sql      换 MySQL 时用的建表脚本
frontend/
  src/pages/              落地页、控制台、清算地图、会员、推荐、博客
  src/components/         仓位、账单、登录
  src/locales/            界面翻译
```

网页路由：`/` 落地页，`/blog` 文章，`/app` 控制台（地址与仓位、`/app/map` 清算地图、`/app/pro` 会员、`/app/referral` 推荐）。开发时 Vite 把 `/api` 代理到 `127.0.0.1:8080`；`npm run build` 之后由 Go 直接提供 `frontend/dist`。

机器人命令：`/start`、`/add`、`/list`、`/remove`、`/set_threshold`、`/status`、`/help`。

## 本地运行

```powershell
# 终端 1
cd backend
go run .

# 终端 2
cd frontend
npm install
npm run dev
```

页面在 http://localhost:5173 ，后端在 http://127.0.0.1:8080 。`.env` 至少要有 `TELEGRAM_BOT_TOKEN`。充值地址、归集和上线配置见 [backend/README.md](backend/README.md)。
