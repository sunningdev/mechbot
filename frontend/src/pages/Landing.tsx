import { useEffect, useState, type ReactNode } from "react";
import { Link, useLocation } from "react-router-dom";
import { api, storedRef, type LiqMap } from "../api";
import { getPosts } from "../blog/posts";
import { PostMeta } from "./Blog";
import Brand from "../components/Brand";
import LangToggle from "../components/LangToggle";
import ThemeToggle from "../components/ThemeToggle";
import LiqMapChart, { liquidatedWithin } from "../components/LiqMapChart";
import { usdShort } from "../format";
import LoginDialog from "../components/LoginDialog";
import { tr } from "../i18n";
import { safeUrl, TELEGRAM_HOSTS } from "../links";
import { coinOrder, sortByChain } from "../payorder";

const navLinks = () => [
  { label: tr("功能", "Features"), href: "#features" },
  { label: tr("清算地图", "Liq. map"), href: "#liqmap" },
  { label: tr("定价", "Pricing"), href: "#pricing" },
  { label: "FAQ", href: "#faq" },
  { label: tr("博客", "Blog"), href: "#blog" },
];

const networks = [
  { name: "Aave", logo: "/logos/aave.svg" },
  { name: "Compound", logo: "/logos/compound.svg" },
  { name: "Morpho", logo: "/logos/morpho.png" },
  { name: "Ethereum", logo: "/logos/ethereum.svg" },
  { name: "Arbitrum", logo: "/logos/arbitrum.svg" },
  { name: "Base", logo: "/logos/base.svg" },
  { name: "Optimism", logo: "/logos/optimism.svg" },
];

const faqs = () => [
  {
    q: tr("如何在 Aave 清算前收到 Telegram 提醒？", "How do I get Telegram alerts before an Aave liquidation?"),
    a: tr(
      "用 Telegram 登录后添加公开钱包地址。MechBot 约每 5 分钟检查 Aave V3 健康因子，跌破预警线时私聊提醒，并告诉你该还多少或补多少抵押。",
      "Log in with Telegram and add your public wallet address. MechBot checks Aave V3 health factor about every 5 minutes and DMs you when it crosses your line, with how much to repay or add as collateral.",
    ),
  },
  {
    q: tr("有没有免费的 Aave、Compound、Morpho 清算预警机器人？", "Is there a free DeFi liquidation alert bot for Aave, Compound and Morpho?"),
    a: tr(
      "有。免费版监控 Aave V3；Pro 解锁 Compound V3、Morpho Blue、更多地址、币安价格提前预警和清算地图。无需连接钱包。",
      "Yes. The free plan monitors Aave V3. Pro unlocks Compound V3, Morpho Blue, more addresses, earlier Binance-price warnings and the liquidation map. No wallet connect required.",
    ),
  },
  {
    q: tr("需要连接钱包吗？", "Do I need to connect my wallet?"),
    a: tr(
      "不需要。MechBot 只要一个公开地址，不需要签名、授权或私钥。",
      "No. MechBot only needs a public address. It never asks for signatures, approvals or private keys.",
    ),
  },
  {
    q: tr("多久检查一次仓位？", "How often are positions checked?"),
    a: tr(
      "链上每 5 分钟检查一次。Pro 还会用币安实时价格，在预言机更新前提前预警。",
      "Every 5 minutes on-chain. Pro also uses live Binance prices to warn you before the oracle updates.",
    ),
  },
  {
    q: tr("支持哪些协议和公链？", "Which protocols and chains are supported?"),
    a: tr(
      "支持 Ethereum、Arbitrum、Base、Optimism 上的 Aave V3、Compound V3 和 Morpho Blue。免费版支持 Aave V3。",
      "Aave V3, Compound V3 and Morpho Blue on Ethereum, Arbitrum, Base and Optimism. The free plan covers Aave V3.",
    ),
  },
  {
    q: tr("MechBot 能监控 Morpho Blue 和 Compound V3 健康因子吗？", "Can MechBot monitor Morpho Blue and Compound V3 health factor?"),
    a: tr(
      "可以。免费版覆盖 Aave V3；Pro 同时监控 Ethereum、Arbitrum、Base、Optimism 上的 Compound V3 与 Morpho Blue。",
      "Yes. Free covers Aave V3; Pro also watches Compound V3 and Morpho Blue on Ethereum, Arbitrum, Base and Optimism.",
    ),
  },
  {
    q: tr("Pro 怎么付款？", "How do I pay for Pro?"),
    a: tr(
      "用 USDT 或 USDC。转到你的专属充值地址，或连接钱包直接转账。TRON 从季付起支持。",
      "With USDT or USDC. Send to your personal deposit address, or connect a wallet and pay with a plain transfer. TRON is available from the quarterly plan.",
    ),
  },
  {
    q: tr("可以退款吗？", "Can I get a refund?"),
    a: tr(
      "已付款项不退，除非服务停止运营，届时按剩余时长比例退款。",
      "Payments are non-refundable unless the service shuts down. In that case, unused time is refunded pro rata.",
    ),
  },
  {
    q: tr("MechBot 开源吗？", "Is MechBot open source?"),
    a: tr(
      "是的。代码在 GitHub 开源，服务器只有收款公钥，碰不到任何资金。",
      "Yes. The code is on GitHub, and the server only holds payment public keys, so it can't move any funds.",
    ),
  },
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

const heroPoints = () => [
  {
    title: tr("比预言机更早", "Ahead of the oracle"),
    body: tr(
      "用币安实时价估算健康因子，插针时抢在预言机更新前提醒你。",
      "Live Binance prices catch a wick before the on-chain oracle does.",
    ),
    note: tr(
      "基于币安实时价估算，仅供参考，不构成链上清算保证。",
      "An estimate for reference only, not a guarantee about on-chain liquidation.",
    ),
  },
  {
    title: tr("告诉你补多少", "Exactly how much to fix"),
    body: tr(
      "直接写明「还 12,800 USDC」或「补 8.27 weETH」，附协议直达链接，一点就能处理。",
      "“Repay 12,800 USDC” or “add 8.27 weETH”, with a direct link to fix it.",
    ),
  },
  {
    title: tr("一个地址，多链监听", "One address, every chain"),
    body: tr(
      "粘贴一个地址，自动找出 Ethereum、Arbitrum、Base、Optimism 上的 Aave、Compound、Morpho 仓位。",
      "Paste one address to find its Aave, Compound and Morpho loans on Ethereum, Arbitrum, Base and Optimism.",
    ),
  },
  {
    title: tr("自定义预警线，灵活提醒", "Your own alert line, smart reminders"),
    body: tr(
      "每个地址单独设线，改完立刻检查。低于线每小时提醒，低于 1.05 每 15 分钟提醒。",
      "Set a line per address, checked instantly. Reminders hourly below it, every 15 min under 1.05.",
    ),
  },
  {
    title: tr("循环贷专属预警", "Built for loopers"),
    body: tr(
      "借款利率飙升、stETH 或稳定币脱锚，在循环收益变负、仓位变危险前提醒你。",
      "Borrow rate spikes and stETH or stablecoin depegs, flagged before your loop turns negative or risky.",
    ),
  },
  {
    title: tr("清算地图", "Liquidation map"),
    body: tr(
      "ETH 跌到哪、爆多少、清算墙在哪，一张图看清。",
      "Where ETH breaks, how much gets liquidated, where the walls are, in one chart.",
    ),
  },
];

const heroChecks = () => [
  tr("永久免费", "Free forever"),
  tr("不连钱包、不签名", "No wallet connect, no signing"),
  tr("每 5 分钟检查", "Checked every 5 min"),
];

const icon = (d: string) => (
  <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true">
    <path d={d} />
  </svg>
);

function OracleViz() {
  return (
    <div className="bx-viz bx-oracle" aria-hidden="true">
      <svg viewBox="0 0 300 100" preserveAspectRatio="none">
        <line className="bx-oracle-line" x1="0" y1="62" x2="300" y2="62" />
        <path className="bx-oracle-feed" d="M0 30 H205 V68 H300" />
        <path className="bx-oracle-price" d="M0 30 L30 27 L60 33 L90 29 L120 34 L140 31 L155 42 L168 88 L178 72 L195 66 L225 64 L260 67 L300 66" />
      </svg>
      <span className="bx-oracle-dot" />
      <span className="bx-oracle-tip">{tr("已提醒", "Alert sent")}</span>
      <div className="bx-legend">
        <span className="price">{tr("币安实时价", "Binance live")}</span>
        <span className="feed">{tr("预言机", "Oracle")}</span>
      </div>
    </div>
  );
}

function FixViz() {
  return (
    <div className="bx-viz bx-fix" aria-hidden="true">
      <div className="bx-msg">
        <img src="/logo.svg" alt="" />
        <div>
          <b>{tr("还 12,800 USDC", "Repay 12,800 USDC")}</b>
          <span>{tr("或补 8.27 weETH", "or add 8.27 weETH")}</span>
        </div>
      </div>
      <span className="bx-fix-btn">{tr("去 Aave 处理", "Fix on Aave")} →</span>
    </div>
  );
}

function ChainsViz() {
  return (
    <div className="bx-viz bx-chains" aria-hidden="true">
      <span className="bx-addr">0x7a3f…c21e</span>
      <span className="bx-wire" />
      <div className="bx-logos">
        {networks.map((n) => (
          <img key={n.name} src={n.logo} alt="" title={n.name} />
        ))}
      </div>
    </div>
  );
}

function AlertLineViz() {
  return (
    <div className="bx-viz bx-line" aria-hidden="true">
      <div className="bx-line-bar">
        <span className="bx-line-crit" />
        <span className="bx-line-mark" style={{ left: "20%" }}>
          <em>{tr("预警线", "Alert line")} 1.20</em>
        </span>
        <span className="bx-line-now" style={{ left: "16%" }} />
      </div>
      <div className="bx-line-scale">
        <span>1.0 {tr("清算", "Liquidation")}</span>
        <span>2.0</span>
      </div>
      <div className="bx-chips">
        <span className="warn">{tr("低于预警线 · 每小时", "Below line · hourly")}</span>
        <span className="danger">{tr("低于 1.05 · 每 15 分钟", "Under 1.05 · every 15 min")}</span>
      </div>
    </div>
  );
}

function LoopViz() {
  return (
    <div className="bx-viz bx-loop" aria-hidden="true">
      <div className="bx-row warn">
        <span>{tr("借款年化", "Borrow APY")}</span>
        <b>15.2% ↑</b>
      </div>
      <div className="bx-row danger">
        <span>stETH</span>
        <b>0.982 · {tr("脱锚", "Depeg")}</b>
      </div>
    </div>
  );
}

const liqBars = [8, 12, 10, 16, 14, 22, 18, 30, 26, 88, 40, 34, 24, 28, 18, 14, 10, 12];

function LiqViz() {
  return (
    <div className="bx-viz bx-liq" aria-hidden="true">
      <div className="bx-bars">
        {liqBars.map((h, i) =>
          h > 60 ? (
            <i key={i} className="wall" style={{ height: `${h}%` }}>
              <span>{tr("清算墙", "Wall")}</span>
            </i>
          ) : (
            <i key={i} style={{ height: `${h}%` }} />
          ),
        )}
      </div>
    </div>
  );
}

// Component types (not pre-built JSX) so tr() inside each viz follows the current language.
const featureViz = [OracleViz, FixViz, ChainsViz, AlertLineViz, LoopViz, LiqViz];
const wideFeatures = new Set([0, 3, 5]);

const trustIcons = [
  icon("M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12zM12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z"),
  icon("M12 3 4 6v6c0 5 3.5 8 8 9 4.5-1 8-4 8-9V6zM9 12l2 2 4-4"),
  icon("M5 5h14v5H5zM5 14h14v5H5zM8 7.5h.01M8 16.5h.01"),
  icon("M21 12a9 9 0 1 1-3-6.7L21 8M21 3v5h-5"),
  icon("M7 11V7a5 5 0 0 1 10 0v4M5 11h14v10H5z"),
  icon("M12 3a9 9 0 1 0 0 18 9 9 0 0 0 0-18zM5.6 5.6l12.8 12.8"),
];

function TelegramIcon() {
  return (
    <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor" aria-hidden="true">
      <path d="M21.9 4.3 18.7 19.4c-.2 1.1-.9 1.3-1.8.8l-4.9-3.6-2.4 2.3c-.3.3-.5.5-1 .5l.3-5 9.1-8.2c.4-.4-.1-.6-.6-.2L6.2 13.1l-4.8-1.5c-1-.3-1.1-1 .2-1.5L20.5 2.8c.9-.3 1.7.2 1.4 1.5z" />
    </svg>
  );
}

const loopPoints = () => [
  {
    pro: true,
    title: tr("借款利率预警", "Borrow rate alerts"),
    body: tr("借款年化超过 15% 立刻提醒，回落了再通知你。", "Get alerted the moment borrow APY tops 15%, and again when it cools off."),
  },
  {
    pro: true,
    title: tr("脱锚预警", "Depeg alerts"),
    body: tr(
      "盯住你仓位里的 USDC、USDT、DAI、USDe、GHO、stETH、eETH，偏离锚定就提醒，回到锚定再通知。",
      "We watch the USDC, USDT, DAI, USDe, GHO, stETH and eETH in your position and alert you when one drifts off its peg, then again when it's back.",
    ),
  },
  {
    pro: false,
    title: tr("覆盖主流市场", "Every major market"),
    body: tr(
      "支持 Aave 主市场、Lido 市场、EtherFi 市场，Morpho Blue 全部市场，Compound V3；覆盖 Ethereum、Arbitrum、Base、Optimism。",
      "Aave Core, Lido and EtherFi markets, every Morpho Blue market, and Compound V3, across Ethereum, Arbitrum, Base and Optimism.",
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
  { key: "month", label: tr("月付", "Monthly"), price: 8, months: 1 },
  { key: "quarter", label: tr("季付", "Quarterly"), price: 20, months: 3 },
  { key: "year", label: tr("年付", "Yearly"), price: 60, months: 12 },
];

function payLine() {
  const networks = sortByChain(
    [
      ["bsc", "BNB Chain"],
      ["arbitrum", "Arbitrum"],
      ["base", "Base"],
      ["optimism", "Optimism"],
      ["ethereum", "Ethereum"],
      ["tron", tr("TRON（季付起）", "TRON (from quarterly)")],
    ],
    ([key]) => key,
  ).map(([, name]) => name);
  return tr("{0} 开通：{1}", "Pay in {0}: {1}")
    .replace("{0}", coinOrder().join(" / "))
    .replace("{1}", networks.join(" / "));
}

const planSave = (p: { price: number; months: number }) => Math.round((1 - p.price / (proPlans()[0].price * p.months)) * 100);

function SectionHead({ label, title, sub, center }: { label: string; title: ReactNode; sub?: string; center?: boolean }) {
  return (
    <div className={center ? "sec-head center" : "sec-head"}>
      <span className="sec-index sec-num">{label}</span>
      <h2>{title}</h2>
      {sub && <p className="muted">{sub}</p>}
    </div>
  );
}

function GitHubIcon({ size }: { size: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
      <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z" />
    </svg>
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
      const bot = safeUrl(c.botUrl, TELEGRAM_HOSTS);
      if (!bot) return;
      const ref = storedRef();
      setBotUrl(ref ? `${bot}?start=ref_${ref}` : bot);
      setBotName(bot.split("/").pop() || "MechFi_bot");
    }, () => undefined);
  }, []);

  useEffect(() => {
    const key = "mb_home_" + new Date().toISOString().slice(0, 10);
    if (sessionStorage.getItem(key)) return;
    sessionStorage.setItem(key, "1");
    api.trackHome().catch(() => undefined);
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
      <header className="topbar">
        <div className="nav container">
          <Brand />
          <nav className="nav-links">
            {navLinks().map((l) => (
              <a key={l.href} href={l.href}>
                {l.label}
              </a>
            ))}
            <a
              className="nav-gh"
              href="https://github.com/sunningdev/mechbot"
              target="_blank"
              rel="noreferrer"
              aria-label="GitHub"
              title="GitHub"
            >
              <GitHubIcon size={18} />
            </a>
            <ThemeToggle />
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
          <div className="hero-copy">
            <span className="sec-index">{tr("DeFi 借贷清算预警", "Aave · Compound · Morpho liquidation alerts")}</span>
            <h1>{tr("清算之前，\n先提醒你。", "Get alerted before you get liquidated.")}</h1>
            <p className="hero-sub">
              {tr(
                "MechBot 是免费的 DeFi 清算预警机器人：7×24 监控 Aave V3、Compound V3、Morpho Blue 健康因子。快被清算时 Telegram 立刻提醒，并告诉你该还多少、补多少。无需连接钱包。",
                "Free DeFi liquidation alert bot: MechBot monitors Aave V3, Compound V3 and Morpho Blue health factors 24/7. Get Telegram alerts before liquidation—with exactly how much to repay or add. No wallet connect.",
              )}
            </p>
            <div className="hero-actions">
              {start(loggedIn ? consoleLabel : tr("免费开始", "Start free"))}
              <a className="btn btn-ghost" href={botUrl} target="_blank" rel="noreferrer" title={`@${botName}`}>
                <TelegramIcon /> {tr("在 Telegram 打开", "Open in Telegram")}
              </a>
            </div>
            <ul className="hero-checks">
              {heroChecks().map((t) => (
                <li key={t}>{t}</li>
              ))}
            </ul>
          </div>
        </section>

        <section className="container logo-strip" aria-label={tr("支持的协议和链", "Supported protocols and chains")}>
          <p>{tr("支持的协议和公链", "Monitors positions on")}</p>
          <ul>
            {networks.map((n) => (
              <li key={n.name}>
                <img src={n.logo} alt="" width="28" height="28" />
                {n.name}
              </li>
            ))}
          </ul>
        </section>

        <section className="container section" id="features">
          <SectionHead
            label={tr("功能", "Features")}
            title={tr("暴跌之夜，也能安心睡觉。", "Everything you need to sleep through a crash.")}
          />
          <div className="bento">
            {heroPoints().map((p, i) => {
              const Viz = featureViz[i];
              return (
                <div className={`bento-card frame ${wideFeatures.has(i) ? "wide" : ""}`} key={i}>
                  <Viz />
                  <div className="bento-copy">
                    <h3>{p.title}</h3>
                    <p className="muted">{p.body}</p>
                    {"note" in p && <small className="bento-note">{p.note}</small>}
                  </div>
                </div>
              );
            })}
          </div>
        </section>

        <section className="container section" id="how">
          <SectionHead
            center
            label={tr("使用方法", "How it works")}
            title={tr("两分钟就能设置好。", "Set up in two minutes.")}
            sub={tr("不用注册，不连钱包。只要 Telegram 和一个地址。", "No sign-up, no wallet connection. Just Telegram and an address.")}
          />
          <ol className="how-steps">
            {steps().map((s, i) => (
              <li key={i}>
                <span className="how-no">{i + 1}</span>
                <h3>{s.title}</h3>
                <p className="muted">{s.body}</p>
              </li>
            ))}
          </ol>
        </section>

        <section className="container section split" id="loop">
          <div>
            <SectionHead
              label={tr("循环贷", "Looping")}
              title={tr("为循环贷而生。", "Built for leveraged loops.")}
              sub={tr(
                "wstETH、weETH 循环，稳定币循环，Morpho PT 循环。健康因子常年压在 1.1，差几分钟就是清算。",
                "wstETH and weETH loops, stablecoin loops, Morpho PT loops. With the health factor parked near 1.1, a few minutes decide whether you get liquidated.",
              )}
            />
            <ul className="check-list">
              {loopPoints().map((p, i) => (
                <li key={i}>
                  <b>
                    {p.title}
                    {p.pro && <span className="loop-pro">Pro</span>}
                  </b>
                  <p className="muted">{p.body}</p>
                </li>
              ))}
            </ul>
            <div className="hero-actions">{start(loggedIn ? consoleLabel : tr("免费盯住我的循环仓位", "Watch my loop for free"))}</div>
          </div>
          <LoopAlertCard />
        </section>

        <section className="container section split reverse" id="liqmap">
          <div>
            <SectionHead
              label={tr("清算地图", "Liquidation map")}
              title={tr("看清 ETH 会在哪里爆仓。", "See where ETH gets liquidated.")}
              sub={tr(
                "汇总 Ethereum 上 Aave V3、Morpho Blue 的 ETH 抵押借贷仓位，算出每个价位会被清算的抵押品。",
                "We add up ETH-backed loans on Aave V3 and Morpho Blue on Ethereum and show how much collateral is liquidated at each price.",
              )}
            />
            <ul className="check-list">
              {liqPoints().map((p, i) => (
                <li key={i}>
                  <b>
                    {p.title}
                    {p.pro && <span className="loop-pro">Pro</span>}
                  </b>
                  <p className="muted">{p.body}</p>
                </li>
              ))}
            </ul>
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

        <section className="container section" id="trust">
          <SectionHead
            center
            label={tr("安全", "Security")}
            title={tr("只读设计，交给它就放心。", "Read-only by design.")}
            sub={tr("不要私钥，不要签名，不连钱包。", "We never ask for keys, signatures or wallet connections.")}
          />
          <div className="trust-grid frame">
            {trust().map((t, i) => (
              <div className="trust-item" key={i}>
                <span className="trust-icon">{trustIcons[i]}</span>
                <h3>{t.title}</h3>
                <p className="muted">{t.body}</p>
              </div>
            ))}
          </div>
        </section>

        <section className="container section" id="pricing">
          <SectionHead
            center
            label={tr("定价", "Pricing")}
            title={tr("简单透明的价格。", "Simple pricing.")}
            sub={tr("免费开始，需要更多时再升级。", "Start free. Upgrade when you need more.")}
          />
          <div className="pricing">
            <div className="price-card frame">
              <div className="plan-head">
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
                <h3 className="plan-title">
                  Pro <span className="plan-pill">{tr("推荐", "Recommended")}</span>
                </h3>
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
                <li>{payLine()}</li>
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

        <section className="container section section-tight" id="referral">
          <div className="referral-note frame">
            <div className="referral-copy">
              <h3>{tr("邀请好友，双方都得 Pro", "Invite a friend, both get Pro")}</h3>
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

        <section className="container section" id="faq">
          <SectionHead center label="FAQ" title={tr("常见问题", "Frequently asked questions")} />
          <div className="faq">
            {faqs().map((f) => (
              <details key={f.q}>
                <summary>{f.q}</summary>
                <p className="muted">{f.a}</p>
              </details>
            ))}
          </div>
        </section>

        <section className="container section" id="blog">
          <div className="blog-section-head">
            <SectionHead label={tr("博客", "Blog")} title={tr("最新文章", "From the blog")} />
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

        <section className="container section" id="about">
          <SectionHead
            label={tr("开发者", "Developer")}
            title={tr("一个人做的，自己也在用。", "Built by one person who uses it too.")}
          />
          <div className="about-card frame">
            <div className="about-copy">
              <p>
                {tr(
                  "我是 @sunningdev，MechBot 的独立开发者。自己在 Aave 上开过仓位，有一次差点被清算。当时找不到一个能同时盯住 Aave、Compound、Morpho 的预警工具，就自己写了 MechBot。",
                  "I'm @sunningdev, the independent developer behind MechBot. I've run positions on Aave myself and once came close to being liquidated. I couldn't find an alert tool that watched Aave, Compound and Morpho together, so I built one.",
                )}
              </p>
              <p className="muted">
                {tr(
                  "代码在 GitHub 开源，服务器只有收款公钥，你可以自己核对。产品建议、问题反馈，直接到 Telegram 客服群找我。",
                  "The code is open source on GitHub and the server only holds payment public keys, so you can check for yourself. For feedback or problems, find me in the Telegram support group.",
                )}
              </p>
            </div>
            <div className="about-links">
              <a className="btn btn-ghost" href="https://github.com/sunningdev/mechbot" target="_blank" rel="noreferrer">
                <GitHubIcon size={16} />
                GitHub
              </a>
              <a className="btn btn-primary" href="https://t.me/+LWTqp4To2ilhY2M1" target="_blank" rel="noreferrer">
                {tr("Telegram 客服群", "Telegram support group")}
              </a>
            </div>
          </div>
        </section>

        <section className="container cta-band">
          <h2>{tr("不用再半夜盯盘。", "Stop watching charts at 3 a.m.")}</h2>
          <p>{tr("两分钟添加地址，剩下的交给 MechBot。", "Add an address in two minutes. MechBot keeps watch from there.")}</p>
          <div className="hero-actions">
            {start(loggedIn ? consoleLabel : tr("免费开始", "Start free"))}
            <a className="btn btn-ghost" href={botUrl} target="_blank" rel="noreferrer">
              <TelegramIcon /> {tr("在 Telegram 打开", "Open in Telegram")}
            </a>
          </div>
        </section>
      </main>

      <footer className="footer">
        <div className="container footer-grid">
          <div className="footer-brand">
            <Brand />
            <p>Stay calm, stay ahead.</p>
            <p className="footer-disclaimer">
              {tr(
                "预警仅供参考，可能因网络或节点原因延迟、失败，不构成投资建议。请自行管理仓位风险。",
                "Alerts are for reference only and may be delayed or fail due to network or node issues. Not investment advice. Manage your own position risk.",
              )}
            </p>
          </div>
          <nav className="footer-col">
            <h4>{tr("产品", "Product")}</h4>
            <a href="#features">{tr("功能", "Features")}</a>
            <a href="#loop">{tr("循环贷", "Looping")}</a>
            <a href="#liqmap">{tr("清算地图", "Liq. map")}</a>
            <a href="#trust">{tr("安全", "Security")}</a>
            <a href="#pricing">{tr("定价", "Pricing")}</a>
          </nav>
          <nav className="footer-col">
            <h4>{tr("资源", "Resources")}</h4>
            <Link to="/blog">{tr("博客", "Blog")}</Link>
            <a href="#faq">FAQ</a>
            <a href="#referral">{tr("推荐计划", "Referral")}</a>
            <a href="#about">{tr("开发者", "Developer")}</a>
            <a href="https://www.defigauge.com" target="_blank" rel="noreferrer" title={tr("DefiGauge 工具站", "DefiGauge tools")}>
              www.DefiGauge.com
            </a>
          </nav>
          <nav className="footer-col">
            <h4>{tr("社区", "Community")}</h4>
            <a href={botUrl} target="_blank" rel="noreferrer">
              MechBot Telegram
            </a>
            <a href="https://t.me/+LWTqp4To2ilhY2M1" target="_blank" rel="noreferrer">
              {tr("Telegram 客服群", "Telegram support group")}
            </a>
            <a href="https://github.com/sunningdev/mechbot" target="_blank" rel="noreferrer">
              GitHub
            </a>
          </nav>
          <nav className="footer-col">
            <h4>{tr("法律", "Legal")}</h4>
            <Link to="/terms">{tr("服务条款", "Terms")}</Link>
            <Link to="/privacy">{tr("隐私政策", "Privacy")}</Link>
          </nav>
        </div>
        <div className="footer-bar">
          <div className="container footer-bar-inner">
            <span>© {new Date().getFullYear()} MechBot. {tr("保留所有权利。", "All rights reserved.")}</span>
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
