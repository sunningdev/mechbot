import { Link } from "react-router-dom";
import type { Position } from "../api";
import { riskClass, tokenOrUsd, usd, usdShort, usdWhole as usd0 } from "../format";
import { tr } from "../i18n";

const STRESS_DROPS = [10, 20, 30];

type Props = { position: Position; threshold: number };

// Borrow APY at or above this is flagged; it matches the backend's rate alert.
export const HIGH_APY = 0.15;

export default function MiniGauge({ position, threshold }: Props) {
  const hf = position.healthFactor;
  const safe = riskClass(hf, threshold) === "ok";
  const drop = Math.max(0, (1 - 1 / hf) * 100);
  const rescue = !safe ? position.rescue : undefined;
  const apy = position.borrowApy ?? 0;

  return (
    <div className={`mini-gauge ${safe ? "safe" : "alert"}`}>
      <div className="mg-read">
        <span className="mg-label">
          {position.label}
          {!position.covered && <span className="tag">{tr("Pro 预警", "Pro alerts")}</span>}
        </span>
        <span className="mg-hf-row">
          <strong className="mg-hf">{hf.toFixed(3)}</strong>
          <span className="mg-state">{safe ? tr("安全", "Safe") : tr("预警", "Alert")}</span>
        </span>
      </div>

      <dl className="mg-stats">
        <div>
          <dt>{tr("抵押", "Collateral")}</dt>
          <dd>{usd(position.collateralUsd)}</dd>
        </div>
        <div>
          <dt>{tr("债务", "Debt")}</dt>
          <dd>{usd(position.debtUsd)}</dd>
        </div>
        <div>
          <dt>{tr("距清算", "To liquidation")}</dt>
          <dd className="mg-tone">−{drop.toFixed(1)}%</dd>
        </div>
        <div title={tr("抵押总值跌到这个数就会被清算", "Liquidated if collateral value falls to this")}>
          <dt>{tr("清算线", "Liquidation at")}</dt>
          <dd className="mg-tone">{usd(position.collateralUsd / hf)}</dd>
        </div>
        {position.liqEthPrice ? (
          <div title={tr(`ETH 现价 ${usd(position.ethPrice ?? 0)}，跌到这个价格就会被清算`, `ETH is ${usd(position.ethPrice ?? 0)} now; liquidated if it falls to this price`)}>
            <dt>{tr("ETH 清算价", "ETH liq. price")}</dt>
            <dd className="mg-tone">{usd(position.liqEthPrice)}</dd>
          </div>
        ) : null}
        {apy > 0 && (
          <div title={tr("主要借款资产的浮动借款年化", "Variable borrow APY of the main debt asset")}>
            <dt>{tr("借款年化", "Borrow APY")}</dt>
            <dd className={apy >= HIGH_APY ? "mg-hot" : ""}>{(apy * 100).toFixed(2)}%</dd>
          </div>
        )}
      </dl>

      {position.stress && (
        <div className="mg-stress" title={tr("按当前仓位估算，其他资产价格不变", "Estimated from the current position, other prices unchanged")}>
          <span className="mg-stress-label">{tr("压力测试", "Stress test")}</span>
          {position.stress.map((v, i) => (
            <span key={i} className={`mg-stress-cell ${v < 1 ? "liq" : riskClass(v, threshold)}`}>
              <em>ETH −{STRESS_DROPS[i]}%</em>
              <b>{v < 1 ? tr("清算", "Liquidated") : v.toFixed(2)}</b>
            </span>
          ))}
        </div>
      )}

      {position.wall && (
        <p className="mg-wall">
          {tr(
            `你的清算价 ${usd0(position.liqEthPrice ?? 0)} 附近有 ${usdShort(position.wall.total)} 清算墙（${usd0(position.wall.price)}）。别人爆仓会把价格推向你，建议把健康因子提到 ${position.wall.safeHf.toFixed(2)} 以上。`,
            `A ${usdShort(position.wall.total)} liquidation wall sits near your liquidation price of ${usd0(position.liqEthPrice ?? 0)} (at ${usd0(position.wall.price)}). Other liquidations could push the price into you; consider raising your health factor above ${position.wall.safeHf.toFixed(2)}.`,
          )}
          <Link to="/app/map">{tr("看地图 →", "See map →")}</Link>
        </p>
      )}

      {rescue && (
        <p className="mg-rescue">
          <span>
            {tr("建议：还款", "Fix: repay")} <b>{tokenOrUsd(rescue.repayAmount, rescue.repaySymbol, rescue.repayUsd)}</b>
            {rescue.addUsd ? (
              <>
                {tr(" 或补抵押", " or add")} <b>{tokenOrUsd(rescue.addAmount, rescue.addSymbol, rescue.addUsd)}</b>
              </>
            ) : null}
            {tr(`，可回到 ${rescue.target.toFixed(2)}`, ` to get back to ${rescue.target.toFixed(2)}`)}
          </span>
          {position.appUrl && (
            <a href={position.appUrl} target="_blank" rel="noreferrer">
              {tr("去处理 →", "Open app →")}
            </a>
          )}
        </p>
      )}
    </div>
  );
}
