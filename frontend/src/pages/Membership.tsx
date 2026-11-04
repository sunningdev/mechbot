import { useState } from "react";
import BillingPanel from "../components/BillingPanel";
import PageHead from "../components/PageHead";
import PurchaseHistory from "../components/PurchaseHistory";
import { proDaysLeft } from "../format";
import { tr } from "../i18n";
import { useConsole } from "./ConsoleLayout";

const features = () => [
  { label: tr("监控地址", "Watched addresses"), starter: "3", pro: "20" },
  { label: tr("Aave V3 预警", "Aave V3 alerts"), starter: "✓", pro: "✓" },
  { label: tr("Compound V3 预警", "Compound V3 alerts"), starter: "—", pro: "✓" },
  { label: tr("Morpho Blue 预警", "Morpho Blue alerts"), starter: "—", pro: "✓" },
  { label: tr("Telegram 私聊通知", "Telegram DM alerts"), starter: "✓", pro: "✓" },
  { label: tr("还款 / 补仓建议", "Repay / add-collateral advice"), starter: "✓", pro: "✓" },
  {
    label: tr("币安实时价格提前预警，循环贷插针也来得及", "Early alerts from live Binance prices, ahead of any wick"),
    starter: "—",
    pro: "✓",
  },
  { label: tr("接近预警线每分钟复查", "Checked every minute near the line"), starter: "—", pro: "✓" },
  { label: tr("ETH 清算地图", "ETH liquidation map"), starter: tr("延迟 1 小时", "1h delayed"), pro: tr("实时", "Live") },
  {
    label: tr("借款年化超 15% 提醒，循环收益变负早知道", "Borrow APY over 15% alerts, before your loop turns negative"),
    starter: "—",
    pro: "✓",
  },
  {
    label: tr("稳定币、stETH、eETH 脱锚预警", "Depeg alerts for stablecoins, stETH and eETH"),
    starter: "—",
    pro: "✓",
  },
  { label: tr("ETH 下跌压力测试", "ETH drop stress test"), starter: "✓", pro: "✓" },
];

export default function Membership() {
  const { me, refresh } = useConsole();
  const [paidAt, setPaidAt] = useState(0);
  const proDays = proDaysLeft(me.pro, me.proExpireAt);
  const soon = proDays !== null && proDays <= 7;

  return (
    <>
      <PageHead
        title={tr("会员", "Membership")}
        sub={
          me.pro && me.trial
            ? tr(
                "你正在 7 天 Pro 体验中。到期自动回到免费的 Starter，不会扣费；现在购买，剩余体验天数会顺延。",
                "You're on a 7-day Pro trial. It returns to free Starter when it ends, with no charge. Buy now and your remaining trial days carry over.",
              )
            : !me.pro && me.proExpiredAt
              ? tr(
                  "Pro 已结束，已回到 Starter：最早添加的 3 个地址继续监控，其余已暂停。升级后立刻恢复。",
                  "Pro has ended and you're back on Starter: your first 3 addresses are still watched, the rest are paused. Upgrade to restore everything.",
                )
              : me.pro
            ? tr(
                "Pro 已开通，全协议预警。续费转到同一个专属地址即可，时间自动顺延。",
                "Pro is active with alerts on every protocol. To renew, send to the same address and your time extends automatically.",
              )
            : tr(
                "升级 Pro，解锁 Compound V3、Morpho Blue 预警和 20 个监控地址。",
                "Upgrade to Pro for Compound V3 and Morpho Blue alerts and 20 watched addresses.",
              )
        }
        stats={[
          {
            label: tr("当前方案", "Current plan"),
            value: me.pro ? (me.trial ? tr("Pro 体验", "Pro trial") : "Pro") : "Starter",
            className: me.pro ? "hf ok" : "",
            hint: me.pro ? tr("全协议预警", "All protocols") : tr("仅 Aave V3 预警", "Aave V3 alerts only"),
            tag: me.pro ? (me.trial ? { text: tr("体验中", "Trial"), tone: "ok" } : { text: tr("生效中", "Active"), tone: "ok" }) : undefined,
          },
          {
            label: tr("剩余天数", "Days left"),
            value: proDays === null ? "—" : proDays,
            className: soon ? "hf danger" : "",
            hint:
              me.pro && me.proExpireAt
                ? tr(`至 ${me.proExpireAt.slice(0, 10)}`, `Until ${me.proExpireAt.slice(0, 10)}`)
                : tr("未开通", "Not active"),
            tag: soon ? { text: tr("即将到期", "Expiring"), tone: "danger" } : undefined,
          },
          {
            label: tr("监控地址", "Watched addresses"),
            value: (
              <>
                {me.addressCount}
                <small> / {me.addressLimit}</small>
              </>
            ),
            hint: me.pro ? tr("Pro 上限 20 个", "Pro limit is 20") : tr("Pro 可监控 20 个", "Pro watches 20"),
            tag: me.addressCount >= me.addressLimit ? { text: tr("已满", "Full"), tone: "danger" } : undefined,
          },
        ]}
      />

      <main className="container dashboard">
        <div className="col-main">
          <BillingPanel
            me={me}
            onPaid={() => {
              void refresh();
              setPaidAt(Date.now());
            }}
          />
          <PurchaseHistory reload={paidAt} />
        </div>
        <div className="col-side">
          <section className="panel">
            <div className="panel-head">
              <h2>{tr("方案对比", "Compare plans")}</h2>
            </div>
            <table className="compare plans">
              <thead>
                <tr>
                  <th />
                  <th className={me.pro ? "" : "on"}>
                    Starter
                    {!me.pro && <small>{tr("当前", "Current")}</small>}
                  </th>
                  <th className="pro-col">
                    Pro
                    {me.pro && <small>{tr("当前", "Current")}</small>}
                  </th>
                </tr>
              </thead>
              <tbody>
                {features().map((f) => (
                  <tr key={f.label} className={f.starter === "—" ? "pro-only" : ""}>
                    <td>{f.label}</td>
                    <td className={f.starter === "—" ? "no" : ""}>{f.starter}</td>
                    <td className="pro-col">{f.pro}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            <p className="muted small-text">
              {tr(
                "Pro 到期后保留全部地址，超出 3 个的地址暂停预警，续费后自动恢复。",
                "When Pro expires all addresses are kept; alerts pause for those beyond 3 and resume after renewal.",
              )}
            </p>
          </section>
        </div>
      </main>
    </>
  );
}
