import type { CSSProperties, ReactNode } from "react";
import { tr } from "../i18n";

type Tone = "ok" | "danger";

export type Stat = {
  label: string;
  value: ReactNode;
  hint?: ReactNode;
  className?: string;
  tag?: { text: string; tone: Tone };
  aside?: ReactNode;
};

type Props = { title: ReactNode; sub: ReactNode; stats: Stat[] };

export default function PageHead({ title, sub, stats }: Props) {
  return (
    <section className="container console-head">
      <div>
        <h1>{title}</h1>
        <p className="muted">{sub}</p>
      </div>
      <dl className="console-stats frame" style={{ "--cols": stats.length } as CSSProperties}>
        {stats.map((s) => (
          <div key={s.label} className="stat">
            <div className="stat-body">
              <dt>
                {s.label}
                {s.tag && <span className={`stat-tag ${s.tag.tone}`}>{s.tag.text}</span>}
              </dt>
              <dd className="stat-main">
                <span className={`stat-value ${s.className ?? ""}`}>{s.value}</span>
                {s.hint && <span className="stat-hint">{s.hint}</span>}
              </dd>
            </div>
            {s.aside && <div className="stat-aside">{s.aside}</div>}
          </div>
        ))}
      </dl>
    </section>
  );
}

const GAUGE_MIN = 1;
const GAUGE_MAX = 2;

export function HfGauge({ value, threshold }: { value: number; threshold: number }) {
  const pos = (v: number) => (Math.min(GAUGE_MAX, Math.max(GAUGE_MIN, v)) - GAUGE_MIN) / (GAUGE_MAX - GAUGE_MIN);
  const t = pos(threshold) * 100;
  return (
    <span className="hf-gauge" title={tr(
        `清算线 1.0 · 预警线 ${threshold} · 当前 ${value.toFixed(3)}`,
        `Liquidation 1.0 · alert line ${threshold} · now ${value.toFixed(3)}`,
      )} aria-hidden>
      <span className="hf-gauge-track" style={{ "--t": `${t}%` } as CSSProperties}>
        <i className="hf-gauge-mark" style={{ left: `${pos(value) * 100}%` }} />
      </span>
      <span className="hf-gauge-scale">
        <span>1.0</span>
        <span>2.0</span>
      </span>
    </span>
  );
}

export function Ring({ ratio }: { ratio: number }) {
  const r = 16;
  const c = 2 * Math.PI * r;
  const p = Math.min(1, Math.max(0, ratio));
  return (
    <svg className="ring" viewBox="0 0 40 40" aria-hidden>
      <circle cx="20" cy="20" r={r} className="ring-track" />
      <circle
        cx="20"
        cy="20"
        r={r}
        className="ring-fill"
        strokeDasharray={`${p * c} ${c}`}
        transform="rotate(-90 20 20)"
      />
      <text x="20" y="20" dominantBaseline="central" textAnchor="middle">
        {Math.round(p * 100)}%
      </text>
    </svg>
  );
}
