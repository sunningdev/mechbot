# MechBot

Telegram 清算预警 + 网页控制台。程序每 5 分钟查询 Aave V3、Compound V3、Morpho Blue，健康因子低于预警线就私聊提醒。

- Starter（永久免费）：3 个地址，Aave V3 预警
- Pro：20 个地址，全协议预警。5 USDT/月，12 USDT/季，36 USDT/年，转 USDT 到专属地址自动开通
- 推荐：新用户首次购买，推荐人和新用户各得 Pro 天数（月付各 7 天，季付各 15 天，年付各 30 天），推荐多人累加；续费不再奖励

## 本机开发

两个终端：

```powershell
# 终端 1：后端（机器人 + 接口，监听 127.0.0.1:8080）
cd backend
go run .

# 终端 2：页面（http://localhost:5173，/api 自动转发到后端）
cd frontend
npm install
npm run dev
```

`.env` 里至少要有 `TELEGRAM_BOT_TOKEN`。数据存在 `backend/mechbot.db`（SQLite），首次启动自动建表；以后换 MySQL 时用 `schema/mechbot.sql` 建库。`PAY_XPUB_EVM` 或 `PAY_XPUB_TRON` 至少填一个，网页才开放充值（见下方「充值」）。

## 打包上线

```powershell
cd frontend
npm run build
```

后端会直接提供 `frontend/dist` 里的页面，打开 `http://127.0.0.1:8080` 即可。上线时用 Nginx 或 Caddy 配 HTTPS 反代到 `127.0.0.1:8080`，并在 `.env` 里设置 `SITE_URL=https://你的域名`、`COOKIE_SECURE=true`。

## 登录

网页点「用 Telegram 登录」，页面显示 4 位验证码，跳到 @MechFi_bot 后核对验证码并点「确认登录」，网页自动进入控制台。登录本身就让机器人拿到了私聊权限，预警一定发得出去。

## 充值

每个用户有一个永久的专属充值地址（HD 钱包派生）：一个 EVM 地址同时用于 Ethereum、BNB Chain、Arbitrum，另有一个 TRON 地址。用户往自己的地址转 USDT，后台每 30 秒扫链，到账按金额自动开通，并在 Telegram 通知用户和推荐人。

| 到账金额 | 开通 |
|---------|------|
| ≥ 5 USDT | 30 天 |
| ≥ 12 USDT | 90 天 |
| ≥ 36 USDT | 365 天 |

超出档位的部分不累计；低于 5 USDT 记为「金额不足」并通知用户。

TRON 只展示季付和年付：归集一个 TRON 地址要烧约 6.5 TRX 能量费（约 2 美元），月付不划算。万一用户在 TRON 上转了 5 USDT，仍照常开通 30 天。

### 一次性设置（在你自己的电脑上，不要在服务器上）

```powershell
cd backend
go run ./cmd/mechwallet init   # 生成 24 词助记词，抄在纸上离线保存
go run ./cmd/mechwallet xpub   # 输入助记词，输出两行 PAY_XPUB_EVM / PAY_XPUB_TRON
```

把两行 xpub 写进服务器的 `.env` 后重启。服务器只有 xpub，只能推出地址，动不了钱。助记词丢了钱就找不回，泄露了钱就会被转走。

`xpub` 会同时打印 index 0 的地址：这是**手续费钱包**，归集时用它给充值地址打 gas。把助记词导入 MetaMask / TronLink，第一个账户就是它，往里放少量 BNB / ETH / TRX。

### 归集

钱分散在各用户的充值地址里，定期在本机归集到你的主钱包（完整步骤、主钱包地址和常见报错见 [SWEEP.md](SWEEP.md)）：

```powershell
go run ./cmd/mechwallet balances -db mechbot.db                       # 各地址余额
go run ./cmd/mechwallet sweep -chain bsc -to 0x你的主钱包 -db mechbot.db  # 演练，只显示计划
go run ./cmd/mechwallet sweep -chain bsc -to 0x你的主钱包 -db mechbot.db -yes
go run ./cmd/mechwallet sweep -chain tron -to T你的主钱包 -min 50 -db mechbot.db -yes
```

`-db` 指向服务器数据库的副本（只读取地址列表）。默认只归集余额 ≥ 20 USDT 的地址。成本参考：BSC、Arbitrum 每笔不到 1 美分；Ethereum 平时约 1 美分，拥堵时 0.2–2 美元；TRON 每笔约烧 6.5 TRX 能量费（收款地址没有 USDT 时约 13 TRX），建议攒到 50 USDT 以上再归集，或提前租能量。

## 机器人命令

```text
/start  /add 0x...  /list  /remove 0x...  /status
```

推荐链接：`https://t.me/MechFi_bot?start=ref_<telegram_id>` 或 `SITE_URL/?ref=<telegram_id>`。
