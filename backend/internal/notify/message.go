package notify

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"mechbot/internal/protocol"
)

func Danger(address string, threshold float64, pos protocol.Position) string {
	var b strings.Builder
	fmt.Fprintf(&b, `⚠️ MechBot 清算预警

协议: %s
地址: %s
健康因子: %.4f（预警线 %.2f）
抵押品: %s
债务: %s
`, pos.Label, address, pos.HealthFactor, threshold, USD(pos.CollateralUSD), USD(pos.DebtUSD))
	if drop := pos.ETHDrop(); drop > 0 {
		fmt.Fprintf(&b, "ETH 再跌 %.1f%% 会被清算（ETH 清算价 %s）\n", drop*100, USD(pos.LiqETHPrice))
	}
	b.WriteString(rescueLines(threshold, pos))
	b.WriteString("\n健康因子低于 1.0 时仓位可能被清算。\nStay calm, stay ahead.")
	return b.String()
}

// Early warns from live market prices before the on-chain oracle has moved the health factor below the line.
func Early(address string, threshold float64, pos protocol.Position, estimated, ethNow, ethAtRead float64) string {
	var b strings.Builder
	fmt.Fprintf(&b, `⏱ MechBot 提前预警（Pro）

协议: %s
地址: %s
链上健康因子: %.4f
按币安实时价格预计: %.4f（预警线 %.2f）
`, pos.Label, address, pos.HealthFactor, estimated, threshold)
	if ethAtRead > 0 {
		fmt.Fprintf(&b, "ETH: %s → %s（%+.1f%%）\n", USD(ethAtRead), USD(ethNow), (ethNow/ethAtRead-1)*100)
	}
	b.WriteString("\n链上预言机通常比交易所慢，更新后就会跌破预警线。现在处理还来得及。\n")
	projected := pos
	projected.HealthFactor = estimated
	b.WriteString(rescueLines(threshold, projected))
	b.WriteString("\nStay calm, stay ahead.")
	return b.String()
}

func RateHigh(address string, pos protocol.Position) string {
	debt := pos.Debt.Symbol
	if debt == "" {
		debt = "借款"
	}
	return fmt.Sprintf(`📈 MechBot 借款利率提醒（Pro）

协议: %s
地址: %s
%s 借款年化: %.2f%%
债务: %s，按这个利率一年约多付 %s

利率突然升高通常是资金池被借空。循环贷收益可能已经变负，可以考虑先还一部分或换个市场。
Stay calm, stay ahead.`, pos.Label, address, debt, pos.BorrowAPY*100, USD(pos.DebtUSD), USD(pos.DebtUSD*pos.BorrowAPY))
}

func RateCalm(address string, pos protocol.Position) string {
	return fmt.Sprintf(`✅ MechBot 借款利率已回落

协议: %s
地址: %s
借款年化: %.2f%%`, pos.Label, address, pos.BorrowAPY*100)
}

// Depeg warns that an asset in the position has drifted off its peg. debtSide is true when the
// asset is borrowed: a debt asset trading above peg makes the debt bigger, collateral below peg
// makes the collateral smaller; either way the health factor falls once oracles follow.
func Depeg(address string, pos protocol.Position, symbol, coin, peg string, price, deviation float64, debtSide bool) string {
	side, effect := "抵押物", "抵押物价值缩水"
	if debtSide {
		side, effect = "借款", "债务变大"
	}
	if (deviation > 0) != debtSide {
		effect = "对你暂时有利，但脱锚通常伴随剧烈波动"
	} else {
		effect += "，健康因子会随预言机更新下降"
	}
	return fmt.Sprintf(`⚠️ MechBot 脱锚预警（Pro）

协议: %s
地址: %s
%s %s 跟踪的 %s 当前 %.4f %s（偏离 %+.2f%%）
健康因子: %.4f

%s。建议先降低杠杆：还一部分借款或补充其他抵押物。%s
Stay calm, stay ahead.`, pos.Label, address, side, symbol, coin, price, peg, deviation*100, pos.HealthFactor, effect, appLine(pos))
}

func DepegCalm(address string, pos protocol.Position, coin, peg string, price float64) string {
	return fmt.Sprintf(`✅ MechBot %s 已回到锚定附近

协议: %s
地址: %s
当前 %.4f %s`, coin, pos.Label, address, price, peg)
}

func appLine(pos protocol.Position) string {
	if pos.AppURL == "" {
		return ""
	}
	return "\n去处理: " + pos.AppURL
}

// rescueLines tells the user how much to repay or add to get back to a comfortable health factor.
func rescueLines(threshold float64, pos protocol.Position) string {
	r, ok := pos.Rescue(threshold)
	if !ok {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\n怎么处理（回到 %.2f）：\n", r.Target)
	fmt.Fprintf(&b, "· 还款 %s\n", tokenOrUSD(r.RepayAmount, pos.Debt.Symbol, r.RepayUSD))
	if r.AddUSD > 0 {
		fmt.Fprintf(&b, "· 或补抵押 %s\n", tokenOrUSD(r.AddAmount, pos.Collateral.Symbol, r.AddUSD))
	}
	if pos.AppURL != "" {
		fmt.Fprintf(&b, "去处理: %s\n", pos.AppURL)
	}
	return b.String()
}

func tokenOrUSD(amount float64, symbol string, usd float64) string {
	if amount <= 0 || symbol == "" || symbol == "?" {
		return USD(usd)
	}
	if strings.Contains(strings.ToUpper(symbol), "USD") || strings.Contains(strings.ToUpper(symbol), "DAI") {
		return Amount(amount) + " " + symbol
	}
	return fmt.Sprintf("%s %s（约 %s）", Amount(amount), symbol, USD(usd))
}

func Safe(address string, threshold float64, pos protocol.Position) string {
	if !pos.HasDebt {
		return fmt.Sprintf(`✅ MechBot 仓位已安全

协议: %s
地址: %s

这笔借款已结清，当前没有债务。`, pos.Label, address)
	}
	return fmt.Sprintf(`✅ MechBot 仓位已安全

协议: %s
地址: %s
健康因子: %.4f（预警线 %.2f）
抵押品: %s
债务: %s

已回到预警线以上。`, pos.Label, address, pos.HealthFactor, threshold, USD(pos.CollateralUSD), USD(pos.DebtUSD))
}

func List(address string, threshold float64, results []protocol.Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "地址 %s\n预警线 %.2f\n", address, threshold)
	var quiet []string
	for _, result := range results {
		if result.Failure != nil {
			fmt.Fprintf(&b, "\n%s\n查询失败: %v\n", result.Failure.Label, result.Failure.Err)
			continue
		}
		pos := result.Position
		if !pos.HasDebt {
			if pos.Label != "" {
				quiet = append(quiet, pos.Label)
			}
			continue
		}
		fmt.Fprintf(&b, "\n%s\n健康因子 %.4f\n抵押 %s\n债务 %s\n", pos.Label, pos.HealthFactor, USD(pos.CollateralUSD), USD(pos.DebtUSD))
		if drop := pos.ETHDrop(); drop > 0 {
			fmt.Fprintf(&b, "ETH 再跌 %.1f%% 清算（%s）\n", drop*100, USD(pos.LiqETHPrice))
		}
		if pos.BorrowAPY > 0 {
			fmt.Fprintf(&b, "借款年化 %.2f%%\n", pos.BorrowAPY*100)
		}
	}
	if len(quiet) > 0 {
		fmt.Fprintf(&b, "\n无借款: %s\n", strings.Join(quiet, "，"))
	}
	return strings.TrimSpace(b.String())
}

// Added answers /add with what the first check found, so an address without debt is not mistaken for a failed lookup.
func Added(address string, threshold float64, results []protocol.Result, pro bool) string {
	hasDebt, failed := false, false
	for _, result := range results {
		if result.Failure != nil {
			failed = true
		} else if result.Position.HasDebt {
			hasDebt = true
		}
	}
	if hasDebt {
		return "已添加，当前仓位：\n\n" + List(address, threshold, results)
	}
	if failed {
		return fmt.Sprintf("已添加 %s。\n\n部分协议暂时查询失败，稍后用 /list 再看。", address)
	}
	text := fmt.Sprintf(`已添加 %s。

这个地址目前没有借款：已检查 Aave V3、Compound V3、Morpho Blue（Ethereum · Arbitrum · Optimism），不会被清算。

地址会继续监控，以后在这些协议借款，会自动开始预警，不用重新添加。`, address)
	if !pro {
		text += "\n\nStarter 只对 Aave V3 发预警，Compound、Morpho 的借款需要 Pro。"
	}
	return text
}

const starterLimits = `· 只保留 Aave V3 预警，Compound、Morpho 预警停止
· 提前预警、利率提醒、脱锚预警停止
· 清算地图改为延迟 1 小时`

// ProEnding warns a day before Pro, or the free trial, runs out.
func ProEnding(trial bool, freeLimit int, renewURL string) string {
	what := "你的 Pro 将在 24 小时内到期"
	if trial {
		what = "你的 7 天 Pro 体验将在 24 小时内结束"
	}
	return fmt.Sprintf(`⏳ MechBot %s

到期后自动降为 Starter（免费）：
· 只监控最早添加的 %d 个地址，其余暂停（不会删除）
%s

继续用 Pro：%s`, what, freeLimit, starterLimits, renewURL)
}

// ProEnded tells the user what changed when Pro ran out.
func ProEnded(trial bool, addresses, freeLimit int, renewURL string) string {
	title := "Pro 已到期，已降为 Starter"
	if trial {
		title = "Pro 体验已结束，已降为 Starter"
	}
	watch := fmt.Sprintf("· 最早添加的 %d 个地址继续监控", freeLimit)
	if addresses > freeLimit {
		watch += fmt.Sprintf("，其余 %d 个已暂停（数据保留，续费后自动恢复）", addresses-freeLimit)
	}
	return fmt.Sprintf(`MechBot %s

%s
%s

续费后立刻恢复全部功能：%s`, title, watch, starterLimits, renewURL)
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
