package notify

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"mechbot/internal/i18n"
	"mechbot/internal/protocol"
)

const signature = "\nStay calm, stay ahead."

func Danger(l i18n.Lang, address string, threshold float64, pos protocol.Position) string {
	t := i18n.T(l)
	var b strings.Builder
	fmt.Fprintf(&b, t.DangerHead, pos.Label, address, pos.HealthFactor, threshold, USD(pos.CollateralUSD), USD(pos.DebtUSD))
	if drop := pos.ETHDrop(); drop > 0 {
		fmt.Fprintf(&b, t.ETHDrop, drop*100, USD(pos.LiqETHPrice))
	}
	b.WriteString(rescueLines(t, threshold, pos))
	b.WriteString(t.DangerFoot + signature)
	return b.String()
}

// Early warns from live market prices before the on-chain oracle has moved the health factor below the line.
func Early(l i18n.Lang, address string, threshold float64, pos protocol.Position, estimated, ethNow, ethAtRead float64) string {
	t := i18n.T(l)
	var b strings.Builder
	fmt.Fprintf(&b, t.EarlyHead, pos.Label, address, pos.HealthFactor, estimated, threshold)
	if ethAtRead > 0 {
		fmt.Fprintf(&b, t.EarlyETH, USD(ethAtRead), USD(ethNow), (ethNow/ethAtRead-1)*100)
	}
	b.WriteString(t.EarlyNote)
	projected := pos
	projected.HealthFactor = estimated
	b.WriteString(rescueLines(t, threshold, projected))
	b.WriteString(signature)
	return b.String()
}

func RateHigh(l i18n.Lang, address string, pos protocol.Position) string {
	t := i18n.T(l)
	debt := pos.Debt.Symbol
	if debt == "" {
		debt = t.BorrowWord
	}
	return fmt.Sprintf(t.RateHigh, pos.Label, address, debt, pos.BorrowAPY*100, USD(pos.DebtUSD), USD(pos.DebtUSD*pos.BorrowAPY)) + signature
}

func RateCalm(l i18n.Lang, address string, pos protocol.Position) string {
	return fmt.Sprintf(i18n.T(l).RateCalm, pos.Label, address, pos.BorrowAPY*100)
}

// Depeg warns that an asset in the position has drifted off its peg. debtSide is true when the
// asset is borrowed: a debt asset trading above peg makes the debt bigger, collateral below peg
// makes the collateral smaller; either way the health factor falls once oracles follow.
func Depeg(l i18n.Lang, address string, pos protocol.Position, symbol, coin, peg string, price, deviation float64, debtSide bool) string {
	t := i18n.T(l)
	side, effect := t.DepegSideCollateral, t.DepegEffectCollateral
	if debtSide {
		side, effect = t.DepegSideDebt, t.DepegEffectDebt
	}
	if (deviation > 0) != debtSide {
		effect = t.DepegEffectFavorable
	} else {
		effect += t.DepegEffectHF
	}
	return fmt.Sprintf(t.Depeg, pos.Label, address, side, symbol, coin, price, peg, deviation*100, pos.HealthFactor, effect, appLine(t, pos)) + signature
}

func DepegCalm(l i18n.Lang, address string, pos protocol.Position, coin, peg string, price float64) string {
	return fmt.Sprintf(i18n.T(l).DepegCalm, coin, pos.Label, address, price, peg)
}

func appLine(t *i18n.Text, pos protocol.Position) string {
	if pos.AppURL == "" {
		return ""
	}
	return "\n" + fmt.Sprintf(t.AppLine, pos.AppURL)
}

// rescueLines tells the user how much to repay or add to get back to a comfortable health factor.
func rescueLines(t *i18n.Text, threshold float64, pos protocol.Position) string {
	r, ok := pos.Rescue(threshold)
	if !ok {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, t.RescueHead, r.Target)
	fmt.Fprintf(&b, t.RescueRepay, tokenOrUSD(t, r.RepayAmount, pos.Debt.Symbol, r.RepayUSD))
	if r.AddUSD > 0 {
		fmt.Fprintf(&b, t.RescueAdd, tokenOrUSD(t, r.AddAmount, pos.Collateral.Symbol, r.AddUSD))
	}
	if pos.AppURL != "" {
		b.WriteString(fmt.Sprintf(t.AppLine, pos.AppURL) + "\n")
	}
	return b.String()
}

func tokenOrUSD(t *i18n.Text, amount float64, symbol string, usd float64) string {
	if amount <= 0 || symbol == "" || symbol == "?" {
		return USD(usd)
	}
	if strings.Contains(strings.ToUpper(symbol), "USD") || strings.Contains(strings.ToUpper(symbol), "DAI") {
		return Amount(amount) + " " + symbol
	}
	return fmt.Sprintf(t.TokenApprox, Amount(amount), symbol, USD(usd))
}

func Safe(l i18n.Lang, address string, threshold float64, pos protocol.Position) string {
	t := i18n.T(l)
	if !pos.HasDebt {
		return fmt.Sprintf(t.SafeNoDebt, pos.Label, address)
	}
	return fmt.Sprintf(t.Safe, pos.Label, address, pos.HealthFactor, threshold, USD(pos.CollateralUSD), USD(pos.DebtUSD))
}

func List(l i18n.Lang, address string, threshold float64, results []protocol.Result) string {
	t := i18n.T(l)
	var b strings.Builder
	fmt.Fprintf(&b, t.ListHead, address, threshold)
	var quiet []string
	for _, result := range results {
		if result.Failure != nil {
			fmt.Fprintf(&b, t.ListFailed, result.Failure.Label, result.Failure.Err)
			continue
		}
		pos := result.Position
		if !pos.HasDebt {
			if pos.Label != "" {
				quiet = append(quiet, pos.Label)
			}
			continue
		}
		fmt.Fprintf(&b, t.ListPos, pos.Label, pos.HealthFactor, USD(pos.CollateralUSD), USD(pos.DebtUSD))
		if drop := pos.ETHDrop(); drop > 0 {
			fmt.Fprintf(&b, t.ListDrop, drop*100, USD(pos.LiqETHPrice))
		}
		if pos.BorrowAPY > 0 {
			fmt.Fprintf(&b, t.ListAPY, pos.BorrowAPY*100)
		}
	}
	if len(quiet) > 0 {
		fmt.Fprintf(&b, t.ListQuiet, strings.Join(quiet, t.ListSep))
	}
	return strings.TrimSpace(b.String())
}

// Added answers /add with what the first check found, so an address without debt is not mistaken for a failed lookup.
func Added(l i18n.Lang, address string, threshold float64, results []protocol.Result, pro bool) string {
	t := i18n.T(l)
	hasDebt, failed := false, false
	for _, result := range results {
		if result.Failure != nil {
			failed = true
		} else if result.Position.HasDebt {
			hasDebt = true
		}
	}
	if hasDebt {
		return t.AddedDebt + List(l, address, threshold, results)
	}
	if failed {
		return fmt.Sprintf(t.AddedFailed, address)
	}
	text := fmt.Sprintf(t.AddedQuiet, address)
	if !pro {
		text += t.AddedStarter
	}
	return text
}

// ProEnding warns a day before Pro, or the free trial, runs out.
func ProEnding(l i18n.Lang, trial bool, freeLimit int, renewURL string) string {
	t := i18n.T(l)
	what := t.ProEndingPaid
	if trial {
		what = t.ProEndingTrial
	}
	return fmt.Sprintf(t.ProEnding, what, freeLimit, t.StarterLimits, renewURL)
}

// ProEnded tells the user what changed when Pro ran out.
func ProEnded(l i18n.Lang, trial bool, addresses, freeLimit int, renewURL string) string {
	t := i18n.T(l)
	title := t.ProEndedPaid
	if trial {
		title = t.ProEndedTrial
	}
	watch := fmt.Sprintf(t.ProEndedWatch, freeLimit)
	if addresses > freeLimit {
		watch += fmt.Sprintf(t.ProEndedPaused, addresses-freeLimit)
	}
	return fmt.Sprintf(t.ProEnded, title, watch, t.StarterLimits, renewURL)
}

// USD formats dollars with thousands separators, e.g. $12,345.67.
func USD(v float64) string {
	return "$" + commas(v, 2)
}

// Amount formats a token amount with precision that suits its size.
func Amount(v float64) string {
	switch {
	case v >= 100:
		return commas(v, 0)
	case v >= 1:
		return commas(v, 2)
	default:
		return strconv.FormatFloat(v, 'f', 4, 64)
	}
}

func commas(v float64, decimals int) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "0"
	}
	s := strconv.FormatFloat(math.Abs(v), 'f', decimals, 64)
	intPart, frac := s, ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart, frac = s[:i], s[i:]
	}
	var b strings.Builder
	if v < 0 {
		b.WriteByte('-')
	}
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String() + frac
}
