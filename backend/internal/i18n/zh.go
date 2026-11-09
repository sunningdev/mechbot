package i18n

var zh = Text{
	DangerHead: `⚠️ MechBot 清算预警

协议: %s
地址: %s
健康因子: %.4f（预警线 %.2f）
抵押品: %s
债务: %s
`,
	ETHDrop:    "ETH 再跌 %.1f%% 会被清算（ETH 清算价 %s）\n",
	DangerFoot: "\n健康因子低于 1.0 时仓位可能被清算。",

	EarlyHead: `⏱ MechBot 提前预警（Pro）

协议: %s
地址: %s
链上健康因子: %.4f
按币安实时价格预计: %.4f（预警线 %.2f）
`,
	EarlyETH:  "ETH: %s → %s（%+.1f%%）\n",
	EarlyNote: "\n链上预言机通常比交易所慢，更新后就会跌破预警线。现在处理还来得及。\n",

	RateHigh: `📈 MechBot 借款利率提醒（Pro）

协议: %s
地址: %s
%s 借款年化: %.2f%%
债务: %s，按这个利率一年约多付 %s

利率突然升高通常是资金池被借空。循环贷收益可能已经变负，可以考虑先还一部分或换个市场。`,
	BorrowWord: "借款",
	RateCalm: `✅ MechBot 借款利率已回落

协议: %s
地址: %s
借款年化: %.2f%%`,

	DepegSideCollateral:   "抵押物",
	DepegSideDebt:         "借款",
	DepegEffectCollateral: "抵押物价值缩水",
	DepegEffectDebt:       "债务变大",
	DepegEffectFavorable:  "对你暂时有利，但脱锚通常伴随剧烈波动",
	DepegEffectHF:         "，健康因子会随预言机更新下降",
	Depeg: `⚠️ MechBot 脱锚预警（Pro）

协议: %s
地址: %s
%s %s 跟踪的 %s 当前 %.4f %s（偏离 %+.2f%%）
健康因子: %.4f

%s。建议先降低杠杆：还一部分借款或补充其他抵押物。%s`,
	DepegCalm: `✅ MechBot %s 已回到锚定附近

协议: %s
地址: %s
当前 %.4f %s`,

	AppLine:     "去处理: %s",
	RescueHead:  "\n怎么处理（回到 %.2f）：\n",
	RescueRepay: "· 还款 %s\n",
	RescueAdd:   "· 或补抵押 %s\n",
	TokenApprox: "%s %s（约 %s）",

	SafeNoDebt: `✅ MechBot 仓位已安全

协议: %s
地址: %s

这笔借款已结清，当前没有债务。`,
	Safe: `✅ MechBot 仓位已安全

协议: %s
地址: %s
健康因子: %.4f（预警线 %.2f）
抵押品: %s
债务: %s

已回到预警线以上。`,

	ListHead:   "地址 %s\n预警线 %.2f\n",
	ListFailed: "\n%s\n查询失败: %v\n",
	ListPos:    "\n%s\n健康因子 %.4f\n抵押 %s\n债务 %s\n",
	ListDrop:   "ETH 再跌 %.1f%% 清算（%s）\n",
	ListAPY:    "借款年化 %.2f%%\n",
	ListQuiet:  "\n无借款: %s\n",
	ListSep:    "，",

	AddedDebt:   "已添加，当前仓位：\n\n",
	AddedFailed: "已添加 %s。\n\n部分协议暂时查询失败，稍后用 /list 再看。",
	AddedQuiet: `已添加 %s。

这个地址目前没有借款：已检查 Aave V3、Compound V3、Morpho Blue（Ethereum · Arbitrum · Optimism），不会被清算。

地址会继续监控，以后在这些协议借款，会自动开始预警，不用重新添加。`,
	AddedStarter: "\n\nStarter 只对 Aave V3 发预警，Compound、Morpho 的借款需要 Pro。",

	StarterLimits: `· 只保留 Aave V3 预警，Compound、Morpho 预警停止
· 提前预警、利率提醒、脱锚预警停止
· 清算地图改为延迟 1 小时`,
	ProEndingPaid:  "你的 Pro 将在 24 小时内到期",
	ProEndingTrial: "你的 7 天 Pro 体验将在 24 小时内结束",
	ProEnding: `⏳ MechBot %s

到期后自动降为 Starter（免费）：
· 只监控最早添加的 %d 个地址，其余暂停（不会删除）
%s

继续用 Pro：%s`,
	ProEndedPaid:   "Pro 已到期，已降为 Starter",
	ProEndedTrial:  "Pro 体验已结束，已降为 Starter",
	ProEndedWatch:  "· 最早添加的 %d 个地址继续监控",
	ProEndedPaused: "，其余 %d 个已暂停（数据保留，续费后自动恢复）",
	ProEnded: `MechBot %s

%s
%s

续费后立刻恢复全部功能：%s`,

	StatusStarter: "Starter（永久免费，仅 Aave V3 预警）",
	StatusPro:     "Pro（全协议预警）",
	StatusTrial:   "Pro 体验（全协议预警，到期回到 Starter）",
	StatusNone:    "无",
	Status: `MechBot 账户状态

版本: %s
监控地址: %d/%d
Pro 到期: %s
推荐人: %s`,

	PlanMonth:   "Pro 月付",
	PlanQuarter: "Pro 季付",
	PlanYear:    "Pro 年付",
	PayTooLow:   "⚠️ 收到 %s USDT（%s）\n\n低于最低套餐 %d USDT，未能开通 Pro。如有疑问请直接回复机器人。",
	PayBonus:    "\n🎁 好友推荐奖励 +%d 天",
	PayDone:     "✅ MechBot Pro 已开通\n\n收到 %s USDT（%s），%s %d 天%s\nPro 到期: %s",
	PayReferrer: "🎁 推荐奖励 +%d 天\n\n你推荐的用户首次购买了%s。\nPro 到期: %s",

	Help: `MechBot 清算预警

/add 0x... 添加监控地址（Starter 3 个，Pro 20 个）
/list 查看各协议健康因子
/remove 0x... 删除地址
/status 查看账户状态
/lang 切换消息语言

每个地址的预警线默认 1.2，可在网页控制台的地址卡片上修改。

Starter（永久免费）预警 Aave V3（Ethereum、Arbitrum、Optimism）。Pro 额外预警 Compound V3 和 Morpho Blue。`,
	HelpConsole:       "\n\n网页控制台: %s",
	TrialNote:         "\n\n🎁 新用户送 %d 天 Pro 体验：全部协议预警、提前预警、实时清算地图。到期不会自动扣费，没续费就回到免费的 Starter。",
	UnknownCommand:    "没有这个命令。",
	RegisterFailed:    "注册失败，请稍后再试。",
	Registered:        "已注册。",
	AlreadyRegistered: "你已经注册过了。",
	QueryFailed:       "查询失败，请稍后再试。",
	AddUsage:          "用法: /add 0x 开头的钱包地址",
	AddFailed:         "添加失败，请稍后再试。",
	AddExists:         "这个地址已经在监控里。",
	AddLimit:          "当前版本最多 %d 个地址。",
	RemoveUsage:       "用法: /remove 0x 开头的钱包地址",
	Removed:           "已删除 %s",
	RemoveMissing:     "监控列表里没有这个地址。",
	RemoveFailed:      "删除失败，请稍后再试。",
	ThresholdMoved:    "预警线现在按地址设置，每个地址默认 1.2，请在网页控制台的地址卡片上修改。",
	NoAddresses:       "还没有监控地址。用 /add 0x... 添加。",
	Paused:            "\n\n（超出 Starter 额度，已暂停预警）",

	LoginPrivate: "请在和机器人的私聊里确认登录。",
	LoginInvalid: "登录链接无效或已过期。回到网页重新点「用 Telegram 登录」。",
	LoginPrompt: `你正在登录 MechBot 网页控制台。

网页上显示的验证码: %s

验证码一致再点「确认登录」。如果不是你本人刚在网页上点的登录，不要确认。`,
	LoginButton:       "确认登录",
	LoginExpiredToast: "登录链接已过期",
	LoginExpired:      "登录链接已过期。回到网页重新点「用 Telegram 登录」。",
	LoginFailedToast:  "登录失败，请重试",
	LoginUsedToast:    "登录链接已失效",
	LoginUsed:         "这个登录链接已经用过或已过期。",
	LoginConfirmToast: "已确认",
	LoginConfirmed:    "已确认登录。回到网页即可，页面会自动进入控制台。",

	LangPrompt: "选择机器人消息的语言：",
	LangSet:    "好的，之后的消息会用简体中文发送。",
}
