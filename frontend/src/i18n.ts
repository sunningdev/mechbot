import { useSyncExternalStore } from "react";
import es from "./locales/es";
import fr from "./locales/fr";
import pt from "./locales/pt";
import ru from "./locales/ru";
import vi from "./locales/vi";
import zhTW from "./locales/zh-TW";

export type Lang = "zh" | "zh-TW" | "en" | "ru" | "vi" | "es" | "fr" | "pt";

export const languages: { code: Lang; label: string; html: string; locale: string }[] = [
  { code: "en", label: "English", html: "en", locale: "en-US" },
  { code: "ru", label: "Русский", html: "ru", locale: "ru-RU" },
  { code: "es", label: "Español", html: "es", locale: "es-ES" },
  { code: "zh", label: "简体中文", html: "zh-CN", locale: "zh-CN" },
  { code: "zh-TW", label: "繁體中文", html: "zh-TW", locale: "zh-TW" },
  { code: "vi", label: "Tiếng Việt", html: "vi", locale: "vi-VN" },
  { code: "fr", label: "Français", html: "fr", locale: "fr-FR" },
  { code: "pt", label: "Português", html: "pt-BR", locale: "pt-BR" },
];

const KEY = "mechbot_lang";
const listeners = new Set<() => void>();
const isLang = (v: unknown): v is Lang => languages.some((l) => l.code === v);

// English unless the visitor picked another language before.
function initial(): Lang {
  const saved = localStorage.getItem(KEY);
  return isLang(saved) ? saved : "en";
}

let current: Lang = initial();

type Dict = Record<string, string>;
// ru / vi / es / fr / pt are keyed by the English text, zh-TW by the Simplified Chinese text.
const dicts: Partial<Record<Lang, Dict>> = { ru, vi, es, fr, pt, "zh-TW": zhTW };
const patternCache = new Map<Dict, [RegExp, string][]>();

const escapeRe = (s: string) => s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");

function patterns(dict: Dict) {
  let list = patternCache.get(dict);
  if (!list) {
    list = Object.keys(dict)
      .filter((k) => /\{\d\}/.test(k))
      .map((k) => {
        const source = k.split(/\{\d\}/).map(escapeRe).join("(.+?)");
        return [new RegExp(`^${source}$`, "s"), k] as [RegExp, string];
      });
    patternCache.set(dict, list);
  }
  return list;
}

// Keys with {0}, {1}… match runtime-formatted text.
function translate(dict: Dict, source: string) {
  const exact = dict[source];
  if (exact !== undefined) return exact;
  for (const [re, key] of patterns(dict)) {
    const m = source.match(re);
    if (!m) continue;
    const slots = key.match(/\{\d\}/g) ?? [];
    let out = dict[key];
    slots.forEach((slot, i) => {
      out = out.replace(slot, m[i + 1]);
    });
    return out;
  }
  return source;
}

export const getLang = () => current;

export const langInfo = (lang: Lang = current) => languages.find((l) => l.code === lang)!;

export const isChinese = () => current === "zh" || current === "zh-TW";

export function tr(zh: string, en: string) {
  switch (current) {
    case "zh":
      return zh;
    case "zh-TW":
      return translate(zhTW, zh);
    case "en":
      return en;
    default:
      return translate(dicts[current]!, en);
  }
}

function applyDocument(lang: Lang) {
  document.documentElement.lang = langInfo(lang).html;
  document.title = "MechBot · " + tr("DeFi 清算预警", "DeFi liquidation alerts");
}

applyDocument(current);

export function setLang(lang: Lang) {
  if (lang === current) return;
  current = lang;
  localStorage.setItem(KEY, lang);
  applyDocument(lang);
  listeners.forEach((fn) => fn());
}

const subscribe = (fn: () => void) => {
  listeners.add(fn);
  return () => listeners.delete(fn);
};

export const useLang = () => useSyncExternalStore(subscribe, getLang);

const planNames: Record<string, [string, string]> = {
  month: ["月付", "Monthly"],
  quarter: ["季付", "Quarterly"],
  year: ["年付", "Yearly"],
};

export const planName = (key: string, fallback = key) => {
  const p = planNames[key];
  return p ? tr(p[0], p[1]) : fallback;
};

const serverErrors: Record<string, string> = {
  接口不存在: "Endpoint not found",
  需要JSON请求: "JSON request required",
  请先登录: "Please log in first",
  "登录已过期，请重新登录": "Session expired, please log in again",
  生成登录链接失败: "Could not create a login link",
  登录失败: "Login failed",
  读取地址失败: "Could not load addresses",
  读取仓位失败: "Could not load positions",
  请输入0x开头的钱包地址: "Enter a wallet address starting with 0x",
  这个地址已经在监控里: "This address is already being watched",
  添加失败: "Could not add the address",
  地址格式不对: "Invalid address format",
  监控列表里没有这个地址: "This address is not in your watch list",
  删除失败: "Could not delete the address",
  "阈值要大于1，且不超过10": "Threshold must be above 1 and at most 10",
  设置失败: "Could not save the setting",
  读取充值信息失败: "Could not load payment info",
  充值暂未开放: "USDT payments are not open yet",
  "暂时无法检查，请稍后再试": "Cannot check right now, please try again later",
  读取推荐数据失败: "Could not load referral data",
  读取购买记录失败: "Could not load purchase history",
  请求格式不对: "Invalid request format",
  请求失败: "Request failed",
  查询失败: "Query failed",
  登录链接无效或已过期: "Login link is invalid or expired",
  地址已在监控中: "Address is already being watched",
};

export function translateServer(msg: string) {
  if (isChinese()) return tr(msg, msg);
  const exact = serverErrors[msg.replace(/\s+/g, "")];
  if (exact) return tr(msg, exact);
  let m = msg.match(/当前版本最多\s*(\d+)\s*个地址/);
  if (m) return tr(msg, `Your plan allows up to ${m[1]} addresses`);
  m = msg.match(/刚检查过，(\d+)\s*秒后再试/);
  if (m) return tr(msg, `Just checked, try again in ${m[1]}s`);
  return msg;
}
