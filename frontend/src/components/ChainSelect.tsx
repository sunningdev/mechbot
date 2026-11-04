import { useEffect, useRef, useState } from "react";
import type { Chain } from "../api";
import { tr } from "../i18n";
import ChainLogo from "./ChainLogo";

type Props = { chains: Chain[]; value: string; onChange: (key: string) => void };

export default function ChainSelect({ chains, value, onChange }: Props) {
  const [open, setOpen] = useState(false);
  const lowest = Math.min(...chains.map((c) => c.minPrice));
  const root = useRef<HTMLDivElement>(null);
  const current = chains.find((c) => c.key === value);
  const onlyLong = tr("仅季付 / 年付", "Quarterly / yearly only");

  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      if (!root.current?.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("mousedown", onDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDown);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  return (
    <div className="chain-select" ref={root}>
      <span className="muted small-text">{tr("充值网络", "Network")}</span>
      <button
        type="button"
        className={`chain-select-btn ${open ? "open" : ""}`}
        aria-haspopup="listbox"
        aria-expanded={open}
        onClick={() => setOpen((v) => !v)}
      >
        {current && <ChainLogo chain={current.key} />}
        <span className="chain-select-name">{current?.name ?? tr("选择网络", "Select network")}</span>
        {current && <span className="muted small-text">{current.token}</span>}
        {current && current.minPrice > lowest && <span className="chain-tag">{onlyLong}</span>}
        <span className="chain-select-caret" aria-hidden />
      </button>
      {open && (
        <ul className="chain-select-list" role="listbox">
          {chains.map((c) => (
            <li key={c.key} role="option" aria-selected={c.key === value}>
              <button
                type="button"
                className={c.key === value ? "on" : ""}
                onClick={() => {
                  onChange(c.key);
                  setOpen(false);
                }}
              >
                <ChainLogo chain={c.key} />
                <span className="chain-select-name">{c.name}</span>
                <span className="muted small-text">{c.token}</span>
                {c.minPrice > lowest && <span className="chain-tag">{onlyLong}</span>}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
