import type { Chain } from "./api";
import { getLang } from "./i18n";

// Asian and Russian-speaking users mostly hold USDT on TRON / BNB Chain; Western users mostly hold USDC on Base / Arbitrum.
export const usdtFirst = () => ["zh", "zh-TW", "ru", "vi"].includes(getLang());

export const coinOrder = () => (usdtFirst() ? ["USDT", "USDC"] : ["USDC", "USDT"]);

const chainOrder = () =>
  usdtFirst()
    ? ["bsc", "tron", "arbitrum", "base", "optimism", "ethereum"]
    : ["base", "arbitrum", "optimism", "ethereum", "bsc", "tron"];

export function sortByChain<T>(items: T[], key: (item: T) => string) {
  const order = chainOrder();
  const rank = (item: T) => (order.includes(key(item)) ? order.indexOf(key(item)) : order.length);
  return [...items].sort((a, b) => rank(a) - rank(b));
}

export const sortChains = (chains: Chain[]) => sortByChain(chains, (c) => c.key);

// "USDT / USDC (BEP20)" → "USDC / USDT (BEP20)" for USDC-first languages.
export function tokenLabel(token: string) {
  const suffix = token.match(/\s*\(.*\)$/)?.[0] ?? "";
  const coins = token.slice(0, token.length - suffix.length).split("/").map((s) => s.trim()).filter(Boolean);
  const order = coinOrder();
  const rank = (s: string) => (order.includes(s) ? order.indexOf(s) : order.length);
  return coins.sort((a, b) => rank(a) - rank(b)).join(" / ") + suffix;
}
