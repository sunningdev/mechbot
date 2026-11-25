import type { Lang } from "./i18n";
import { languages } from "./i18n";
import zhTW from "./locales/zh-TW";

export type Meta = {
  title: string;
  description: string;
  keywords: string;
};

/** Homepage FAQ pairs used for FAQPage JSON-LD (GEO-friendly Q&A). */
export type SeoFaq = { q: string; a: string };

const zh: Meta = {
  title: "MechBot · Aave 清算预警｜Compound Morpho 健康因子 Telegram 提醒",
  description:
    "免费 DeFi 清算预警机器人：7×24 监控 Aave V3、Compound V3、Morpho Blue 健康因子，清算前 Telegram 提醒，并告诉你该还多少、补多少抵押。无需连接钱包，Ethereum / Arbitrum / Base / Optimism。",
  keywords:
    "Aave 清算预警,Aave 健康因子提醒,Compound 清算提醒,Morpho 清算预警,DeFi 清算机器人,Telegram 清算提醒,借贷清算监控,health factor 预警,免费清算预警",
};

const en: Meta = {
  title: "MechBot · Free Aave Liquidation Alert Bot | Telegram Health Factor Monitor",
  description:
    "Free DeFi liquidation alert bot for Aave V3, Compound V3 and Morpho Blue. Monitor health factor 24/7 and get Telegram alerts before liquidation—with how much to repay or add. No wallet connect. Ethereum, Arbitrum, Base, Optimism.",
  keywords:
    "aave liquidation alert,aave health factor telegram,compound v3 liquidation warning,morpho blue liquidation alert,defi liquidation alert bot,telegram liquidation bot,monitor aave health factor,free defi liquidation monitor,prevent aave liquidation,lending health factor alert",
};

const others: Record<Exclude<Lang, "zh" | "zh-TW" | "en">, Meta> = {
  ru: {
    title: "MechBot · Бесплатный бот алертов ликвидации Aave | Telegram health factor",
    description:
      "Бесплатный бот оповещений о ликвидации DeFi для Aave V3, Compound V3 и Morpho Blue. Следит за health factor 24/7 и пишет в Telegram до ликвидации — сколько погасить или сколько залога добавить. Без подключения кошелька.",
    keywords:
      "оповещение ликвидации Aave,health factor Telegram,бот ликвидации Compound,алерт Morpho Blue,DeFi бот ликвидации,мониторинг health factor",
  },
  es: {
    title: "MechBot · Bot gratis de alertas de liquidación Aave | Telegram health factor",
    description:
      "Bot gratis de alertas de liquidación DeFi para Aave V3, Compound V3 y Morpho Blue. Vigila el health factor 24/7 y avisa por Telegram antes de la liquidación, con cuánto pagar o cuánto colateral añadir. Sin conectar wallet.",
    keywords:
      "alerta liquidación Aave,health factor Telegram,aviso liquidación Compound,alerta Morpho Blue,bot liquidación DeFi,monitor health factor",
  },
  fr: {
    title: "MechBot · Bot gratuit d'alerte liquidation Aave | Health factor Telegram",
    description:
      "Bot gratuit d'alerte de liquidation DeFi pour Aave V3, Compound V3 et Morpho Blue. Surveille le health factor 24h/24 et prévient sur Telegram avant liquidation — montant à rembourser ou collatéral à ajouter. Sans connexion de wallet.",
    keywords:
      "alerte liquidation Aave,health factor Telegram,alerte Compound V3,alerte Morpho Blue,bot liquidation DeFi,surveiller health factor",
  },
  pt: {
    title: "MechBot · Bot grátis de alerta de liquidação Aave | Health factor no Telegram",
    description:
      "Bot grátis de alertas de liquidação DeFi para Aave V3, Compound V3 e Morpho Blue. Monitora o health factor 24/7 e avisa no Telegram antes da liquidação — quanto pagar ou quanto colateral adicionar. Sem conectar carteira.",
    keywords:
      "alerta liquidação Aave,health factor Telegram,aviso liquidação Compound,alerta Morpho Blue,bot liquidação DeFi,monitorar health factor",
  },
  vi: {
    title: "MechBot · Bot cảnh báo thanh lý Aave miễn phí | Health factor Telegram",
    description:
      "Bot cảnh báo thanh lý DeFi miễn phí cho Aave V3, Compound V3 và Morpho Blue. Theo dõi health factor 24/7, báo qua Telegram trước khi bị thanh lý — kèm số cần trả hoặc tài sản thế chấp cần bổ sung. Không cần kết nối ví.",
    keywords:
      "cảnh báo thanh lý Aave,health factor Telegram,cảnh báo Compound V3,cảnh báo Morpho Blue,bot thanh lý DeFi,theo dõi health factor",
  },
};

export function seoMeta(lang: Lang): Meta {
  switch (lang) {
    case "zh":
      return zh;
    case "zh-TW":
      return {
        title: zhTW[zh.title] ?? zh.title,
        description: zhTW[zh.description] ?? zh.description,
        keywords: zhTW[zh.keywords] ?? zh.keywords,
      };
    case "en":
      return en;
    default:
      return others[lang];
  }
}

/** Long-tail FAQ content for JSON-LD / GEO (answers match product facts). */
export function seoFaqs(lang: Lang): SeoFaq[] {
  const enFaqs: SeoFaq[] = [
    {
      q: "How do I get Telegram alerts before an Aave liquidation?",
      a: "Add your public wallet address in MechBot after Telegram login. MechBot checks Aave V3 health factor about every 5 minutes and DMs you on Telegram when it crosses your alert line, including how much to repay or add as collateral.",
    },
    {
      q: "Is there a free DeFi liquidation alert bot for Aave, Compound and Morpho?",
      a: "Yes. MechBot’s free plan monitors Aave V3. Pro unlocks Compound V3, Morpho Blue, more addresses, earlier Binance-price warnings and the live liquidation map. No wallet connect or private keys required.",
    },
    {
      q: "Can MechBot monitor Morpho Blue and Compound V3 health factor?",
      a: "Yes on Ethereum, Arbitrum, Base and Optimism. Free covers Aave V3; Pro also watches Compound V3 and Morpho Blue positions on those chains.",
    },
    {
      q: "Do I need to connect my wallet for liquidation alerts?",
      a: "No. MechBot only needs a public 0x address. It never asks for signatures, approvals or private keys.",
    },
  ];

  if (lang === "en") return enFaqs;

  const localized: Partial<Record<Lang, SeoFaq[]>> = {
    zh: [
      {
        q: "如何在 Aave 清算前收到 Telegram 提醒？",
        a: "用 Telegram 登录 MechBot 后添加公开钱包地址。系统约每 5 分钟检查 Aave V3 健康因子，跌破你的预警线时私聊提醒，并给出该还多少或补多少抵押。",
      },
      {
        q: "有没有免费的 Aave、Compound、Morpho 清算预警机器人？",
        a: "有。MechBot 免费版监控 Aave V3；Pro 解锁 Compound V3、Morpho Blue、更多地址、币安价格提前预警和清算地图。无需连接钱包或提供私钥。",
      },
      {
        q: "MechBot 能监控 Morpho Blue 和 Compound V3 健康因子吗？",
        a: "可以，支持 Ethereum、Arbitrum、Base、Optimism。免费版覆盖 Aave V3；Pro 同时监控 Compound V3 与 Morpho Blue。",
      },
      {
        q: "清算预警需要连接钱包吗？",
        a: "不需要。只要公开 0x 地址，无需签名、授权或私钥。",
      },
    ],
    ru: [
      {
        q: "Как получать алерты в Telegram до ликвидации на Aave?",
        a: "Войдите через Telegram и добавьте публичный адрес. MechBot проверяет health factor Aave V3 примерно каждые 5 минут и пишет в Telegram при пересечении вашей линии, с суммой погашения или залога.",
      },
      {
        q: "Есть ли бесплатный бот алертов ликвидации для Aave, Compound и Morpho?",
        a: "Да. Бесплатный план MechBot следит за Aave V3. Pro открывает Compound V3, Morpho Blue, больше адресов и ранние предупреждения по цене Binance. Кошелёк подключать не нужно.",
      },
      {
        q: "Может ли MechBot следить за Morpho Blue и Compound V3?",
        a: "Да, на Ethereum, Arbitrum, Base и Optimism. Free — Aave V3; Pro также смотрит Compound V3 и Morpho Blue.",
      },
      {
        q: "Нужно ли подключать кошелёк для алертов ликвидации?",
        a: "Нет. Достаточно публичного 0x-адреса. Подписи, approve и приватные ключи не нужны.",
      },
    ],
    es: [
      {
        q: "¿Cómo recibo alertas de Telegram antes de una liquidación en Aave?",
        a: "Inicia sesión con Telegram y añade tu dirección pública. MechBot revisa el health factor de Aave V3 cada ~5 minutos y te avisa por DM al cruzar tu umbral, indicando cuánto pagar o añadir de colateral.",
      },
      {
        q: "¿Hay un bot gratis de alertas de liquidación DeFi para Aave, Compound y Morpho?",
        a: "Sí. El plan gratis de MechBot monitoriza Aave V3. Pro desbloquea Compound V3, Morpho Blue, más direcciones y avisos tempranos con precio de Binance. Sin conectar wallet.",
      },
      {
        q: "¿Puede MechBot vigilar Morpho Blue y Compound V3?",
        a: "Sí en Ethereum, Arbitrum, Base y Optimism. Gratis cubre Aave V3; Pro también Compound V3 y Morpho Blue.",
      },
      {
        q: "¿Necesito conectar la wallet para alertas de liquidación?",
        a: "No. Solo una dirección 0x pública. Sin firmas, approvals ni claves privadas.",
      },
    ],
    fr: [
      {
        q: "Comment recevoir des alertes Telegram avant une liquidation Aave ?",
        a: "Connectez-vous via Telegram et ajoutez votre adresse publique. MechBot vérifie le health factor Aave V3 toutes les ~5 minutes et vous DM sur Telegram sous votre seuil, avec le montant à rembourser ou le collatéral à ajouter.",
      },
      {
        q: "Existe-t-il un bot gratuit d'alerte de liquidation DeFi pour Aave, Compound et Morpho ?",
        a: "Oui. L'offre gratuite MechBot surveille Aave V3. Pro débloque Compound V3, Morpho Blue, plus d'adresses et des alertes anticipées via le prix Binance. Sans connexion de wallet.",
      },
      {
        q: "MechBot peut-il surveiller Morpho Blue et Compound V3 ?",
        a: "Oui sur Ethereum, Arbitrum, Base et Optimism. Gratuit = Aave V3 ; Pro ajoute Compound V3 et Morpho Blue.",
      },
      {
        q: "Faut-il connecter un wallet pour les alertes de liquidation ?",
        a: "Non. Une adresse 0x publique suffit. Pas de signature, d'approbation ni de clé privée.",
      },
    ],
    pt: [
      {
        q: "Como receber alertas no Telegram antes de uma liquidação na Aave?",
        a: "Entre com o Telegram e adicione seu endereço público. O MechBot checa o health factor da Aave V3 a cada ~5 minutos e avisa no Telegram ao cruzar sua linha, com quanto pagar ou quanto colateral adicionar.",
      },
      {
        q: "Existe um bot grátis de alerta de liquidação DeFi para Aave, Compound e Morpho?",
        a: "Sim. O plano grátis do MechBot monitora Aave V3. O Pro libera Compound V3, Morpho Blue, mais endereços e avisos antecipados com preço da Binance. Sem conectar carteira.",
      },
      {
        q: "O MechBot monitora Morpho Blue e Compound V3?",
        a: "Sim em Ethereum, Arbitrum, Base e Optimism. Grátis cobre Aave V3; Pro também Compound V3 e Morpho Blue.",
      },
      {
        q: "Preciso conectar a carteira para alertas de liquidação?",
        a: "Não. Só um endereço 0x público. Sem assinaturas, approvals ou chaves privadas.",
      },
    ],
    vi: [
      {
        q: "Làm sao nhận cảnh báo Telegram trước khi bị thanh lý trên Aave?",
        a: "Đăng nhập Telegram và thêm địa chỉ ví công khai. MechBot kiểm tra health factor Aave V3 khoảng mỗi 5 phút và nhắn Telegram khi vượt ngưỡng cảnh báo, kèm số cần trả hoặc tài sản thế chấp cần bổ sung.",
      },
      {
        q: "Có bot cảnh báo thanh lý DeFi miễn phí cho Aave, Compound và Morpho không?",
        a: "Có. Gói miễn phí của MechBot theo dõi Aave V3. Pro mở khóa Compound V3, Morpho Blue, thêm địa chỉ và cảnh báo sớm theo giá Binance. Không cần kết nối ví.",
      },
      {
        q: "MechBot có theo dõi Morpho Blue và Compound V3 không?",
        a: "Có trên Ethereum, Arbitrum, Base và Optimism. Miễn phí gồm Aave V3; Pro thêm Compound V3 và Morpho Blue.",
      },
      {
        q: "Cảnh báo thanh lý có cần kết nối ví không?",
        a: "Không. Chỉ cần địa chỉ 0x công khai. Không chữ ký, không phê duyệt, không khóa riêng.",
      },
    ],
  };

  if (lang === "zh-TW") {
    return (localized.zh ?? enFaqs).map((f) => ({
      q: zhTW[f.q] ?? f.q,
      a: zhTW[f.a] ?? f.a,
    }));
  }
  return localized[lang] ?? enFaqs;
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

export function setKeywords(text: string) {
  headTag("meta[name='keywords']", () => Object.assign(document.createElement("meta"), { name: "keywords" })).content = text;
}

export function setCanonical(href: string) {
  headTag("link[rel='canonical']", () => Object.assign(document.createElement("link"), { rel: "canonical" })).href = href;
}

function setMetaProperty(property: string, content: string) {
  headTag(`meta[property='${property}']`, () => {
    const el = document.createElement("meta");
    el.setAttribute("property", property);
    return el;
  }).setAttribute("content", content);
}

function setMetaName(name: string, content: string) {
  headTag(`meta[name='${name}']`, () => Object.assign(document.createElement("meta"), { name })).content = content;
}

export function setOpenGraph(opts: { title: string; description: string; url: string; locale: string }) {
  setMetaProperty("og:type", "website");
  setMetaProperty("og:site_name", "MechBot");
  setMetaProperty("og:title", opts.title);
  setMetaProperty("og:description", opts.description);
  setMetaProperty("og:url", opts.url);
  setMetaProperty("og:locale", opts.locale.replace("-", "_"));
  setMetaName("twitter:card", "summary_large_image");
  setMetaName("twitter:title", opts.title);
  setMetaName("twitter:description", opts.description);
}

export function setHreflang(pathname: string) {
  document.head.querySelectorAll("link[rel='alternate'][hreflang]").forEach((el) => el.remove());
  const origin = window.location.origin;
  const path = pathname || "/";
  for (const l of languages) {
    const href = l.code === "en" ? `${origin}${path}` : `${origin}${path}?lang=${l.code}`;
    const link = document.createElement("link");
    link.rel = "alternate";
    link.hreflang = l.html;
    link.href = href;
    document.head.appendChild(link);
  }
  const def = document.createElement("link");
  def.rel = "alternate";
  def.hreflang = "x-default";
  def.href = `${origin}${path}`;
  document.head.appendChild(def);
}

export function setJsonLd(data: object | object[]) {
  let el = document.head.querySelector<HTMLScriptElement>("script[data-seo-jsonld]");
  if (!el) {
    el = document.createElement("script");
    el.type = "application/ld+json";
    el.setAttribute("data-seo-jsonld", "1");
    document.head.appendChild(el);
  }
  el.textContent = JSON.stringify(data);
}

export function buildHomeJsonLd(lang: Lang, meta: Meta, pageUrl: string) {
  const faqs = seoFaqs(lang);
  const app = {
    "@type": "SoftwareApplication",
    name: "MechBot",
    alternateName: [
      "MechBot liquidation alert bot",
      "MechBot Aave alert",
      "MechBot 清算预警机器人",
      "MechBot бот ликвидации",
    ],
    url: "https://www.mechbot.pro/",
    applicationCategory: "FinanceApplication",
    operatingSystem: "Telegram, Web",
    description: meta.description,
    inLanguage: languages.map((l) => l.html),
    offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
    featureList: [
      "Aave V3 health factor monitoring",
      "Compound V3 and Morpho Blue alerts (Pro)",
      "Telegram liquidation alerts",
      "Repay and collateral guidance",
      "No wallet connect",
    ],
  };
  const faqPage = {
    "@type": "FAQPage",
    mainEntity: faqs.map((f) => ({
      "@type": "Question",
      name: f.q,
      acceptedAnswer: { "@type": "Answer", text: f.a },
    })),
  };
  const org = {
    "@type": "Organization",
    name: "MechBot",
    url: "https://www.mechbot.pro/",
    sameAs: ["https://github.com/sunningdev/mechbot", "https://www.defigauge.com"],
  };
  const web = {
    "@type": "WebSite",
    name: "MechBot",
    url: pageUrl,
    inLanguage: languages.find((l) => l.code === lang)?.html ?? "en",
    description: meta.description,
  };
  return { "@context": "https://schema.org", "@graph": [app, faqPage, org, web] };
}

export function applyPageSeo(opts: {
  lang: Lang;
  pathname: string;
  title?: string;
  description?: string;
  home?: boolean;
}) {
  const meta = seoMeta(opts.lang);
  const title = opts.title ?? meta.title;
  const description = opts.description ?? meta.description;
  const info = languages.find((l) => l.code === opts.lang)!;
  const path = opts.pathname || "/";
  const canonical =
    window.location.origin + path + (opts.lang === "en" ? "" : `?lang=${opts.lang}`);

  document.title = title;
  document.documentElement.lang = info.html;
  setDescription(description);
  setKeywords(meta.keywords);
  setCanonical(canonical);
  setOpenGraph({ title, description, url: canonical, locale: info.locale });
  setHreflang(path);
  if (opts.home) {
    setJsonLd(buildHomeJsonLd(opts.lang, meta, canonical));
  }
}
