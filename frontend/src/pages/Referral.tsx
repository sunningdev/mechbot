import { useEffect, useState } from "react";
import { api, type Referral as ReferralStats } from "../api";
import CopyButton from "../components/CopyButton";
import PageHead, { Ring } from "../components/PageHead";
import { planName, tr } from "../i18n";

const rewards = [
  { plan: "month", days: 7 },
  { plan: "quarter", days: 15 },
  { plan: "year", days: 30 },
];

export default function Referral() {
  const [ref, setRef] = useState<ReferralStats | null>(null);

  useEffect(() => {
    api.referral().then(setRef, () => undefined);
  }, []);

  const days = (n: number) => tr(`+${n} 天`, `+${n} days`);

  return (
    <>
      <PageHead
        title={tr("推荐", "Referral")}
        sub={tr(
          "把链接发给朋友，对方首次购买后，你们各得 Pro 天数。",
          "Share your link. When a friend makes their first purchase, you both get Pro days.",
        )}
        stats={[
          {
            label: tr("总推荐数", "Total referrals"),
            value: ref?.invited ?? "—",
            hint: tr("通过你的链接注册", "Signed up via your link"),
          },
          {
            label: tr("已付费推荐", "Paid referrals"),
            value: ref ? <span className="hf ok">{ref.paid}</span> : "—",
            hint: ref && ref.invited > 0 ? tr("转化率", "Conversion") : tr("首次购买后计入", "Counted after first purchase"),
            aside: ref && ref.invited > 0 ? <Ring ratio={ref.paid / ref.invited} /> : undefined,
          },
          {
            label: tr("奖励天数", "Days earned"),
            value: ref ? <span className="hf ok">{ref.daysEarned}</span> : "—",
            hint: tr("已加到 Pro 时长", "Added to your Pro time"),
          },
        ]}
      />

      <main className="container dashboard">
        <div className="col-main">
          <section className="panel">
            <div className="panel-head">
              <h2>{tr("我的推荐链接", "My referral links")}</h2>
            </div>
            {ref ? (
              <div className="link-list">
                <div className="link-item">
                  <span className="muted small-text">
                    {tr("网页链接 · 好友打开网页，用 Telegram 登录即绑定", "Web link · your friend opens the site and logs in with Telegram")}
                  </span>
                  <div className="link-box">
                    <span className="mono break">{ref.webLink}</span>
                    <CopyButton text={ref.webLink} />
                  </div>
                </div>
                <div className="link-item">
                  <span className="muted small-text">
                    {tr("机器人链接 · 好友在 Telegram 里点开即绑定", "Bot link · your friend opens it in Telegram")}
                  </span>
                  <div className="link-box">
                    <span className="mono break">{ref.botLink}</span>
                    <CopyButton text={ref.botLink} />
                  </div>
                </div>
              </div>
            ) : (
              <p className="muted">{tr("加载中…", "Loading…")}</p>
            )}
          </section>
        </div>
        <div className="col-side">
          <section className="panel">
            <div className="panel-head">
              <h2>{tr("奖励规则", "Reward rules")}</h2>
            </div>
            <table className="compare">
              <thead>
                <tr>
                  <th>{tr("好友首次购买", "Friend's first buy")}</th>
                  <th>{tr("你得", "You get")}</th>
                  <th>{tr("好友得", "Friend gets")}</th>
                </tr>
              </thead>
              <tbody>
                {rewards.map((r) => (
                  <tr key={r.plan}>
                    <td>{planName(r.plan)}</td>
                    <td className="reward">{days(r.days)}</td>
                    <td className="reward">{days(r.days)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            <p className="small-text referral-rule">
              {tr(
                "每推荐一位新用户，对方首次购买后，你和对方各得 Pro 天数：月付各 +7 天，季付各 +15 天，年付各 +30 天。推荐多人付费，奖励累加。",
                "For each new user you refer, after their first purchase you both get Pro days: +7 for monthly, +15 for quarterly, +30 for yearly. Rewards stack across referrals.",
              )}
            </p>
          </section>
        </div>
      </main>
    </>
  );
}
