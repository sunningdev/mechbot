import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, type LiqMap as LiqMapData, type LiqSource } from "../api";
import LiqMapChart, { dropPct, liquidatedWithin, LiqLegend, walls } from "../components/LiqMapChart";
import PageHead from "../components/PageHead";
import { shortAddress, usdShort, usdWhole } from "../format";
import { tr } from "../i18n";

const REFRESH_MS = 60_000;

const terms = () => [
  {
    term: tr("清算墙", "Liquidation wall"),
    body: tr(
      "某个价格附近，集中了一大批可能被清算的杠杆仓位。价格跌到这里会触发集中抛售，可能加速下跌。图上金额最大的 3 档标为清算墙。",
      "A price level where many leveraged positions could be liquidated at once. A drop into it triggers clustered selling that can speed up the fall. The 3 largest bands on the chart are marked as walls.",
    ),
  },
  {
    term: tr("清算价", "Liquidation price"),
    body: tr(
      "ETH 跌到这个价格，仓位的健康因子降到 1，就会被清算人强制卖掉部分抵押品还债。",
      "The ETH price at which a position's health factor hits 1 and liquidators sell part of its collateral to repay the debt.",
    ),
  },
  {
    term: tr("清算规模", "Liquidation size"),
    body: tr(
      "每一档里，按清算那一刻的价格计算的抵押品总价值，也就是可能被抛售的量。",
      "The total collateral in a band, valued at the liquidation price: roughly how much could be sold.",
    ),
  },
  {
    term: tr("跌幅", "Drop"),
    body: tr("每行右侧的百分比，表示这一档距离当前 ETH 价格还要跌多少。", "The percentage on each row: how far ETH must fall from now to reach that band."),
  },
];

function updatedAt(map: LiqMapData) {
  const min = Math.max(0, Math.round((Date.now() / 1000 - map.at) / 60));
  const ago = min < 1 ? tr("刚刚", "just now") : tr(`${min} 分钟前`, `${min} min ago`);
  return map.pro ? tr(`${ago}更新`, `Updated ${ago}`) : tr(`延迟数据 · ${ago}`, `Delayed · ${ago}`);
}

function sourceState(s: LiqSource) {
  switch (s.state) {
    case "ok":
      return tr(`${s.positions} 个仓位`, `${s.positions} positions`);
    case "syncing":
      return s.progress > 0 && s.progress < 1
        ? tr(`收集借款人 ${Math.floor(s.progress * 100)}%`, `Collecting borrowers ${Math.floor(s.progress * 100)}%`)
        : tr("读取中", "Reading");
    case "error":
      return tr("暂时读取失败，稍后重试", "Read failed, retrying later");
    case "disabled":
      return tr("未开启", "Not enabled");
    default:
      return tr("排队中", "Queued");
  }
}

export default function LiqMap() {
  const [map, setMap] = useState<LiqMapData | null>(null);
  const [failed, setFailed] = useState(false);
  const [mine, setMine] = useState<number[]>([]);

  useEffect(() => {
    const load = () =>
      api.liqmap().then(
        (m) => {
          setMap(m);
          setFailed(false);
        },
        () => setFailed(true),
      );
    void load();
    const timer = window.setInterval(load, REFRESH_MS);
    return () => window.clearInterval(timer);
  }, []);

  useEffect(() => {
    api.addresses().then(
      (list) =>
        setMine(
          list
            .filter((a) => a.active)
            .flatMap((a) => a.positions)
            .map((p) => p.liqEthPrice ?? 0)
            .filter((p) => p > 0),
        ),
      () => undefined,
    );
  }, []);

  const ready = map?.ready ? map : null;
  const wall = ready ? walls(ready.buckets, 1)[0] : undefined;
  const within = (d: number) => (ready ? usdShort(liquidatedWithin(ready, d)) : "—");

  return (
    <>
      <PageHead
        title={tr("ETH 清算地图", "ETH liquidation map")}
        sub={tr(
          "ETH 跌到哪个价位，会有多少借贷仓位被清算。统计 Ethereum 上 Aave V3 和 Morpho Blue。",
          "How much lending collateral gets liquidated at each ETH price. Covers Aave V3 and Morpho Blue on Ethereum.",
        )}
        stats={[
          {
            label: tr("ETH 价格", "ETH price"),
            value: ready ? usdWhole(ready.ethPrice) : "—",
            hint: ready ? updatedAt(ready) : undefined,
            tag: map ? (map.pro ? { text: tr("实时", "Live"), tone: "ok" } : { text: tr("延迟 1 小时", "1h delay"), tone: "danger" }) : undefined,
          },
          {
            label: tr("跌 10% 内清算", "Liquidated within −10%"),
            value: within(0.1),
            hint: tr("被清算的抵押品", "Collateral liquidated"),
          },
          {
            label: tr("跌 20% 内清算", "Liquidated within −20%"),
            value: within(0.2),
            hint: tr("被清算的抵押品", "Collateral liquidated"),
          },
          {
            label: tr("最大清算墙", "Largest wall"),
            value: ready && wall ? <span className="hf danger">{usdWhole(wall.price)}</span> : "—",
            hint: ready && wall ? `${usdShort(wall.total)} · ${dropPct(wall.price, ready.ethPrice)}%` : undefined,
          },
        ]}
      />

      <main className="container dashboard">
        <div className="col-main">
          <section className="panel">
            <div className="panel-head">
              <h2>{tr("清算分布", "Liquidation levels")}</h2>
              {ready && <LiqLegend pro={ready.pro} mine={mine.length > 0} />}
            </div>
            {ready ? (
              <>
                <LiqMapChart map={ready} mine={mine} />
                <p className="small-text muted liq-foot">
                  {tr(
                    `每行 ${usdWhole(ready.bucket)} 一档。柱子越长，ETH 跌到这一档时被清算的抵押越多；清算会抛售抵押品，密集区可能加速下跌。只统计借款 1 万美元以上的仓位，按当前仓位估算，仅供参考。`,
                    `Each row is a ${usdWhole(ready.bucket)} band. Longer bars mean more collateral is liquidated if ETH falls there; liquidations sell collateral, so dense bands can speed up a drop. Positions with at least $10K borrowed, estimated from current balances. For reference only.`,
                  )}
                </p>
                <dl className="liq-terms">
                  {terms().map((t) => (
                    <div key={t.term}>
                      <dt>{t.term}</dt>
                      <dd>{t.body}</dd>
                    </div>
                  ))}
                </dl>
              </>
            ) : (
              <p className="muted">
                {failed
                  ? tr("清算地图暂时无法加载，稍后再试。", "The map can't load right now. Try again later.")
                  : map
                    ? tr("首轮数据收集中，几分钟后刷新查看。", "Collecting the first round of data. Check back in a few minutes.")
                    : tr("加载中…", "Loading…")}
              </p>
            )}
          </section>
        </div>

        <div className="col-side">
          {map?.pro ? (
            <section className="panel">
              <div className="panel-head">
                <h2>{tr("大额仓位", "Largest positions")}</h2>
                <span className="micro">TOP {map.top.length}</span>
              </div>
              {map.top.length ? (
                <ol className="liq-top">
                  {map.top.map((p) => (
                    <li key={p.market + p.address}>
                      <div>
                        <b>{usdWhole(p.liq)}</b>
                        <span className="liq-top-drop">{dropPct(p.liq, map.ethPrice)}%</span>
                      </div>
                      <div className="muted small-text">
                        {p.market} ·{" "}
                        <a href={`https://etherscan.io/address/${p.address}`} target="_blank" rel="noreferrer" className="mono">
                          {shortAddress(p.address)}
                        </a>
                      </div>
                      <div className="small-text">
                        {tr("抵押", "Collateral")} {usdShort(p.coll)} · {tr("借款", "Debt")} {usdShort(p.debt)}
                      </div>
                    </li>
                  ))}
                </ol>
              ) : (
                <p className="muted">{tr("暂无数据", "No data yet")}</p>
              )}
            </section>
          ) : (
            <section className="panel liq-upsell">
              <div className="panel-head">
                <h2>{tr("Pro 实时清算地图", "Live map with Pro")}</h2>
              </div>
              <ul className="plan-features">
                <li>{tr("实时数据，约 15 分钟刷新（免费版延迟 1 小时）", "Live data, refreshed about every 15 min (free is 1h delayed)")}</li>
                <li>{tr("$25 精细价位（免费版 $100）", "$25 price bands (free is $100)")}</li>
                <li>{tr("Aave、Morpho 分开显示", "Aave and Morpho shown separately")}</li>
                <li>{tr("大额仓位 Top 20：清算价、抵押、地址", "Top 20 positions: liquidation price, collateral, address")}</li>
              </ul>
              <Link className="btn btn-primary btn-block" to="/app/pro">
                {tr("升级 Pro", "Upgrade to Pro")} <span aria-hidden="true">→</span>
              </Link>
            </section>
          )}

          {map && (
            <section className="panel">
              <div className="panel-head">
                <h2>{tr("数据来源", "Sources")}</h2>
              </div>
              <ul className="liq-sources">
                {map.sources.map((s) => (
                  <li key={s.key}>
                    <span>
                      <i className={`dot ${s.state === "ok" ? "ok" : s.state === "error" ? "danger" : ""}`} />
                      {s.label}
                    </span>
                    <span className="muted small-text">{sourceState(s)}</span>
                  </li>
                ))}
              </ul>
              <p className="small-text muted">
                {tr(
                  "Ethereum 主网。抵押为 ETH 系资产（WETH、wstETH、weETH 等）、借稳定币等非 ETH 资产的仓位才会因 ETH 下跌被清算。",
                  "Ethereum mainnet. Only positions with ETH-family collateral (WETH, wstETH, weETH…) borrowing non-ETH assets get liquidated by an ETH drop.",
                )}
              </p>
            </section>
          )}
        </div>
      </main>
    </>
  );
}
