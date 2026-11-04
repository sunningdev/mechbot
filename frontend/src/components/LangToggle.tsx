import { useEffect, useRef, useState } from "react";
import { languages, setLang, useLang, type Lang } from "../i18n";

const short: Record<Lang, string> = { zh: "简", "zh-TW": "繁", en: "EN", ru: "RU", vi: "VI", es: "ES", fr: "FR", pt: "PT" };

export default function LangToggle() {
  const lang = useLang();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const onDown = (e: PointerEvent) => {
      if (!ref.current?.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && setOpen(false);
    document.addEventListener("pointerdown", onDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("pointerdown", onDown);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  return (
    <div className="lang-menu" ref={ref}>
      <button
        type="button"
        className="lang-toggle"
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-label="Language"
        onClick={() => setOpen((v) => !v)}
      >
        <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true">
          <circle cx="12" cy="12" r="9" />
          <path d="M3 12h18M12 3c2.5 2.6 3.8 5.6 3.8 9s-1.3 6.4-3.8 9c-2.5-2.6-3.8-5.6-3.8-9S9.5 5.6 12 3z" />
        </svg>
        <span>{short[lang]}</span>
      </button>
      {open && (
        <ul className="lang-list" role="listbox" aria-label="Language">
          {languages.map((l) => (
            <li key={l.code}>
              <button
                type="button"
                role="option"
                aria-selected={l.code === lang}
                lang={l.html}
                className={l.code === lang ? "on" : ""}
                onClick={() => {
                  setLang(l.code);
                  setOpen(false);
                }}
              >
                <span className="lang-check">{l.code === lang ? "✓" : ""}</span>
                {l.label}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
