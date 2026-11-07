import type { Lang } from "./i18n";
import zhTW from "./locales/zh-TW";

type Meta = { title: string; description: string };

const zh: Meta = {
  title: "MechBot · DeFi 清算预警机器人｜Aave、Compound、Morpho 清算提醒",
  description:
    "MechBot 是 DeFi 借贷清算预警机器人：7×24 监控 Aave、Compound、Morpho 仓位的健康因子，快被清算时通过 Telegram 立刻提醒，并算出该还多少、补多少抵押。免费使用。",
};

const en: Meta = {
  title: "MechBot · DeFi Liquidation Alert Bot for Aave, Compound & Morpho",
  description:
    "MechBot is a DeFi liquidation alert bot. It watches the health factor of your Aave, Compound and Morpho positions 24/7 and pings you on Telegram before liquidation, with exactly how much to repay or add. Free to start.",
};

const others: Record<Exclude<Lang, "zh" | "zh-TW" | "en">, Meta> = {
  ru: {
    title: "MechBot · Бот оповещений о ликвидации в DeFi для Aave, Compound и Morpho",
    description:
      "MechBot — бот оповещений о ликвидации в DeFi. Круглосуточно следит за фактором здоровья ваших позиций в Aave, Compound и Morpho и предупреждает в Telegram до ликвидации, сразу показывая, сколько погасить или сколько залога довнести. Можно начать бесплатно.",
  },
  es: {
    title: "MechBot · Bot de alertas de liquidación DeFi para Aave, Compound y Morpho",
    description:
      "MechBot es un bot de alertas de liquidación DeFi. Vigila 24/7 el factor de salud de tus posiciones en Aave, Compound y Morpho y te avisa por Telegram antes de la liquidación, indicando cuánto pagar o cuánto colateral añadir. Empieza gratis.",
  },
  fr: {
    title: "MechBot · Bot d'alerte de liquidation DeFi pour Aave, Compound et Morpho",
    description:
      "MechBot est un bot d'alerte de liquidation DeFi. Il surveille 24h/24 le facteur de santé de vos positions Aave, Compound et Morpho et vous prévient sur Telegram avant la liquidation, en indiquant combien rembourser ou quel collatéral ajouter. Gratuit pour commencer.",
  },
  pt: {
    title: "MechBot · Bot de alertas de liquidação DeFi para Aave, Compound e Morpho",
    description:
      "MechBot é um bot de alertas de liquidação DeFi. Monitora 24/7 o fator de saúde das suas posições na Aave, Compound e Morpho e avisa no Telegram antes da liquidação, dizendo quanto pagar ou quanto colateral adicionar. Comece grátis.",
  },
  vi: {
    title: "MechBot · Bot cảnh báo thanh lý DeFi cho Aave, Compound và Morpho",
    description:
      "MechBot là bot cảnh báo thanh lý DeFi. Theo dõi 24/7 hệ số sức khỏe các vị thế Aave, Compound và Morpho của bạn, báo ngay qua Telegram trước khi bị thanh lý, kèm số tiền cần trả hoặc tài sản thế chấp cần bổ sung. Bắt đầu miễn phí.",
  },
};

export function seoMeta(lang: Lang): Meta {
  switch (lang) {
    case "zh":
      return zh;
    case "zh-TW":
      return { title: zhTW[zh.title] ?? zh.title, description: zhTW[zh.description] ?? zh.description };
    case "en":
      return en;
    default:
      return others[lang];
  }
}

function headTag<T extends HTMLElement>(selector: string, create: () => T) {
  let el = document.head.querySelector<T>(selector);
  if (!el) {
    el = create();
    document.head.appendChild(el);
  }
  return el;
}

export function setDescription(text: string) {
  headTag("meta[name='description']", () => Object.assign(document.createElement("meta"), { name: "description" })).content = text;
}

export function setCanonical(href: string) {
  headTag("link[rel='canonical']", () => Object.assign(document.createElement("link"), { rel: "canonical" })).href = href;
}
