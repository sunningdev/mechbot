import type { CSSProperties } from "react";
import type { LiqBucket, LiqMap } from "../api";
import { usdShort, usdWhole } from "../format";
import { tr } from "../i18n";

const WALLS = 3;

// The few largest buckets, which is where a falling price meets the most forced selling.
export function walls(buckets: LiqBucket[], n = WALLS) {
  return [...buckets]
    .filter((b) => b.total > 0)
    .sort((a, b) => b.total - a.total)
    .slice(0, n);
}

// Collateral that liquidates before ETH falls by `drop` (0.1 = 10%).
export function liquidatedWithin(map: LiqMap, drop: number) {
  const floor = map.ethPrice * (1 - drop);
  return map.buckets.filter((b) => b.price >= floor).reduce((s, b) => s + b.total, 0);
}

export const dropPct = (price: number, eth: number) => Math.round((price / eth - 1) * 100);

// mine holds the viewer's own ETH liquidation prices, marked on the rows they fall in.
type Props = { map: LiqMap; maxRows?: number; mine?: number[] };

export default function LiqMapChart({ map, maxRows, mine = [] }: Props) {
  const rows = maxRows ? map.buckets.slice(0, maxRows) : map.buckets;
  const max = Math.max(1, ...rows.map((b) => b.total));
  const wallSet = new Set(walls(rows).map((b) => b.price));
  const split = map.pro;
  const pct = (v = 0) => ({ "--w": `${(v / max) * 100}%` }) as CSSProperties;

  return (
    <div className={`liq-chart ${maxRows ? "compact" : ""}`}>
      <div className="liq-now">
        <span className="liq-price">{usdWhole(map.ethPrice)}</span>
        <span className="liq-now-line" />
        <span className="liq-now-label">{tr("ETH 现价", "ETH now")}</span>
      </div>
      {rows.map((b) => {
        const wall = wallSet.has(b.price);
        const own = mine.some((p) => p >= b.price && p < b.price + map.bucket);
        return (
          <div className={`liq-row ${wall ? "wall" : ""} ${own ? "mine" : ""}`} key={b.price} title={tr(
            `${usdWhole(b.price)}–${usdWhole(b.price + map.bucket)}：${b.count} 个仓位，抵押 ${usdShort(b.total)}`,
            `${usdWhole(b.price)}–${usdWhole(b.price + map.bucket)}: ${b.count} positions, ${usdShort(b.total)} collateral`,
          )}>
            <span className="liq-price">{usdWhole(b.price)}</span>
            <span className="liq-track">
              {split ? (
                <>
                  <i className="liq-bar aave" style={pct(b.aave)} />
                  <i className="liq-bar morpho" style={pct(b.morpho)} />
                </>
              ) : (
                <i className="liq-bar all" style={pct(b.total)} />
              )}
            </span>
            <span className="liq-val">
              {b.total > 0 ? usdShort(b.total) : ""}
              {wall && <em>{tr("清算墙", "Wall")}</em>}
              {own && <em className="mine-tag">{tr("我的", "Mine")}</em>}
            </span>
            <span className="liq-drop">{dropPct(b.price, map.ethPrice)}%</span>
          </div>
        );
      })}
    </div>
  );
}

export function LiqLegend({ pro, mine }: { pro: boolean; mine?: boolean }) {
  return (
    <div className="liq-legend">
      {pro ? (
        <>
          <span>
            <i className="liq-swatch aave" /> Aave V3
          </span>
          <span>
            <i className="liq-swatch morpho" /> Morpho Blue
          </span>
        </>
      ) : (
        <span>
          <i className="liq-swatch all" /> {tr("全部协议", "All protocols")}
        </span>
      )}
      <span>
        <i className="liq-swatch wall" /> {tr("清算墙", "Wall")}
      </span>
      {mine && (
        <span>
          <i className="liq-swatch mine" /> {tr("我的清算价", "My liq. price")}
        </span>
      )}
    </div>
  );
}
