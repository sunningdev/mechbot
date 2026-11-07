import { useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, type WatchedAddress } from "../api";
import AddressList from "../components/AddressList";
import PageHead, { HfGauge } from "../components/PageHead";
import { proDaysLeft, riskClass } from "../format";
import { tr } from "../i18n";
import { useConsole } from "./ConsoleLayout";

// The phrase must appear verbatim inside the sentence's translation in every language.
function highlight(text: string, phrase: string) {
  const i = text.indexOf(phrase);
  if (i < 0) return text;
  return (
    <>
      {text.slice(0, i)}
      <mark className="hl">{phrase}</mark>
      {text.slice(i + phrase.length)}
    </>
  );
}

export default function Dashboard() {
  const { me, refresh } = useConsole();
  const [addresses, setAddresses] = useState<WatchedAddress[]>([]);

  const loadAddresses = useCallback(() => api.addresses().then(setAddresses, () => undefined), []);

  useEffect(() => {
    void loadAddresses();
    const timer = window.setInterval(() => void loadAddresses(), 30_000);
    return () => window.clearInterval(timer);
  }, [loadAddresses]);

  const reload = () => {
    void loadAddresses();
    void refresh();
  };

  const name = me.username ? `@${me.username}` : me.firstName || String(me.telegramId);
  const lowest = addresses
    .flatMap((a) => a.positions.map((p) => ({ hf: p.healthFactor, threshold: a.threshold, p })))
    .filter(({ p }) => !p.error && p.covered && p.debtUsd > 0 && Number.isFinite(p.healthFactor))
    .reduce<{ hf: number; threshold: number } | null>((low, x) => (!low || x.hf < low.hf ? x : low), null);
  const minHf = lowest?.hf ?? null;
  const minThreshold = lowest?.threshold ?? 0;
  const proDays = proDaysLeft(me.pro, me.proExpireAt);
  const left = Math.max(0, me.addressLimit - me.addressCount);
  const risk = minHf === null ? "" : riskClass(minHf, minThreshold);

  return (
    <>
      <PageHead
        title={<>{tr("你好，", "Hi, ")}{name}</>}
        sub={
          <>
            {highlight(
              tr(
                "每 5 分钟检查一次健康因子，低于该地址的预警线时通过 Telegram 私聊提醒你。",
                "We check health factors every 5 minutes and DM you on Telegram when one drops below that address's alert line.",
              ),
              tr("通过 Telegram 私聊提醒你", "DM you on Telegram"),
            )}
            <span className="sub-note">
              {tr(
                "提醒频率：低于预警线每 1 小时提醒一次；健康因子低于 1.05（非常接近清算）每 15 分钟提醒一次；回到预警线以上会发一条恢复通知。",
                "Reminder frequency: every hour while below the alert line, every 15 minutes once the health factor is under 1.05 (very close to liquidation), plus one all-clear message when it recovers.",
              )}
            </span>
          </>
        }
        stats={[
          {
            label: tr("监控地址", "Watched addresses"),
            value: (
              <>
                {me.addressCount}
                <small> / {me.addressLimit}</small>
              </>
            ),
            hint:
              left > 0
                ? tr(`还可添加 ${left} 个`, `${left} more available`)
                : me.pro
                  ? tr("已达上限", "Limit reached")
                  : tr("Pro 可监控 20 个", "Pro watches 20"),
            tag: left === 0 ? { text: tr("已满", "Full"), tone: "danger" } : undefined,
          },
          {
            label: tr("最低健康因子", "Lowest health factor"),
            value: minHf === null ? "—" : minHf.toFixed(3),
            className: minHf === null ? "" : `hf ${risk}`,
            hint: minHf === null ? tr("暂无借贷仓位", "No borrow positions") : tr(`预警线 ${minThreshold}`, `Alert line ${minThreshold}`),
            tag:
              minHf === null
                ? undefined
                : risk === "ok"
                  ? { text: tr("安全", "Safe"), tone: "ok" }
                  : { text: tr("需关注", "Watch"), tone: "danger" },
            aside: minHf === null ? undefined : <HfGauge value={minHf} threshold={minThreshold} />,
          },
          {
            label: tr("当前方案", "Current plan"),
            value: me.pro ? (me.trial ? tr("Pro 体验", "Pro trial") : "Pro") : "Starter",
            className: me.pro ? "hf ok" : "",
            hint:
              proDays !== null
                ? tr(`剩 ${proDays} 天`, `${proDays} days left`)
                : me.proExpiredAt
                  ? me.trial
                    ? tr("体验已结束", "Trial ended")
                    : tr("Pro 已到期", "Pro expired")
                  : tr("仅 Aave V3", "Aave V3 only"),
            tag:
              proDays === null
                ? undefined
                : me.trial
                  ? { text: tr("体验中", "Trial"), tone: "ok" }
                  : proDays <= 7
                    ? { text: tr("即将到期", "Expiring"), tone: "danger" }
                    : { text: tr("生效中", "Active"), tone: "ok" },
            aside: (
              <Link to="/app/pro" className={`btn btn-small ${me.pro && !me.trial ? "btn-ghost" : "btn-primary"}`}>
                {me.pro && !me.trial ? tr("续费", "Renew") : tr("升级 Pro", "Upgrade")}
              </Link>
            ),
          },
        ]}
      />

      <main className="container dashboard dashboard-single">
        <AddressList me={me} addresses={addresses} onChange={reload} />
      </main>
    </>
  );
}
