package i18n

var zhTW = Text{
	DangerHead: `⚠️ MechBot 清算預警

協議: %s
地址: %s
健康因子: %.4f（預警線 %.2f）
抵押品: %s
債務: %s
`,
	ETHDrop:    "ETH 再跌 %.1f%% 會被清算（ETH 清算價 %s）\n",
	DangerFoot: "\n健康因子低於 1.0 時倉位可能被清算。",

	EarlyHead: `⏱ MechBot 提前預警（Pro）

協議: %s
地址: %s
鏈上健康因子: %.4f
按幣安即時價格預計: %.4f（預警線 %.2f）
`,
	EarlyETH:  "ETH: %s → %s（%+.1f%%）\n",
	EarlyNote: "\n鏈上預言機通常比交易所慢，更新後就會跌破預警線。現在處理還來得及。\n",

	RateHigh: `📈 MechBot 借款利率提醒（Pro）

協議: %s
地址: %s
%s 借款年化: %.2f%%
債務: %s，按這個利率一年約多付 %s

利率突然升高通常是資金池被借空。循環貸收益可能已經變負，可以考慮先還一部分或換個市場。`,
	BorrowWord: "借款",
	RateCalm: `✅ MechBot 借款利率已回落

協議: %s
地址: %s
借款年化: %.2f%%`,

	DepegSideCollateral:   "抵押物",
	DepegSideDebt:         "借款",
	DepegEffectCollateral: "抵押物價值縮水",
	DepegEffectDebt:       "債務變大",
	DepegEffectFavorable:  "對你暫時有利，但脫錨通常伴隨劇烈波動",
	DepegEffectHF:         "，健康因子會隨預言機更新下降",
	Depeg: `⚠️ MechBot 脫錨預警（Pro）

協議: %s
地址: %s
%s %s 追蹤的 %s 目前 %.4f %s（偏離 %+.2f%%）
健康因子: %.4f

%s。建議先降低槓桿：還一部分借款或補充其他抵押物。%s`,
	DepegCalm: `✅ MechBot %s 已回到錨定附近

協議: %s
地址: %s
目前 %.4f %s`,

	AppLine:     "去處理: %s",
	RescueHead:  "\n怎麼處理（回到 %.2f）：\n",
	RescueRepay: "· 還款 %s\n",
	RescueAdd:   "· 或補抵押 %s\n",
	TokenApprox: "%s %s（約 %s）",

	SafeNoDebt: `✅ MechBot 倉位已安全

協議: %s
地址: %s

這筆借款已結清，目前沒有債務。`,
	Safe: `✅ MechBot 倉位已安全

協議: %s
地址: %s
健康因子: %.4f（預警線 %.2f）
抵押品: %s
債務: %s

已回到預警線以上。`,

	ListHead:   "地址 %s\n預警線 %.2f\n",
	ListFailed: "\n%s\n查詢失敗: %v\n",
	ListPos:    "\n%s\n健康因子 %.4f\n抵押 %s\n債務 %s\n",
	ListDrop:   "ETH 再跌 %.1f%% 清算（%s）\n",
	ListAPY:    "借款年化 %.2f%%\n",
	ListQuiet:  "\n無借款: %s\n",
	ListSep:    "，",

	AddedDebt:   "已新增，目前倉位：\n\n",
	AddedFailed: "已新增 %s。\n\n部分協議暫時查詢失敗，稍後用 /list 再看。",
	AddedQuiet: `已新增 %s。

這個地址目前沒有借款：已檢查 Aave V3、Compound V3、Morpho Blue（Ethereum · Arbitrum · Base · Optimism），不會被清算。

地址會繼續監控，以後在這些協議借款，會自動開始預警，不用重新新增。`,
	AddedStarter: "\n\nStarter 只對 Aave V3 發預警，Compound、Morpho 的借款需要 Pro。",

	StarterLimits: `· 只保留 Aave V3 預警，Compound、Morpho 預警停止
· 提前預警、利率提醒、脫錨預警停止
· 清算地圖改為延遲 1 小時`,
	ProEndingPaid:  "你的 Pro 將在 24 小時內到期",
	ProEndingTrial: "你的 7 天 Pro 體驗將在 24 小時內結束",
	ProEnding: `⏳ MechBot %s

到期後自動降為 Starter（免費）：
· 只監控最早新增的 %d 個地址，其餘暫停（不會刪除）
%s

繼續用 Pro：%s`,
	ProEndedPaid:   "Pro 已到期，已降為 Starter",
	ProEndedTrial:  "Pro 體驗已結束，已降為 Starter",
	ProEndedWatch:  "· 最早新增的 %d 個地址繼續監控",
	ProEndedPaused: "，其餘 %d 個已暫停（資料保留，續費後自動恢復）",
	ProEnded: `MechBot %s

%s
%s

續費後立刻恢復全部功能：%s`,

	StatusStarter: "Starter（永久免費，僅 Aave V3 預警）",
	StatusPro:     "Pro（全協議預警）",
	StatusTrial:   "Pro 體驗（全協議預警，到期回到 Starter）",
	StatusNone:    "無",
	Status: `MechBot 帳戶狀態

版本: %s
監控地址: %d/%d
Pro 到期: %s
推薦人: %s`,

	PlanMonth:   "Pro 月付",
	PlanQuarter: "Pro 季付",
	PlanYear:    "Pro 年付",
	PayTooLow:   "⚠️ 收到 %s（%s）\n\n低於最低方案 %d USD，未能開通 Pro。請補足到方案金額後重新轉帳。",
	PayBonus:    "\n🎁 好友推薦獎勵 +%d 天",
	PayDone:     "✅ MechBot Pro 已開通\n\n收到 %s（%s），%s %d 天%s\nPro 到期: %s",
	PayReferrer: "🎁 推薦獎勵 +%d 天\n\n你推薦的用戶首次購買了%s。\nPro 到期: %s",

	Help: `MechBot 清算預警

/add 0x... 新增監控地址（Starter 3 個，Pro 20 個）
/list 查看各協議健康因子
/remove 0x... 刪除地址
/status 查看帳戶狀態
/lang 切換訊息語言

每個地址的預警線預設 1.2，可在網頁控制台的地址卡片上修改。

Starter（永久免費）預警 Aave V3（Ethereum、Arbitrum、Base、Optimism）。Pro 額外預警 Compound V3 和 Morpho Blue。`,
	HelpConsole:       "\n\n網頁控制台: %s",
	TrialNote:         "\n\n🎁 新用戶送 %d 天 Pro 體驗：全部協議預警、提前預警、即時清算地圖。到期不會自動扣費，沒續費就回到免費的 Starter。",
	UnknownCommand:    "沒有這個指令。",
	RegisterFailed:    "註冊失敗，請稍後再試。",
	Registered:        "已註冊。",
	AlreadyRegistered: "你已經註冊過了。",
	QueryFailed:       "查詢失敗，請稍後再試。",
	AddUsage:          "用法: /add 0x 開頭的錢包地址",
	AddFailed:         "新增失敗，請稍後再試。",
	AddExists:         "這個地址已經在監控中。",
	AddLimit:          "目前版本最多 %d 個地址。",
	RemoveUsage:       "用法: /remove 0x 開頭的錢包地址",
	Removed:           "已刪除 %s",
	RemoveMissing:     "監控列表裡沒有這個地址。",
	RemoveFailed:      "刪除失敗，請稍後再試。",
	ThresholdMoved:    "預警線現在按地址設定，每個地址預設 1.2，請在網頁控制台的地址卡片上修改。",
	NoAddresses:       "還沒有監控地址。用 /add 0x... 新增。",
	Paused:            "\n\n（超出 Starter 額度，已暫停預警）",

	LoginPrivate: "請在和機器人的私聊裡確認登入。",
	LoginInvalid: "登入連結無效或已過期。回到網頁重新點「用 Telegram 登入」。",
	LoginPrompt: `你正在登入 MechBot 網頁控制台。

網頁上顯示的驗證碼: %s

驗證碼一致再點「確認登入」。如果不是你本人剛在網頁上點的登入，不要確認。`,
	LoginButton:       "確認登入",
	LoginExpiredToast: "登入連結已過期",
	LoginExpired:      "登入連結已過期。回到網頁重新點「用 Telegram 登入」。",
	LoginFailedToast:  "登入失敗，請重試",
	LoginUsedToast:    "登入連結已失效",
	LoginUsed:         "這個登入連結已經用過或已過期。",
	LoginConfirmToast: "已確認",
	LoginConfirmed:    "已確認登入。回到網頁即可，頁面會自動進入控制台。",

	LangPrompt: "選擇機器人訊息的語言：",
	LangSet:    "好的，之後的訊息會用繁體中文傳送。",
}
