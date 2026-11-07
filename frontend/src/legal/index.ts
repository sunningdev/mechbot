import type { Block } from "../blog/posts";
import { getLang, tr } from "../i18n";
import { privacy } from "./privacy";
import { terms } from "./terms";

export type LegalDoc = { title: string; summary: string; updated: string; body: Block[] };
export type LegalKind = "terms" | "privacy";
export type LegalSet = Record<"zh" | "en" | "ru" | "es" | "fr" | "pt" | "vi", LegalDoc>;

const convert = (doc: LegalDoc, fn: (s: string) => string): LegalDoc => ({
  title: fn(doc.title),
  summary: fn(doc.summary),
  updated: fn(doc.updated),
  body: doc.body.map((b) => (b.type === "list" ? { ...b, items: b.items.map(fn) } : { ...b, text: fn(b.text) })),
});

export function legalDoc(kind: LegalKind): LegalDoc {
  const docs = kind === "terms" ? terms : privacy;
  const lang = getLang();
  return lang === "zh-TW" ? convert(docs.zh, (s) => tr(s, s)) : docs[lang];
}
