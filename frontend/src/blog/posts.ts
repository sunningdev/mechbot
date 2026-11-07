import { getLang, tr } from "../i18n";
import { esPosts } from "./posts.es";
import { frPosts } from "./posts.fr";
import { ptPosts } from "./posts.pt";
import { ruPosts } from "./posts.ru";
import { viPosts } from "./posts.vi";

export type Block =
  | { type: "p"; text: string }
  | { type: "h"; text: string }
  | { type: "list"; items: string[] }
  | { type: "code"; text: string }
  | { type: "note"; text: string };

export type Post = {
  slug: string;
  title: string;
  summary: string;
  date: string;
  tag: string;
  readMin: number;
  body: Block[];
};

const zhPosts: Post[] = [
  {
    slug: "health-factor",
    title: "健康因子是怎么算的？读懂「距清算」这个数字",
    summary: "Aave、Compound、Morpho 的清算线各叫不同名字，本质是同一件事。用一个公式说清楚，再算出价格还能跌多少。",
    date: "2026-10-28",
    tag: "原理",
    readMin: 5,
    body: [
      { type: "p", text: "借贷协议不关心你借了多少，只关心一件事：抵押物按清算线打折后，还够不够覆盖债务。健康因子（Health Factor，HF）就是这个比值。" },
      { type: "h", text: "一个公式" },
      { type: "code", text: "HF = Σ(抵押价值 × 清算阈值) / 债务价值" },
      { type: "p", text: "HF 大于 1 表示安全，等于 1 就到了清算线，低于 1 任何人都可以清算你的仓位，并拿走一部分抵押物作为奖励。" },
      { type: "h", text: "三个协议，三种叫法" },
      {
        type: "list",
        items: [
          "Aave V3：每个资产有 Liquidation Threshold，合约直接返回健康因子。",
          "Compound V3：每个抵押资产有 liquidateCollateralFactor，我们按同一个公式自己算。",
          "Morpho Blue：每个市场有一个 LLTV（清算贷款价值比），公式相同，只是只有一种抵押物。",
        ],
      },
      { type: "h", text: "「距清算」是什么" },
      { type: "p", text: "控制台里的「距清算」，指的是在债务不变的前提下，抵押物整体价格还能下跌多少才会触发清算：" },
      { type: "code", text: "距清算 = 1 − 1 / HF" },
      { type: "p", text: "HF 为 1.25 时，抵押物还能跌 20%；HF 为 1.07 时，只剩约 6.5%。一根大阴线就够了。" },
      { type: "note", text: "如果你借的是波动资产（比如借 ETH），债务也会涨跌，实际空间会和这个数字有出入。它是一个直观参考，不是精确预言。" },
      { type: "h", text: "预警线设多少合适" },
      { type: "p", text: "稳定币作抵押、借稳定币，可以设得低一些，比如 1.1；ETH、BTC 抵押借稳定币，建议 1.3 到 1.5，给自己留出补仓的时间。" },
    ],
  },
  {
    slug: "rpc-failover",
    title: "节点挂了怎么办？MechBot 的多节点容灾设计",
    summary: "公共 RPC 会限流、会超时、会宕机。我们怎么做到节点出问题时预警照常送达，并且绝不把「查不到」当成「安全」。",
    date: "2026-10-26",
    tag: "工程",
    readMin: 6,
    body: [
      { type: "p", text: "清算预警最怕的不是报错，而是沉默：节点挂了，程序没读到数据，也没发出提醒，用户以为一切正常。MechBot 的容灾设计围绕一条原则展开：读不到数据，绝不当作安全。" },
      { type: "h", text: "每条链都有一组节点" },
      { type: "p", text: "每条链配置多个 RPC 节点：先用你配置的，后面再追加内置的公共备用节点。每次查询按顺序尝试，哪个节点成功了就记住它，下次优先用。" },
      { type: "code", text: "配置节点 → 备用节点 1 → 备用节点 2 → …\n成功：记住当前节点\n失败：自动切到下一个" },
      { type: "h", text: "全部失败时" },
      {
        type: "list",
        items: [
          "这一轮该仓位标记为「查询失败，下轮重试」，控制台里能看到，不会显示成安全。",
          "某个数据源连续多轮对所有地址都失败，值班告警会发到运营的 Telegram。",
          "节点恢复后，会再发一条恢复通知，确认链路已经正常。",
        ],
      },
      { type: "h", text: "程序自己也要扛得住" },
      { type: "p", text: "协议读取、轮询、收款扫描、机器人消息处理，每一处都有独立的崩溃保护。单个仓位的异常数据不会拖垮整轮检查，更不会让服务停掉。" },
      { type: "note", text: "Telegram 接口报错时，日志里会自动隐去机器人令牌，避免密钥意外泄露到日志系统。" },
    ],
  },
  {
    slug: "no-spam",
    title: "为什么 MechBot 不刷屏：一条预警的完整生命周期",
    summary: "群里一天几百条通知，等于没有通知。我们只在跌破预警线时叫你，并且同一仓位一小时内不重复打扰，只有快被清算时才会提醒得更勤。",
    date: "2026-10-24",
    tag: "产品",
    readMin: 4,
    body: [
      { type: "p", text: "很多监控机器人把每次价格波动都推送出来，结果是用户关掉通知，真正危险的那一条也错过了。MechBot 的目标正相反：平时安静，只在清算之前叫醒你。" },
      { type: "h", text: "检查：每 5 分钟一次" },
      { type: "p", text: "后台每 5 分钟读取一次你所有地址在各协议的仓位，直接读链上合约，不依赖第三方数据接口。" },
      { type: "h", text: "判断：只看一条线" },
      {
        type: "list",
        items: [
          "没有借款，或健康因子高于你的预警线：什么都不发。",
          "跌破预警线：立刻私聊提醒，附上协议、链、健康因子和抵押、债务。",
          "一直低于预警线：1 小时内不重复提醒，1 小时后如果仍未恢复，再提醒一次。",
          "健康因子低于 1.05，非常接近清算：改为每 15 分钟提醒一次，直到你处理或仓位恢复。",
          "回到预警线以上：发一条「已恢复安全」，然后重新安静下来。",
        ],
      },
      { type: "code", text: "安全 ──跌破──▶ 预警（私聊）\n  ▲              │ 每 1 小时提醒（低于 1.05 时每 15 分钟）\n  └──恢复── 已恢复安全（一条）" },
      { type: "h", text: "结果" },
      { type: "p", text: "一个仓位从安全到危险再到恢复，你通常只会收到两条消息。每一条都值得你停下手头的事去看一眼。" },
    ],
  },
];

const enPosts: Post[] = [
  {
    slug: "health-factor",
    title: "How is the health factor calculated? Reading the “to liquidation” number",
    summary:
      "Aave, Compound and Morpho name their liquidation lines differently, but it is the same idea. One formula explains it, and tells you how far prices can still fall.",
    date: "2026-10-28",
    tag: "Basics",
    readMin: 5,
    body: [
      {
        type: "p",
        text: "A lending protocol does not care how much you borrowed. It cares about one thing: after discounting your collateral by its liquidation threshold, does it still cover your debt? The health factor (HF) is that ratio.",
      },
      { type: "h", text: "One formula" },
      { type: "code", text: "HF = Σ(collateral value × liquidation threshold) / debt value" },
      {
        type: "p",
        text: "Above 1 is safe. At 1 you hit the liquidation line. Below 1, anyone can liquidate your position and take part of your collateral as a bonus.",
      },
      { type: "h", text: "Three protocols, three names" },
      {
        type: "list",
        items: [
          "Aave V3: each asset has a Liquidation Threshold, and the contract returns the health factor directly.",
          "Compound V3: each collateral asset has a liquidateCollateralFactor; we apply the same formula ourselves.",
          "Morpho Blue: each market has an LLTV (liquidation loan-to-value). Same formula, with a single collateral asset.",
        ],
      },
      { type: "h", text: "What “to liquidation” means" },
      {
        type: "p",
        text: "In the console, “to liquidation” is how much your collateral can fall in price, with debt unchanged, before liquidation triggers:",
      },
      { type: "code", text: "To liquidation = 1 − 1 / HF" },
      {
        type: "p",
        text: "At HF 1.25 your collateral can still drop 20%. At HF 1.07 only about 6.5% is left — one big red candle is enough.",
      },
      {
        type: "note",
        text: "If you borrow a volatile asset (say, ETH), your debt moves too, so the real buffer will differ. Treat this number as an intuitive guide, not a precise prediction.",
      },
      { type: "h", text: "Where to set your alert line" },
      {
        type: "p",
        text: "Stablecoin collateral borrowing stablecoins can sit lower, around 1.1. For ETH or BTC collateral borrowing stablecoins, 1.3 to 1.5 is a good range that leaves time to top up.",
      },
    ],
  },
  {
    slug: "rpc-failover",
    title: "What happens when a node dies? MechBot's multi-node failover",
    summary:
      "Public RPCs rate-limit, time out and go down. Here is how alerts still get delivered when nodes fail — and why “couldn't read” is never treated as “safe”.",
    date: "2026-10-26",
    tag: "Engineering",
    readMin: 6,
    body: [
      {
        type: "p",
        text: "The worst failure for a liquidation alert is not an error, it is silence: a node is down, nothing is read, nothing is sent, and the user assumes everything is fine. MechBot's failover is built around one rule: data we could not read is never treated as safe.",
      },
      { type: "h", text: "Every chain has a pool of nodes" },
      {
        type: "p",
        text: "Each chain has several RPC endpoints: your configured ones first, then built-in public backups. Each query tries them in order, remembers whichever succeeded, and starts there next time.",
      },
      { type: "code", text: "Configured node → backup 1 → backup 2 → …\nSuccess: remember this node\nFailure: move on to the next" },
      { type: "h", text: "When every node fails" },
      {
        type: "list",
        items: [
          "The position is marked “query failed, retrying next round” in the console — never shown as safe.",
          "If a data source fails for every address several rounds in a row, an on-call alert goes to the operator's Telegram.",
          "Once the node recovers, a recovery notice confirms the path is healthy again.",
        ],
      },
      { type: "h", text: "The program has to survive too" },
      {
        type: "p",
        text: "Protocol reads, polling, payment scanning and bot message handling each have their own crash protection. Bad data from one position cannot take down a whole round, let alone the service.",
      },
      {
        type: "note",
        text: "When the Telegram API returns an error, the bot token is stripped from the log so secrets never leak into logging systems.",
      },
    ],
  },
  {
    slug: "no-spam",
    title: "Why MechBot doesn't spam: the life of a single alert",
    summary:
      "Hundreds of notifications a day is the same as none. We only ping you when you cross your alert line, and never twice within an hour for the same position unless it's about to be liquidated.",
    date: "2026-10-24",
    tag: "Product",
    readMin: 4,
    body: [
      {
        type: "p",
        text: "Many monitoring bots push every price wiggle. Users mute them, and then miss the one alert that matters. MechBot aims for the opposite: quiet by default, loud only before liquidation.",
      },
      { type: "h", text: "Check: every 5 minutes" },
      {
        type: "p",
        text: "Every 5 minutes the backend reads all your addresses' positions on each protocol, straight from on-chain contracts, without third-party data APIs.",
      },
      { type: "h", text: "Decide: one line only" },
      {
        type: "list",
        items: [
          "No debt, or health factor above your alert line: nothing is sent.",
          "Drops below the line: an immediate DM with protocol, chain, health factor, collateral and debt.",
          "Stays below the line: no repeat within 1 hour; if it still hasn't recovered after an hour, one more reminder.",
          "Health factor under 1.05, right next to liquidation: reminders every 15 minutes until you act or it recovers.",
          "Back above the line: one “back to safe” message, then quiet again.",
        ],
      },
      { type: "code", text: "Safe ──drops──▶ Alert (DM)\n  ▲               │ hourly reminder (every 15 min under 1.05)\n  └──recovers── Back to safe (one message)" },
      { type: "h", text: "The result" },
      {
        type: "p",
        text: "A position going from safe to danger and back usually costs you just two messages. Each one is worth stopping what you are doing to look.",
      },
    ],
  },
];

const convert = (post: Post, fn: (s: string) => string): Post => ({
  ...post,
  title: fn(post.title),
  summary: fn(post.summary),
  tag: fn(post.tag),
  body: post.body.map((b) => (b.type === "list" ? { ...b, items: b.items.map(fn) } : { ...b, text: fn(b.text) })),
});

export function getPosts() {
  switch (getLang()) {
    case "en":
      return enPosts;
    case "ru":
      return ruPosts;
    case "vi":
      return viPosts;
    case "es":
      return esPosts;
    case "fr":
      return frPosts;
    case "pt":
      return ptPosts;
    case "zh-TW":
      return zhPosts.map((p) => convert(p, (s) => tr(s, s)));
    default:
      return zhPosts;
  }
}

export const findPost = (slug: string | undefined) => getPosts().find((p) => p.slug === slug);
