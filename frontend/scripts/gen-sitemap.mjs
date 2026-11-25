// Builds public/sitemap.xml: one canonical URL per page (English), with hreflang alternates.
// Listing every ?lang= URL as its own <loc> floods Google with near-duplicates and
// often ends up as "Discovered – currently not indexed".
import { readFileSync, writeFileSync } from "node:fs";

const SITE = "https://www.mechbot.pro";
const langs = [
  ["en", ""],
  ["zh-CN", "?lang=zh"],
  ["zh-TW", "?lang=zh-TW"],
  ["ru", "?lang=ru"],
  ["es", "?lang=es"],
  ["fr", "?lang=fr"],
  ["pt-BR", "?lang=pt"],
  ["vi", "?lang=vi"],
];

const slugs = [...new Set([...readFileSync("src/blog/posts.ts", "utf8").matchAll(/slug: "([^"]+)"/g)].map((m) => m[1]))];
const paths = ["/", "/blog", ...slugs.map((s) => `/blog/${s}`), "/terms", "/privacy"];

let out = '<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">\n';
for (const path of paths) {
  out += `  <url>\n    <loc>${SITE}${path}</loc>\n`;
  for (const [hreflang, q] of langs) {
    out += `    <xhtml:link rel="alternate" hreflang="${hreflang}" href="${SITE}${path}${q}"/>\n`;
  }
  out += `    <xhtml:link rel="alternate" hreflang="x-default" href="${SITE}${path}"/>\n`;
  out += `  </url>\n`;
}
out += "</urlset>\n";
writeFileSync("public/sitemap.xml", out);
console.log(`sitemap: ${paths.length} canonical pages (${langs.length} hreflang each)`);
