import { langInfo, tr } from "./i18n";

export const shortAddress = (a: string) => `${a.slice(0, 6)}…${a.slice(-4)}`;

export const usd = (v: number) =>
  "$" + v.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 });

// Short dollars for charts: $1.2B, $34.5M, $820K.
export function usdShort(v: number) {
  const units: [number, string][] = [
    [1e9, "B"],
    [1e6, "M"],
    [1e3, "K"],
  ];
  for (const [n, s] of units) {
    if (v >= n) return `$${(v / n).toFixed(v >= n * 100 ? 0 : 1)}${s}`;
  }
  return `$${Math.round(v)}`;
}

export const usdWhole = (v: number) => "$" + Math.round(v).toLocaleString("en-US");

export function tokenAmount(v: number) {
  const digits = v >= 100 ? 0 : v >= 1 ? 2 : 4;
  return v.toLocaleString("en-US", { minimumFractionDigits: digits, maximumFractionDigits: digits });
}

// A token amount with its symbol, or dollars when the token price is unknown.
export function tokenOrUsd(amount: number | undefined, symbol: string | undefined, dollars: number) {
  return amount && symbol ? `${tokenAmount(amount)} ${symbol}` : usd(dollars);
}

export function timeAgo(iso?: string) {
  if (!iso) return tr("等待首次查询", "Waiting for first check");
  const sec = Math.max(0, Math.round((Date.now() - new Date(iso).getTime()) / 1000));
  if (sec < 60) return tr("刚刚更新", "Updated just now");
  if (sec < 3600) {
    const m = Math.round(sec / 60);
    return tr(`${m} 分钟前更新`, `Updated ${m} min ago`);
  }
  const h = Math.round(sec / 3600);
  return tr(`${h} 小时前更新`, `Updated ${h} h ago`);
}

export const dateTime = (iso: string) =>
  new Date(iso).toLocaleString(langInfo().locale, {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });

// Whole days of Pro left, counting a partial day as one; null when Pro is not active.
export function proDaysLeft(pro: boolean, expireAt: string) {
  const ms = pro && expireAt ? Date.parse(expireAt) : NaN;
  return Number.isFinite(ms) ? Math.max(0, Math.ceil((ms - Date.now()) / 86_400_000)) : null;
}

export function riskClass(hf: number, threshold: number) {
  if (hf < 1.1) return "danger";
  if (hf < threshold) return "warn";
  return "ok";
}

export async function copy(text: string) {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    return false;
  }
}
