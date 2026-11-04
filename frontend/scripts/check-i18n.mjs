// Reports English source strings in tr() calls that have no ru / vi / es / fr / pt translation.
import { readFileSync, readdirSync, statSync } from "node:fs";
import { join } from "node:path";

const lit = String.raw`("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'|\`(?:[^\`\\]|\\.)*\`)`;
const re = new RegExp(String.raw`\btr\(\s*` + lit + String.raw`\s*,\s*` + lit + String.raw`\s*,?\s*\)`, "g");

function walk(dir, out = []) {
  for (const f of readdirSync(dir)) {
    const p = join(dir, f);
    if (statSync(p).isDirectory()) {
      if (f !== "locales") walk(p, out);
    } else if (/\.tsx?$/.test(f)) out.push(p);
  }
  return out;
}

const keys = new Set();
for (const file of walk("src")) {
  for (const m of readFileSync(file, "utf8").matchAll(re)) {
    const raw = m[2];
    if (raw.startsWith("`")) {
      let i = 0;
      keys.add(raw.slice(1, -1).replace(/\$\{[^}]*\}/g, () => `{${i++}}`));
    } else keys.add(JSON.parse(raw.startsWith("'") ? `"${raw.slice(1, -1).replace(/"/g, '\\"')}"` : raw));
  }
}
for (const m of readFileSync("src/components/FailoverDiagram.tsx", "utf8").matchAll(/log: \["[^"]*", "([^"]*)"\]/g)) keys.add(m[1]);
for (const m of readFileSync("src/i18n.ts", "utf8").matchAll(/: "([A-Z][^"]*)",$/gm)) keys.add(m[1]);

let missing = 0;
for (const lang of ["ru", "vi", "es", "fr", "pt"]) {
  const src = readFileSync(`src/locales/${lang}.ts`, "utf8");
  const dict = new Function(src.replace(/^\/\/.*$/m, "").replace("const " + lang + ": Record<string, string> =", "return").replace(/export default \w+;/, ""))();
  for (const k of keys) {
    if (!(k in dict) && k !== ", " && k !== "、") {
      missing++;
      console.log(`${lang} missing: ${JSON.stringify(k)}`);
    }
  }
  for (const k of Object.keys(dict)) if (!keys.has(k) && !/\{\d\}/.test(k)) console.log(`${lang} unused: ${JSON.stringify(k)}`);
}
console.log(`checked ${keys.size} keys, ${missing} missing`);
