package i18n

var en = Text{
	DangerHead: `⚠️ MechBot liquidation alert

Protocol: %s
Address: %s
Health factor: %.4f (alert line %.2f)
Collateral: %s
Debt: %s
`,
	ETHDrop:    "Liquidated if ETH falls another %.1f%% (ETH liquidation price %s)\n",
	DangerFoot: "\nA position can be liquidated once its health factor drops below 1.0.",

	EarlyHead: `⏱ MechBot early alert (Pro)

Protocol: %s
Address: %s
On-chain health factor: %.4f
Estimated from live Binance prices: %.4f (alert line %.2f)
`,
	EarlyETH:  "ETH: %s → %s (%+.1f%%)\n",
	EarlyNote: "\nOn-chain oracles usually lag exchanges, so this will cross your alert line once they update. There is still time to act.\n",

	RateHigh: `📈 MechBot borrow rate alert (Pro)

Protocol: %s
Address: %s
%s borrow APY: %.2f%%
Debt: %s, about %s extra per year at this rate

A sudden rate spike usually means the pool has been borrowed dry. Your loop may already be losing money; consider repaying some or moving to another market.`,
	BorrowWord: "Borrow",
	RateCalm: `✅ MechBot borrow rate is back down

Protocol: %s
Address: %s
Borrow APY: %.2f%%`,

	DepegSideCollateral:   "Collateral",
	DepegSideDebt:         "Debt",
	DepegEffectCollateral: "Your collateral is worth less",
	DepegEffectDebt:       "Your debt is getting bigger",
	DepegEffectFavorable:  "This helps you for now, but depegs usually come with sharp swings",
	DepegEffectHF:         ", and your health factor will fall as oracles update",
	Depeg: `⚠️ MechBot depeg alert (Pro)

Protocol: %s
Address: %s
%s %s, which tracks %s, is at %.4f %s (%+.2f%% off peg)
Health factor: %.4f

%s. Consider lowering leverage: repay part of the loan or add other collateral.%s`,
	DepegCalm: `✅ MechBot %s is back near its peg

Protocol: %s
Address: %s
Now %.4f %s`,

	AppLine:     "Fix it: %s",
	RescueHead:  "\nWhat to do (back to %.2f):\n",
	RescueRepay: "· Repay %s\n",
	RescueAdd:   "· Or add %s collateral\n",
	TokenApprox: "%s %s (about %s)",

	SafeNoDebt: `✅ MechBot position is safe

Protocol: %s
Address: %s

This loan has been repaid; there is no debt left.`,
	Safe: `✅ MechBot position is safe

Protocol: %s
Address: %s
Health factor: %.4f (alert line %.2f)
Collateral: %s
Debt: %s

Back above your alert line.`,

	ListHead:   "Address %s\nAlert line %.2f\n",
	ListFailed: "\n%s\nLookup failed: %v\n",
	ListPos:    "\n%s\nHealth factor %.4f\nCollateral %s\nDebt %s\n",
	ListDrop:   "Liquidated if ETH falls %.1f%% (%s)\n",
	ListAPY:    "Borrow APY %.2f%%\n",
	ListQuiet:  "\nNo debt: %s\n",
	ListSep:    ", ",

	AddedDebt:   "Added. Current positions:\n\n",
	AddedFailed: "Added %s.\n\nSome protocols couldn't be checked right now. Try /list again in a moment.",
	AddedQuiet: `Added %s.

This address has no loans right now: we checked Aave V3, Compound V3 and Morpho Blue (Ethereum · Arbitrum · Optimism), so it can't be liquidated.

We'll keep watching it. If it borrows on these protocols later, alerts start automatically; no need to add it again.`,
	AddedStarter: "\n\nStarter only alerts on Aave V3. Compound and Morpho loans need Pro.",

	StarterLimits: `· Only Aave V3 alerts remain; Compound and Morpho alerts stop
· Early, borrow rate and depeg alerts stop
· The liquidation map switches to a 1-hour delay`,
	ProEndingPaid:  "Your Pro ends within 24 hours",
	ProEndingTrial: "Your 7-day Pro trial ends within 24 hours",
	ProEnding: `⏳ MechBot: %s

After that you move to Starter (free):
· Only your %d earliest addresses stay watched; the rest are paused (not deleted)
%s

Keep Pro: %s`,
	ProEndedPaid:   "Pro has ended; you're now on Starter",
	ProEndedTrial:  "Your Pro trial has ended; you're now on Starter",
	ProEndedWatch:  "· Your %d earliest addresses are still watched",
	ProEndedPaused: "; the other %d are paused (data kept, resumes when you renew)",
	ProEnded: `MechBot: %s

%s
%s

Renew to get everything back right away: %s`,

	StatusStarter: "Starter (free forever, Aave V3 alerts only)",
	StatusPro:     "Pro (all protocols)",
	StatusTrial:   "Pro trial (all protocols, returns to Starter when it ends)",
	StatusNone:    "None",
	Status: `MechBot account

Plan: %s
Addresses: %d/%d
Pro until: %s
Referred by: %s`,

	PlanMonth:   "Pro monthly",
	PlanQuarter: "Pro quarterly",
	PlanYear:    "Pro yearly",
	PayTooLow:   "⚠️ Received %s USDT (%s)\n\nThat's below the cheapest plan (%d USDT), so Pro was not activated. Reply to the bot if you have questions.",
	PayBonus:    "\n🎁 Referral bonus +%d days",
	PayDone:     "✅ MechBot Pro activated\n\nReceived %s USDT (%s): %s, %d days%s\nPro until: %s",
	PayReferrer: "🎁 Referral reward +%d days\n\nSomeone you invited just bought %s for the first time.\nPro until: %s",

	Help: `MechBot liquidation alerts

/add 0x... Watch an address (Starter 3, Pro 20)
/list Health factors across protocols
/remove 0x... Stop watching an address
/status Account status
/lang Change message language

Each address starts with an alert line of 1.2. Change it on the address card in the web console.

Starter (free forever) alerts on Aave V3 (Ethereum, Arbitrum, Optimism). Pro adds Compound V3 and Morpho Blue.`,
	HelpConsole:       "\n\nWeb console: %s",
	TrialNote:         "\n\n🎁 New users get %d days of Pro: alerts on every protocol, early alerts and the live liquidation map. Nothing is charged when it ends; you simply return to the free Starter plan.",
	UnknownCommand:    "Unknown command.",
	RegisterFailed:    "Sign-up failed. Please try again later.",
	Registered:        "You're signed up.",
	AlreadyRegistered: "You're already signed up.",
	QueryFailed:       "Lookup failed. Please try again later.",
	AddUsage:          "Usage: /add followed by a wallet address starting with 0x",
	AddFailed:         "Couldn't add it. Please try again later.",
	AddExists:         "This address is already being watched.",
	AddLimit:          "Your plan allows up to %d addresses.",
	RemoveUsage:       "Usage: /remove followed by a wallet address starting with 0x",
	Removed:           "Removed %s",
	RemoveMissing:     "That address isn't on your watch list.",
	RemoveFailed:      "Couldn't remove it. Please try again later.",
	ThresholdMoved:    "Alert lines are now set per address (1.2 by default). Change them on the address card in the web console.",
	NoAddresses:       "No addresses yet. Add one with /add 0x...",
	Paused:            "\n\n(Over the Starter limit, alerts paused)",

	LoginPrivate: "Please confirm the login in a private chat with the bot.",
	LoginInvalid: "This login link is invalid or has expired. Go back to the website and tap “Log in with Telegram” again.",
	LoginPrompt: `You're logging in to the MechBot web console.

Code shown on the website: %s

Only tap “Confirm login” if the codes match. If you didn't just start a login on the website, don't confirm.`,
	LoginButton:       "Confirm login",
	LoginExpiredToast: "Login link expired",
	LoginExpired:      "This login link has expired. Go back to the website and tap “Log in with Telegram” again.",
	LoginFailedToast:  "Login failed, please try again",
	LoginUsedToast:    "Login link no longer valid",
	LoginUsed:         "This login link has already been used or has expired.",
	LoginConfirmToast: "Confirmed",
	LoginConfirmed:    "Login confirmed. Go back to the website; it will open the console automatically.",

	LangPrompt: "Choose the language for bot messages:",
	LangSet:    "Done. Messages will be sent in English from now on.",
}
