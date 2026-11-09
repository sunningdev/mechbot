package i18n

// Text is one language's copy of every user-facing message. Fields ending in a format verb list are fmt templates;
// every language must use the same verbs in the same order (checked by the tests).
type Text struct {
	// Liquidation alert: label, address, health factor, alert line, collateral, debt.
	DangerHead string
	// ETH drop to liquidation: percent, ETH liquidation price.
	ETHDrop    string
	DangerFoot string

	// Early alert: label, address, on-chain HF, estimated HF, alert line.
	EarlyHead string
	// ETH move: price at read, price now, change percent.
	EarlyETH  string
	EarlyNote string

	// Rate alert: label, address, debt symbol, APY percent, debt, extra interest per year.
	RateHigh   string
	BorrowWord string // used when the debt symbol is unknown
	// Rate back to normal: label, address, APY percent.
	RateCalm string

	DepegSideCollateral   string
	DepegSideDebt         string
	DepegEffectCollateral string
	DepegEffectDebt       string
	DepegEffectFavorable  string
	DepegEffectHF         string // appended to the unfavorable effect
	// Depeg alert: label, address, side, symbol, coin, price, peg, deviation percent, HF, effect, app line.
	Depeg string
	// Depeg recovered: coin, label, address, price, peg.
	DepegCalm string

	AppLine     string // fix link: URL
	RescueHead  string // target HF
	RescueRepay string // amount
	RescueAdd   string // amount
	TokenApprox string // amount, symbol, USD value

	// Back to safe without debt: label, address.
	SafeNoDebt string
	// Back to safe: label, address, HF, alert line, collateral, debt.
	Safe string

	ListHead   string // address, alert line
	ListFailed string // label, error
	ListPos    string // label, HF, collateral, debt
	ListDrop   string // percent, ETH liquidation price
	ListAPY    string // APY percent
	ListQuiet  string // labels
	ListSep    string

	AddedDebt    string
	AddedFailed  string // address
	AddedQuiet   string // address
	AddedStarter string

	StarterLimits     string
	ProEndingPaid     string
	ProEndingTrial    string
	ProEnding         string // what, free limit, starter limits, renew URL
	ProEndedPaid      string
	ProEndedTrial     string
	ProEndedWatch     string // free limit
	ProEndedPaused    string // paused count
	ProEnded          string // title, watch line, starter limits, renew URL
	StatusStarter     string
	StatusPro         string
	StatusTrial       string
	StatusNone        string
	Status            string // plan, address count, limit, expiry, referrer
	PlanMonth         string
	PlanQuarter       string
	PlanYear          string
	PayTooLow         string // amount, network, cheapest price
	PayBonus          string // days
	PayDone           string // amount, network, plan, days, bonus line, expiry
	PayReferrer       string // days, plan, expiry
	Help              string
	HelpConsole       string // URL
	TrialNote         string // trial days
	UnknownCommand    string
	RegisterFailed    string
	Registered        string
	AlreadyRegistered string
	QueryFailed       string
	AddUsage          string
	AddFailed         string
	AddExists         string
	AddLimit          string // limit
	RemoveUsage       string
	Removed           string // address
	RemoveMissing     string
	RemoveFailed      string
	ThresholdMoved    string
	NoAddresses       string
	Paused            string
	LoginPrivate      string
	LoginInvalid      string
	LoginPrompt       string // code
	LoginButton       string
	LoginExpiredToast string
	LoginExpired      string
	LoginFailedToast  string
	LoginUsedToast    string
	LoginUsed         string
	LoginConfirmToast string
	LoginConfirmed    string
	LangPrompt        string
	LangSet           string
}
