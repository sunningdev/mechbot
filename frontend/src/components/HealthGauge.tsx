import { useEffect, useState } from "react";
import { tr } from "../i18n";

const MIN = 1.0;
const MAX = 2.0;
const THRESHOLD = 1.2;
const DEBT = 12000;
const PERIOD_MS = 14000;
const TICKS = Array.from({ length: 11 }, (_, i) => MIN + i * 0.1);

function sample(t: number) {
  return 1.4 + 0.3 * Math.cos((2 * Math.PI * t) / PERIOD_MS);
}

const pct = (v: number) => `${((v - MIN) / (MAX - MIN)) * 100}%`;
const usd = (n: number) => `$${Math.round(n).toLocaleString("en-US")}`;

export default function HealthGauge() {
  const [hf, setHf] = useState(() => sample(0));

  useEffect(() => {
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      setHf(1.18);
      return;
    }
    const start = performance.now();
    const timer = window.setInterval(() => setHf(sample(performance.now() - start)), 120);
    return () => window.clearInterval(timer);
  }, []);

  const alert = hf < THRESHOLD;
  const drop = (1 - 1 / hf) * 100;

  return (
    <div className={`gauge frame ${alert ? "alert" : "safe"}`} aria-label={tr("健康因子仪表示意", "Health factor gauge demo")}>
      <div className="gauge-head">
        <span>POSITION · AAVE V3 · ETHEREUM</span>
        <span className="gauge-live">
          <i className="signal-dot" /> LIVE
        </span>
      </div>

      <div className="gauge-readout">
        <div>
          <span className="micro">HEALTH FACTOR</span>
          <strong className="gauge-tone">{hf.toFixed(3)}</strong>
        </div>
        <div className={`gauge-state ${alert ? "on" : ""}`}>
          {alert ? tr("预警已发送 → TELEGRAM", "ALERT SENT → TELEGRAM") : tr("安全 · 持续监控", "SAFE · MONITORING")}
        </div>
      </div>

      <div className="scale">
        <div className="scale-liq" style={{ width: pct(1.1) }} />
        {TICKS.map((v, i) => (
          <i key={v} className={i % 5 === 0 ? "tick major" : "tick"} style={{ left: pct(v) }} />
        ))}
        <div className="scale-threshold" style={{ left: pct(THRESHOLD) }}>
          <span>{tr("阈值 1.20", "THRESHOLD 1.20")}</span>
        </div>
        <div className="scale-needle" style={{ left: pct(hf) }} />
      </div>
      <div className="scale-labels">
        <span>1.00 LIQ</span>
        <span>1.50</span>
        <span>2.00</span>
      </div>

      <div className="gauge-stats">
        <div>
          <span className="micro">{tr("抵押（加权）", "Collateral (weighted)")}</span>
          <strong>{usd(hf * DEBT)}</strong>
        </div>
        <div>
          <span className="micro">{tr("债务", "Debt")}</span>
          <strong>{usd(DEBT)}</strong>
        </div>
        <div>
          <span className="micro">{tr("距清算", "To liquidation")}</span>
          <strong className="gauge-tone">−{drop.toFixed(1)}%</strong>
        </div>
      </div>
    </div>
  );
}
