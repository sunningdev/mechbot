import { tr } from "../i18n";

type FlowNode = { y: number; title: string; sub: string; tone?: "danger" | "ok" };

const sources: FlowNode[] = [
  { y: 60, title: "Aave V3", sub: "Ethereum · HF 1.06", tone: "danger" },
  { y: 160, title: "Compound V3", sub: "Arbitrum · HF 1.84" },
  { y: 260, title: "Morpho Blue", sub: "Ethereum · HF 2.31" },
];

const results = (): FlowNode[] => [
  { y: 90, title: tr("Telegram 预警", "Telegram alert"), sub: tr("1.06 < 阈值 1.50", "1.06 < threshold 1.50"), tone: "danger" },
  { y: 230, title: tr("恢复安全", "Back to safe"), sub: tr("HF 1.62 · 发一次安全通知", "HF 1.62 · one safe notice"), tone: "ok" },
];

const W = 1000;
const H = 320;
const CENTER_Y = 160;

const inPath = (y: number) => `M200,${y} C300,${y} 300,${CENTER_Y} 400,${CENTER_Y}`;
const outPath = (y: number) => `M600,${CENTER_Y} C700,${CENTER_Y} 700,${y} 800,${y}`;
const top = (y: number) => `${(y / H) * 100}%`;

export default function LiquidationFlow() {
  return (
    <div className="flow-scroll">
      <div className="flow">
        <svg viewBox={`0 0 ${W} ${H}`} aria-hidden="true">
          {sources.map((s, i) => (
            <g key={s.title}>
              <path id={`flow-in-${i}`} className="flow-line" d={inPath(s.y)} />
              <circle className="flow-dot" r="4">
                <animateMotion dur="2.4s" begin={`${i * 0.8}s`} repeatCount="indefinite">
                  <mpath href={`#flow-in-${i}`} />
                </animateMotion>
              </circle>
            </g>
          ))}
          {results().map((r, i) => (
            <g key={r.y}>
              <path id={`flow-out-${i}`} className="flow-line" d={outPath(r.y)} />
              <circle className={`flow-dot ${r.tone ?? ""}`} r="4">
                <animateMotion dur="2.4s" begin={`${1.2 + i * 1.2}s`} repeatCount="indefinite">
                  <mpath href={`#flow-out-${i}`} />
                </animateMotion>
              </circle>
            </g>
          ))}
        </svg>

        {sources.map((s) => (
          <div className="flow-node frame" key={s.title} style={{ left: 0, top: top(s.y) }}>
            <strong>{s.title}</strong>
            <span className={s.tone}>{s.sub}</span>
          </div>
        ))}

        <div className="flow-node flow-center" style={{ left: "40%", top: top(CENTER_Y) }}>
          <img className="logo-dark" src="/logo.svg" alt="" />
          <div>
            <strong>MechBot</strong>
            <span>{tr("每 5 分钟检查健康因子", "Checks health factors every 5 min")}</span>
          </div>
        </div>

        {results().map((r) => (
          <div className={`flow-node frame ${r.tone ?? ""}`} key={r.y} style={{ left: "80%", top: top(r.y) }}>
            <strong>{r.title}</strong>
            <span>{r.sub}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
