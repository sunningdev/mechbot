import { useEffect, useState } from "react";
import { tr } from "../i18n";

type Tone = "ok" | "warn" | "";
type Phase = {
  active: number; // index of the node carrying traffic, -1 while switching
  down: number[];
  send: "idle" | "fail" | "ok";
  log: [string, string];
  tone: Tone;
};

const NODES = ["RPC-1", "RPC-2", "RPC-3"];

const PHASES: Phase[] = [
  { active: 0, down: [], send: "idle", log: ["RPC-1 读取 Aave V3 · HF 1.42", "RPC-1 read Aave V3 · HF 1.42"], tone: "" },
  { active: -1, down: [0], send: "idle", log: ["RPC-1 超时，未当作安全", "RPC-1 timeout, not treated as safe"], tone: "warn" },
  { active: 1, down: [0], send: "idle", log: ["自动切换 RPC-2 · 读取成功", "Switched to RPC-2 · read OK"], tone: "ok" },
  { active: 1, down: [0], send: "fail", log: ["HF 1.18 < 预警线 1.20 · 推送失败", "HF 1.18 < line 1.20 · push failed"], tone: "warn" },
  { active: 1, down: [0], send: "ok", log: ["下一轮补发 · 已送达 Telegram", "Retried next round · delivered"], tone: "ok" },
  { active: 1, down: [], send: "idle", log: ["RPC-1 恢复 · 转为备用", "RPC-1 back · now standby"], tone: "ok" },
];

const STEP_MS = 2200;
const LOG_LINES = 4;

const clock = (tick: number) => {
  const s = 12 * 3600 + tick * 7;
  const p = (n: number) => String(n).padStart(2, "0");
  return `${p(Math.floor(s / 3600) % 24)}:${p(Math.floor(s / 60) % 60)}:${p(s % 60)}`;
};

export default function FailoverDiagram() {
  const [tick, setTick] = useState(0);

  useEffect(() => {
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      setTick(4);
      return;
    }
    const timer = window.setInterval(() => setTick((t) => t + 1), STEP_MS);
    return () => window.clearInterval(timer);
  }, []);

  const phase = PHASES[tick % PHASES.length];
  const logs = Array.from({ length: Math.min(LOG_LINES, tick + 1) }, (_, i) => tick - Math.min(LOG_LINES, tick + 1) + 1 + i);
  const sendLabel = {
    idle: tr("待命", "Idle"),
    fail: tr("重试中", "Retrying"),
    ok: tr("已送达", "Delivered"),
  }[phase.send];

  return (
    <div className="failover frame" aria-label={tr("多节点容灾示意", "Multi-node failover demo")}>
      <div className="gauge-head">
        <span>SYSTEM · FAILOVER</span>
        <span className="gauge-live">
          <i className="signal-dot" /> LIVE
        </span>
      </div>

      <div className="fo-diagram">
        <div className="fo-nodes">
          {NODES.map((name, i) => {
            const down = phase.down.includes(i);
            const state = down ? "down" : phase.active === i ? "active" : "standby";
            return (
              <div className={`fo-row ${state}`} key={name}>
                <span className="fo-node">
                  <i />
                  <b>{name}</b>
                </span>
                <span className="fo-state">{down ? "DOWN" : state === "active" ? "ACTIVE" : "STANDBY"}</span>
                <span className="fo-wire" />
              </div>
            );
          })}
        </div>

        <div className={`fo-target contract ${phase.active >= 0 ? "live" : ""}`}>
          <span className="micro">{tr("链上合约", "Contracts")}</span>
          <b>Aave</b>
          <b>Comp</b>
        </div>
        <span className={`fo-link ${phase.send === "fail" ? "fail" : phase.send === "ok" ? "live" : ""}`} />
        <div className={`fo-target tg ${phase.send}`}>
          <span className="micro">Telegram</span>
          <b>{sendLabel}</b>
        </div>
      </div>

      <ol className="fo-log">
        {logs.map((t) => {
          const p = PHASES[t % PHASES.length];
          return (
            <li key={t} className={p.tone}>
              <time>{clock(t)}</time>
              <span>{tr(...p.log)}</span>
            </li>
          );
        })}
      </ol>

      <div className="gauge-stats">
        <div>
          <span className="micro">{tr("节点", "Nodes")}</span>
          <strong>{NODES.length - phase.down.length} / {NODES.length}</strong>
        </div>
        <div>
          <span className="micro">{tr("当前", "Active")}</span>
          <strong>{phase.active >= 0 ? NODES[phase.active] : tr("切换中", "Switching")}</strong>
        </div>
        <div>
          <span className="micro">{tr("漏报", "Missed")}</span>
          <strong className="fo-zero">0</strong>
        </div>
      </div>
    </div>
  );
}
