import { useEffect, useState, type ReactNode } from "react";
import { Link, useLocation } from "react-router-dom";
import { api, storedRef, type LiqMap } from "../api";
import { getPosts } from "../blog/posts";
import { PostMeta } from "./Blog";
import Brand from "../components/Brand";
import FailoverDiagram from "../components/FailoverDiagram";
import HealthGauge from "../components/HealthGauge";
import LangToggle from "../components/LangToggle";
import LiqMapChart, { liquidatedWithin } from "../components/LiqMapChart";
import { usdShort } from "../format";
import LiquidationFlow from "../components/LiquidationFlow";
import LoginDialog from "../components/LoginDialog";
import StatusStrip from "../components/StatusStrip";
import { tr } from "../i18n";

const navLinks = () => [
  { label: tr("清算地图", "Liq. map"), href: "#liqmap" },
  { label: tr("定价", "Pricing"), href: "#pricing" },
  { label: tr("推荐计划", "Referral"), href: "#referral" },
  { label: tr("博客", "Blog"), href: "#blog" },
];

const steps = () => [
  { title: tr("Telegram 登录", "Log in with Telegram"), body: tr("不用注册，在机器人里点一下确认。", "No sign-up. Just confirm in the bot.") },
  {
    title: tr("添加钱包地址", "Add a wallet address"),
    body: tr("粘贴 0x 地址，自动识别各协议借贷仓位。", "Paste a 0x address. Lending positions are detected automatically."),
  },
  {
    title: tr("收到预警", "Get alerted"),
    body: tr("健康因子跌破预警线，私聊立刻提醒。", "When the health factor drops below your line, you get a DM right away."),
  },
];

const trust = () => [
  {
    title: tr("只读，不碰资产", "Read-only, never touches funds"),
    body: tr("只要地址，不连钱包、不签名、不要私钥。", "Just an address. No wallet connection, no signatures, no private keys."),
  },
  {
    title: tr("链上直读", "Read straight from chain"),
    body: tr("Aave、Compound 健康因子直接读协议合约。", "Aave and Compound health factors come directly from protocol contracts."),
  },
  {
    title: tr("多节点容灾", "Multi-node failover"),
    body: tr("节点挂了自动切换，持续异常值班告警。", "Dead nodes are switched out automatically; persistent issues page the on-call."),
  },
  {
    title: tr("不会漏报", "No missed alerts"),
    body: tr("查询失败不当作安全，推送失败自动补发。", "A failed query is never treated as safe. Failed pushes are retried."),
  },
  {
    title: tr("收款私钥离线", "Payment keys stay offline"),
    body: tr("服务器只有收款公钥，碰不到任何资金。", "The server only holds public keys and cannot move any funds."),
  },
  { title: tr("最少数据", "Minimal data"), body: tr("不要邮箱和密码，不记录 IP。", "No email, no password, no IP logging.") },
];

const loopPains = () => [
  {
    pro: true,
    pain: tr("ETH 一根针，预言机还没动，你已经来不及", "ETH wicks down before the oracle moves, and by then it's too late"),
    fix: tr(
      "按币安实时价格估算健康因子，抢在链上预言机更新之前提醒你。",
      "We estimate your health factor from live Binance prices and alert you before the on-chain oracle updates.",
    ),
  },
  {
    pro: true,
    pain: tr("利率悄悄飙到 20%，循环收益变负，自己没察觉", "Borrow APY quietly hits 20% and your loop turns negative"),
    fix: tr("借款年化超过 15% 立刻提醒，回落了再通知你。", "Get alerted the moment borrow APY tops 15%, and again when it cools off."),
  },
  {
    pro: true,
    pain: tr("stETH 或稳定币脱锚，循环仓位一夜变危险", "stETH or a stablecoin depegs and your loop turns dangerous overnight"),
    fix: tr(
      "盯住你仓位里的 USDC、USDT、DAI、USDe、GHO、stETH、eETH，偏离锚定就提醒，回到锚定再通知。",
      "We watch the USDC, USDT, DAI, USDe, GHO, stETH and eETH in your position and alert you when one drifts off its peg, then again when it's back.",
    ),
  },
  {
    pro: false,
    pain: tr("半夜收到预警，还得自己算还多少", "An alert at 3 a.m., and you still have to do the math"),
    fix: tr(
      "预警里直接写明还多少、补多少，附处理链接，一点就去。",
      "The alert says exactly how much to repay or add, with a link to go fix it.",
    ),
  },
];

function LoopAlertCard() {
  return (
    <figure className="loop-card frame">
      <figcaption className="micro">{tr("示例消息", "Sample alert")}</figcaption>
      <div className="loop-card-head">
        <img src="/logo.svg" alt="" />
        <b>MechBot</b>
        <span className="loop-card-bot">BOT</span>
        <time>03:12</time>
      </div>
      <div className="loop-card-body">
        <p className="loop-card-title">⏱ {tr("提前预警", "Early alert")}</p>
        <p className="loop-card-market">Aave V3 · Ethereum EtherFi · weETH / USDC</p>
        <div className="loop-card-hf">
          <span className="muted">{tr("健康因子：链上 → 按币安实时价格预计", "Health factor: on-chain → Binance estimate")}</span>
          <strong>
            1.12 <i>→</i> <em>1.04</em>
          </strong>
          <span className="muted">
            {tr("预警线 1.10，预言机一更新就会跌破", "Alert line 1.10. It breaks the moment the oracle updates.")}
          </span>
        </div>
        <p className="loop-card-eth">
          ETH <span>$2,345 → $2,180</span> <em>−7.0%</em>
        </p>
        <div className="loop-card-plan">
          <b>{tr("怎么处理（回到 1.40）", "What to do (back to 1.40)")}</b>
          <div className="loop-card-options">
            <div>
              <span>{tr("还款", "Repay")}</span>
              <em>12,800 USDC</em>
            </div>
            <i>{tr("或", "or")}</i>
            <div>
              <span>{tr("补抵押", "Add collateral")}</span>
              <em>8.27 weETH</em>
            </div>
          </div>
        </div>
        <span className="loop-card-link">{tr("去处理 →", "Fix it →")}</span>
      </div>
    </figure>
  );
}

const liqPoints = () => [
  {
    pro: false,
    title: tr("一眼看到清算墙", "See the liquidation walls"),
    body: tr("哪个价位清算最密集，跌到那里会集中抛售。", "Where liquidations cluster, and where forced selling piles up."),
  },
  {
    pro: false,
    title: tr("跌 10%、20% 会清算多少", "How much goes at −10% and −20%"),
    body: tr("按当前仓位估算，免费可看（延迟 1 小时）。", "Estimated from current positions. Free with a 1-hour delay."),
  },
  {
    pro: true,
    title: tr("实时、更细、看到大户", "Live, finer, with the whales"),
    body: tr("约 15 分钟刷新，$25 一档，Aave / Morpho 分开，大额仓位 Top 20。", "Refreshed about every 15 min, $25 bands, Aave and Morpho split, top 20 positions."),
  },
];

const TEASER_ROWS = 12;

function LiqTeaser() {
  const [map, setMap] = useState<LiqMap | null>(null);

  useEffect(() => {
    api.liqmap().then(setMap, () => undefined);
  }, []);

  const ready = map?.ready ? map : null;
  return (
    <figure className="liq-teaser frame">
      <figcaption className="liq-teaser-head">
        <span className="micro">{tr("ETH 清算地图 · Ethereum", "ETH liquidation map · Ethereum")}</span>
        {ready && <span className="micro">{tr("延迟 1 小时", "1h delayed")}</span>}
      </figcaption>
      {ready ? (
        <>
          <dl className="liq-teaser-stats">
            <div>
              <dt>{tr("跌 10% 内", "Within −10%")}</dt>
              <dd>{usdShort(liquidatedWithin(ready, 0.1))}</dd>
            </div>
            <div>
              <dt>{tr("跌 20% 内", "Within −20%")}</dt>
              <dd>{usdShort(liquidatedWithin(ready, 0.2))}</dd>
            </div>
          </dl>
          <LiqMapChart map={ready} maxRows={TEASER_ROWS} />
        </>
      ) : (
        <p className="muted liq-teaser-empty">{tr("数据收集中…", "Collecting data…")}</p>
      )}
    </figure>
  );
}

const referralTiers = () => [
  { label: tr("好友月付", "Friend buys monthly"), days: 7 },
  { label: tr("好友季付", "Friend buys quarterly"), days: 15 },
  { label: tr("好友年付", "Friend buys yearly"), days: 30 },
];

const proPlans = () => [
  { key: "month", label: tr("月付", "Monthly"), price: 5, months: 1 },
  { key: "quarter", label: tr("季付", "Quarterly"), price: 12, months: 3 },
  { key: "year", label: tr("年付", "Yearly"), price: 36, months: 12 },
];

const planSave = (p: { price: number; months: number }) => Math.round((1 - p.price / (5 * p.months)) * 100);

function SectionHead({ index, label, title, sub }: { index: string; label: string; title: ReactNode; sub?: string }) {
  return (
    <div className="sec-head">
      <span className="sec-index sec-num">
        {index} / {label}
      </span>
      <h2>{title}</h2>
      {sub && <p className="muted">{sub}</p>}
    </div>
  );
}

export default function Landing() {
  const [open, setOpen] = useState(false);
  const [loggedIn, setLoggedIn] = useState(false);
  const [botUrl, setBotUrl] = useState("https://t.me/MechFi_bot");
  const [botName, setBotName] = useState("MechFi_bot");
  const { hash } = useLocation();

  useEffect(() => {
    if (hash) document.getElementById(hash.slice(1))?.scrollIntoView();
  }, [hash]);

  useEffect(() => {
    api.me().then(() => setLoggedIn(true), () => setLoggedIn(false));
    api.config().then((c) => {
      const ref = storedRef();
      setBotUrl(ref ? `${c.botUrl}?start=ref_${ref}` : c.botUrl);
      setBotName(c.botUrl.split("/").pop() || "MechFi_bot");
    }, () => undefined);
  }, []);

  const consoleLabel = tr("进入控制台", "Open console");
  const plans = proPlans();

  const start = (label: string) =>
    loggedIn ? (
      <Link className="btn btn-primary" to="/app">
        {label} <span aria-hidden="true">→</span>
      </Link>
    ) : (
      <button className="btn btn-primary" onClick={() => setOpen(true)}>
        {label} <span aria-hidden="true">→</span>
      </button>
    );

  return (
    <div className="page">
      <StatusStrip />
      <header className="topbar">
        <div className="nav container">
          <Brand />
          <nav className="nav-links">
            {navLinks().map((l) => (
              <a key={l.href} href={l.href}>
                {l.label}
              </a>
            ))}
            <LangToggle />
            {loggedIn ? (
              <Link className="btn btn-small btn-primary" to="/app">
                {consoleLabel}
              </Link>
            ) : (
              <button className="btn btn-small btn-ghost" onClick={() => setOpen(true)}>
                {tr("登录", "Log in")}
              </button>
            )}
          </nav>
        </div>
      </header>

      <main>
        <section className="hero container">
          <div>
            <span className="sec-index sec-num">{tr("00 / DEFI 借贷清算预警", "00 / DEFI LIQUIDATION ALERTS")}</span>
            <h1>
              {tr("清算之前，", "Before liquidation,")}
              <br />
              {tr("先提醒你。", "we wake you up.")}
            </h1>
            <p className="lead">
              {tr("不只提醒，还告诉你还多少、补多少。", "Not just a warning: exactly how much to repay or add.")}
              <br />
              <b className="lead-hl">
                {tr(
                  "用交易所实时价格提前预警，抢在链上预言机更新之前。",
                  "Early alerts from live exchange prices, before the on-chain oracle catches up.",
                )}
              </b>
            </p>
            <div className="hero-actions">
              {start(loggedIn ? consoleLabel : tr("免费追踪", "Track for free"))}
              <a className="btn btn-ghost" href={botUrl} target="_blank" rel="noreferrer">
                {tr("打开", "Open")} @{botName}
              </a>
            </div>
            <dl className="hero-specs">
              <div>
                <dt>{tr("检查间隔", "Interval")}</dt>
                <dd>5 MIN</dd>
              </div>
              <div>
                <dt>{tr("协议", "Protocols")}</dt>
                <dd>3</dd>
              </div>
              <div>
                <dt>{tr("链", "Chains")}</dt>
                <dd>3</dd>
              </div>
              <div>
                <dt>Starter</dt>
                <dd>FREE</dd>
              </div>
            </dl>
          </div>
          <HealthGauge />
        </section>

        <section className="container section" id="flow">
          <SectionHead
            index="01"
            label={tr("流程", "Flow")}
            title={tr("清算预警流程", "How alerts work")}
            sub={tr(
              "盯住你在各协议的借贷仓位，健康因子跌破预警线，第一时间私聊提醒。",
              "We watch your positions across protocols and DM you the moment the health factor crosses your line.",
            )}
          />
          <LiquidationFlow />
        </section>

        <section className="container section">
          <SectionHead index="02" label={tr("使用", "Start")} title={tr("三步开始", "Three steps to start")} />
          <div className="steps">
            {steps().map((s, i) => (
              <div className="step frame" key={i}>
                <span className="step-no">{String(i + 1).padStart(2, "0")}</span>
                <h3>{s.title}</h3>
                <p className="muted">{s.body}</p>
              </div>
            ))}
          </div>
        </section>

        <section className="container section loop" id="loop">
          <div>
            <SectionHead
              index="03"
              label={tr("循环贷", "Looping")}
              title={
                <>
                  {tr("循环贷玩家，", "Running a loop?")}
                  <br />
                  {tr("专门为你盯盘。", "We watch it for you.")}
                </>
              }
              sub={tr(
                "wstETH、weETH 循环，稳定币循环，Morpho PT 循环。健康因子常年压在 1.1，差几分钟就是清算。",
                "wstETH and weETH loops, stablecoin loops, Morpho PT loops. With the health factor parked near 1.1, a few minutes decide whether you get liquidated.",
              )}
            />
            <ol className="loop-list">
              {loopPains().map((p, i) => (
                <li key={i}>
                  <span className="loop-no">{String(i + 1).padStart(2, "0")}</span>
                  <div>
                    <b>
                      {p.pain}
                      {p.pro && <span className="loop-pro">Pro</span>}
                    </b>
                    <p className="muted">{p.fix}</p>
                  </div>
                </li>
              ))}
            </ol>
            <p className="loop-cover">
              {tr(
                "支持 Aave 主市场、Lido 市场、EtherFi 市场，Morpho Blue 全部市场，Compound V3。",
                "Aave Core, Lido and EtherFi markets, every Morpho Blue market, and Compound V3.",
              )}
            </p>
            <div className="hero-actions">{start(loggedIn ? consoleLabel : tr("免费盯住我的循环仓位", "Watch my loop for free"))}</div>
          </div>
          <LoopAlertCard />
        </section>

        <section className="container section liq-promo" id="liqmap">
          <div>
            <SectionHead
              index="04"
              label={tr("清算地图", "Liquidation map")}
              title={
                <>
                  {tr("ETH 跌到哪里，", "Where ETH falls,")}
                  <br />
                  {tr("会爆多少仓。", "who gets liquidated.")}
                </>
              }
              sub={tr(
                "汇总 Ethereum 上 Aave V3、Morpho Blue 的 ETH 抵押借贷仓位，算出每个价位会被清算的抵押品。",
                "We add up ETH-backed loans on Aave V3 and Morpho Blue on Ethereum and show how much collateral is liquidated at each price.",
              )}
            />
            <ol className="loop-list">
              {liqPoints().map((p, i) => (
                <li key={i}>
                  <span className="loop-no">{String(i + 1).padStart(2, "0")}</span>
                  <div>
                    <b>
                      {p.title}
                      {p.pro && <span className="loop-pro">Pro</span>}
                    </b>
                    <p className="muted">{p.body}</p>
                  </div>
                </li>
              ))}
            </ol>
            <div className="hero-actions">
              {loggedIn ? (
                <Link className="btn btn-primary" to="/app/map">
                  {tr("查看清算地图", "Open the map")} <span aria-hidden="true">→</span>
                </Link>
              ) : (
                <button className="btn btn-primary" onClick={() => setOpen(true)}>
                  {tr("登录查看清算地图", "Log in to see the map")} <span aria-hidden="true">→</span>
                </button>
              )}
            </div>
          </div>
          <LiqTeaser />
        </section>

        <section className="container section trust" id="trust">
          <div>
            <SectionHead
              index="05"
              label={tr("安全", "Security")}
              title={
                <>
                  {tr("交给它，", "Hand it over,")}
                  <br />
                  {tr("你可以放心。", "and rest easy.")}
                </>
              }
            />
            <ul className="trust-list">
              {trust().map((t, i) => (
                <li key={i}>
                  <b>{t.title}</b>
                  <span className="muted">{t.body}</span>
                </li>
              ))}
            </ul>
          </div>
          <FailoverDiagram />
        </section>

        <section className="container section" id="pricing">
          <SectionHead index="06" label={tr("定价", "Pricing")} title={tr("定价", "Pricing")} />
          <div className="pricing">
            <div className="price-card frame">
              <div className="plan-head">
                <span className="micro">PLAN A</span>
                <h3 className="plan-title">Starter</h3>
                <p className="plan-sub plan-free">{tr("永久免费", "Free forever")}</p>
              </div>
              <div className="plan-free-price">
                <strong>
                  0<small>USDT</small>
                </strong>
                <span className="muted">
                  {tr("不绑卡、不留邮箱，Telegram 登录即用", "No card, no email. Log in with Telegram and go.")}
                </span>
              </div>
              <ul className="plan-features">
                <li>{tr("3 个监控地址", "3 watched addresses")}</li>
                <li>{tr("Aave V3 预警", "Aave V3 alerts")}</li>
                <li>{tr("Telegram 私聊通知", "Telegram DM alerts")}</li>
                <li>{tr("预警附还款 / 补仓建议", "Alerts say how much to repay or add")}</li>
                <li>{tr("ETH 清算地图（延迟 1 小时）", "ETH liquidation map (1h delayed)")}</li>
                <li className="off">{tr("Compound V3、Morpho Blue 预警", "Compound V3 and Morpho Blue alerts")}</li>
                <li className="off">{tr("币安实时价格提前预警", "Early alerts from live Binance prices")}</li>
              </ul>
              {loggedIn ? (
                <Link className="btn btn-ghost btn-block" to="/app">
                  {consoleLabel}
                </Link>
              ) : (
                <button className="btn btn-ghost btn-block" onClick={() => setOpen(true)}>
                  {tr("免费开始", "Start free")}
                </button>
              )}
            </div>
            <div className="price-card featured frame">
              <div className="plan-head">
                <span className="micro signal">{tr("PLAN B · 推荐", "PLAN B · RECOMMENDED")}</span>
                <h3 className="plan-title">Pro</h3>
                <p className="plan-sub">{tr("循环贷和多仓位首选", "Built for loops and multiple positions")}</p>
              </div>
              <div className="plan-grid">
                {plans.map((p, i) => (
                  <div className={`plan-cell ${i === plans.length - 1 ? "best" : ""}`} key={p.key}>
                    <span className="plan-label">{p.label}</span>
                    <strong>
                      {p.price}
                      <small>USDT</small>
                    </strong>
                    <span className={`plan-save ${planSave(p) > 0 ? "" : "none"}`}>
                      {tr(`省 ${planSave(p)}%`, `Save ${planSave(p)}%`)}
                    </span>
                  </div>
                ))}
              </div>
              <ul className="plan-features">
                <li>{tr("币安实时价格提前预警，循环贷插针也来得及", "Early alerts from live Binance prices, ahead of any wick")}</li>
                <li>{tr("借款年化超 15% 提醒，循环收益变负早知道", "Borrow APY over 15% alerts, before your loop turns negative")}</li>
                <li>{tr("稳定币、stETH、eETH 脱锚预警", "Depeg alerts for stablecoins, stETH and eETH")}</li>
                <li>{tr("ETH 清算地图实时版：$25 价位、协议拆分、大额仓位", "Live ETH liquidation map: $25 bands, protocol split, top positions")}</li>
                <li>{tr("20 个监控地址", "20 watched addresses")}</li>
                <li>{tr("Aave V3、Compound V3、Morpho Blue 全协议预警", "Alerts for Aave V3, Compound V3 and Morpho Blue")}</li>
                <li>{tr("Telegram 私聊通知", "Telegram DM alerts")}</li>
                <li>
                  {tr(
                    "USDT 开通：BNB Chain / Arbitrum / Ethereum / TRON（季付起）",
                    "Pay in USDT: BNB Chain / Arbitrum / Ethereum / TRON (quarterly and up)",
                  )}
                </li>
              </ul>
              <div className="plan-cta">
                {loggedIn ? (
                  <Link className="btn btn-primary btn-block" to="/app/pro">
                    {tr("升级 Pro", "Upgrade to Pro")} <span aria-hidden="true">→</span>
                  </Link>
                ) : (
                  <button className="btn btn-primary btn-block" onClick={() => setOpen(true)}>
                    {tr("免费体验 7 天 Pro", "Try Pro free for 7 days")} <span aria-hidden="true">→</span>
                  </button>
                )}
                <p className="plan-trial">
                  {tr(
                    "新用户登录即送 7 天 Pro 体验，到期自动回到 Starter，不扣费。",
                    "New users get 7 days of Pro on sign-up. It returns to Starter afterwards, with no charge.",
                  )}
                </p>
              </div>
            </div>
          </div>
        </section>

        <section className="container section" id="referral">
          <SectionHead
            index="07"
            label={tr("推荐计划", "Referral")}
            title={tr("邀请好友，双方都得 Pro", "Invite a friend, both get Pro")}
          />
          <div className="referral-note frame">
            <div className="referral-copy">
              <b>{tr("好友首次购买，你和对方各得同样天数。", "On a friend's first purchase, you both get the same Pro days.")}</b>
              <p className="muted">
                {tr("邀请多人，奖励累加。续费不重复计算。", "Rewards stack across friends. Renewals don't count again.")}
              </p>
            </div>
            <dl className="referral-tiers">
              {referralTiers().map((t) => (
                <div key={t.days}>
                  <dt>{t.label}</dt>
                  <dd>
                    +{t.days}
                    <small>{tr("天", " days")}</small>
                  </dd>
                </div>
              ))}
            </dl>
            <div className="referral-cta">
              {loggedIn ? (
                <Link className="btn btn-primary" to="/app/referral">
                  {tr("获取邀请链接", "Get invite link")} <span aria-hidden="true">→</span>
                </Link>
              ) : (
                <button className="btn btn-primary" onClick={() => setOpen(true)}>
                  {tr("获取邀请链接", "Get invite link")} <span aria-hidden="true">→</span>
                </button>
              )}
            </div>
          </div>
        </section>

        <section className="container section" id="blog">
          <div className="blog-section-head">
            <SectionHead index="08" label={tr("博客", "Blog")} title={tr("技术文章", "Engineering notes")} />
            <Link className="text-link" to="/blog">
              {tr("更多文章 →", "More posts →")}
            </Link>
          </div>
          <div className="post-list">
            {getPosts()
              .slice(0, 3)
              .map((p) => (
                <Link className="post-card frame" to={`/blog/${p.slug}`} key={p.slug}>
                  <PostMeta date={p.date} tag={p.tag} readMin={p.readMin} />
                  <h3>{p.title}</h3>
                  <p className="muted">{p.summary}</p>
                  <span className="post-more">
                    {tr("阅读全文", "Read more")} <span aria-hidden="true">→</span>
                  </span>
                </Link>
              ))}
          </div>
        </section>
      </main>

      <footer className="footer">
        <div className="container footer-inner">
          <div className="footer-brand">
            <Brand />
            <p className="muted">Stay calm, stay ahead.</p>
          </div>
          <p className="footer-note">
            {tr(
              "预警仅供参考，可能因网络或节点原因延迟、失败，不构成投资建议。请自行管理仓位风险。",
              "Alerts are for reference only and may be delayed or fail due to network or node issues. Not investment advice. Manage your own position risk.",
            )}
          </p>
        </div>
        <div className="footer-bar">
          <div className="container footer-bar-inner">
            <span>© {new Date().getFullYear()} MechBot. {tr("保留所有权利。", "All rights reserved.")}</span>
            <nav className="footer-links">
              <a href="#flow">{tr("清算预警", "Alerts")}</a>
              <a href="#loop">{tr("循环贷", "Looping")}</a>
              <a href="#liqmap">{tr("清算地图", "Liq. map")}</a>
              <a href="#trust">{tr("安全", "Security")}</a>
              <a href="#pricing">{tr("定价", "Pricing")}</a>
              <a href="#referral">{tr("推荐计划", "Referral")}</a>
              <a href="#blog">{tr("博客", "Blog")}</a>
            </nav>
            <nav className="footer-links">
              <a href="#">{tr("服务条款", "Terms")}</a>
              <a href="#">{tr("隐私政策", "Privacy")}</a>
              <a className="footer-tg" href={botUrl} target="_blank" rel="noreferrer">
                <svg
                  width="14"
                  height="14"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinejoin="round"
                  aria-hidden="true"
                >
                  <path d="M22 2 11 13" />
                  <path d="M22 2 15 22l-4-9-9-4 20-7z" />
                </svg>
                MechBot Telegram
              </a>
            </nav>
          </div>
        </div>
      </footer>

      <a
        className="support-fab"
        href="https://t.me/+LWTqp4To2ilhY2M1"
        target="_blank"
        rel="noreferrer"
        aria-label={tr("客服 Telegram 群 MechBot", "Support Telegram group MechBot")}
      >
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinejoin="round" aria-hidden="true">
          <path d="M22 2 11 13" />
          <path d="M22 2 15 22l-4-9-9-4 20-7z" />
        </svg>
        <span className="support-fab-text">
          <b>{tr("客服群", "Support")}</b>
          <small>MechBot</small>
        </span>
      </a>

      {open && <LoginDialog onClose={() => setOpen(false)} />}
    </div>
  );
}
