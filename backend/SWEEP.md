# USDT 归集说明

用户付款的 USDT 分散在每个人的专属充值地址上，需要定期**由你本人在自己电脑上**运行命令，把钱转到主钱包。服务器不会、也没有能力自动归集。

## 为什么要手动

- 服务器 `.env` 里只有 xpub（公钥）：能生成充值地址、能查到账，**不能签名转账**。
- 转账要用助记词推出私钥。助记词只在运行归集命令时在本机输入（输入时不显示），用完不保存。
- 服务器就算被入侵，也转不走任何资金。

## 归集到哪里

| 链 | `-chain` | 主钱包（`-to`） |
|---|---|---|
| BNB Chain | `bsc` | `0x4C02C6bE8A572c1dDf401ee03B7a0a02d6B57842` |
| Arbitrum One | `arbitrum` | `0x4C02C6bE8A572c1dDf401ee03B7a0a02d6B57842` |
| Ethereum | `ethereum` | `0x4C02C6bE8A572c1dDf401ee03B7a0a02d6B57842` |
| TRON | `tron` | `TBCVLnWiKHSRQVkXUbeiYtTJHB4p4KKDuK` |

主钱包必须是你自己掌握私钥的地址，填错转出去追不回。

## 准备：手续费钱包

充值地址里只有 USDT，没有付手续费的币。归集时程序先从**手续费钱包**给每个充值地址补 gas，再把 USDT 转走。

手续费钱包 = 助记词推出的 **index 0** 地址（运行 `go run ./cmd/mechwallet xpub` 会打印出来；把助记词导入 MetaMask / TronLink，第一个账户也是它）。提前往里转：

| 链 | 放什么 | 每个充值地址大约消耗 |
|---|---|---|
| BNB Chain | BNB | 不到 1 美分 |
| Arbitrum | ETH（Arbitrum 网络） | 不到 1 美分 |
| Ethereum | ETH（主网） | 平时约 1 美分，拥堵时 0.2–2 美元 |
| TRON | TRX | 每笔烧约 6.5 TRX 能量费，收款地址从没收过 USDT 时约 13 TRX |

TRON 默认会把每个待归集的充值地址补到 **30 TRX**（`-trx 30`），手续费钱包里要按「待归集地址数 × 30」备好 TRX。补进去没烧完的 TRX 会留在充值地址上，下次归集还能接着用。

## 数据库

命令需要读取充值地址列表（只读，不改数据）：

- 在服务器上：把服务器的 `mechbot.db` 复制一份到你电脑的 `backend` 目录。**不要在服务器上运行归集命令**，也不要在服务器上输入助记词。
- 本地开发时：直接在 `backend` 目录运行即可，默认读 `mechbot.db`；也可以用 `-db 路径` 指定。

## 操作步骤

以下命令都在 `backend` 目录运行（PowerShell）。

### 0. 查看所有充值地址（可选）

```powershell
go run ./cmd/mechwallet addresses -db mechbot.db
```

列出每个用户的 index、Telegram ID、用户名、EVM 地址（BNB Chain / Arbitrum / Ethereum 共用）和 TRON 地址，不需要助记词。

### 1. 看余额

```powershell
go run ./cmd/mechwallet balances -db mechbot.db
```

按链列出有 USDT 的充值地址和合计金额，不需要助记词。

### 2. 演练（不发交易）

```powershell
go run ./cmd/mechwallet sweep -chain bsc -to 0x4C02C6bE8A572c1dDf401ee03B7a0a02d6B57842 -db mechbot.db
```

会要求输入助记词，然后显示：要归集哪些地址、各补多少 gas、手续费钱包地址。**不加 `-yes` 不会发送任何交易。**

### 3. 正式归集

确认演练结果无误后，加上 `-yes`：

```powershell
go run ./cmd/mechwallet sweep -chain bsc      -to 0x4C02C6bE8A572c1dDf401ee03B7a0a02d6B57842 -db mechbot.db -yes
go run ./cmd/mechwallet sweep -chain arbitrum -to 0x4C02C6bE8A572c1dDf401ee03B7a0a02d6B57842 -db mechbot.db -yes
go run ./cmd/mechwallet sweep -chain ethereum -to 0x4C02C6bE8A572c1dDf401ee03B7a0a02d6B57842 -db mechbot.db -yes
go run ./cmd/mechwallet sweep -chain tron     -to TBCVLnWiKHSRQVkXUbeiYtTJHB4p4KKDuK -min 50 -db mechbot.db -yes
```

每条链单独跑一次。

### 4. 核对

在区块浏览器查看主钱包是否收到：

- BNB Chain：https://bscscan.com
- Arbitrum：https://arbiscan.io
- Ethereum：https://etherscan.io
- TRON：https://tronscan.org

再跑一次 `balances`，已归集的地址应不再出现。

## 参数

| 参数 | 默认 | 说明 |
|---|---|---|
| `-chain` | 必填 | `tron` / `bsc` / `arbitrum` / `ethereum` |
| `-to` | 必填 | 主钱包地址：EVM 链用 0x 开头，TRON 用 T 开头 |
| `-min` | `20` | 余额低于这个数（USDT）的地址先不归集，免得手续费比钱还多 |
| `-trx` | `30` | 仅 TRON：每个充值地址补到多少 TRX 用来烧能量 |
| `-db` | `mechbot.db` | 数据库路径（也读 `.env` 里的 `DB_PATH`） |
| `-yes` | 不加 | 加上才真正发送交易；不加只演练 |

RPC 节点默认用公共节点，可在 `.env` 里用 `BSC_RPC_URL`、`ARB_RPC_URL`、`ETH_RPC_URL` 指定；TRON 可设 `TRON_API_KEY` 提高查询限额。

## 建议节奏

- 不必每笔都归集，攒到一定金额集中归集一次。
- BNB Chain、Arbitrum 手续费极低，`-min 20` 即可。
- TRON 每笔成本约 2 美元，建议 `-min 50` 以上；量大时可以先租能量再归集。
- Ethereum 主网手续费波动大，挑不拥堵的时候归集。

## 常见报错

| 报错 | 原因 / 处理 |
|---|---|
| `助记词无效，检查拼写和词数` | 单词拼错或数量不对（应为 24 个词） |
| `index N 推出的地址 … 与数据库 … 不一致，助记词不对` | 输入的助记词不是生成这批充值地址的那一套，立即停止，核对助记词 |
| `补 gas 失败（手续费钱包 … 余额够吗？）` / `补 TRX 失败` | 手续费钱包里的 BNB / ETH / TRX 不够，充值后重跑 |
| `找不到数据库 mechbot.db` | 不在 `backend` 目录运行，或用 `-db` 指定正确路径 |
| `-to 需要 T 开头的 TRON 地址` / `-to 需要 0x 开头的地址` | 链和地址类型不匹配 |

中途失败可以直接重跑：已转走的地址余额为 0，会被自动跳过。

## 安全

- 助记词只抄在纸上离线保存：不截图、不存云盘、不发聊天、不放服务器。
- 只在自己信任的电脑上运行归集命令。
- 助记词丢了，充值地址里的钱就找不回；泄露了，钱会被转走。
