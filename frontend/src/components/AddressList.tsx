import { useState, type FormEvent } from "react";
import { api, ApiError, type Me, type WatchedAddress } from "../api";
import { riskClass, shortAddress, timeAgo, usd } from "../format";
import { tr, translateServer } from "../i18n";
import CopyButton from "./CopyButton";
import MiniGauge from "./MiniGauge";

function riskiest(a: WatchedAddress) {
  return a.positions
    .filter((p) => !p.error && p.debtUsd > 0 && Number.isFinite(p.healthFactor))
    .reduce<WatchedAddress["positions"][number] | undefined>(
      (low, p) => (!low || p.healthFactor < low.healthFactor ? p : low),
      undefined,
    );
}

function AddressBody({ a, threshold }: { a: WatchedAddress; threshold: number }) {
  const top = riskiest(a);
  const rest = a.positions.filter((p) => p !== top);

  return (
    <div className="address-body">
      {top && <MiniGauge position={top} threshold={threshold} />}
      {rest.length > 0 && (
        <div className="positions">
          {rest.map((p) =>
            p.error ? (
              <div className="position" key={p.key}>
                <span className="position-label">{p.label}</span>
                <span className="muted small-text">
                  {translateServer(p.error)}
                  {tr("，下轮重试", ", retrying next round")}
                </span>
              </div>
            ) : (
              <div className="position" key={p.key}>
                <span className="position-main">
                  <span className="position-label">
                    {p.label}
                    {!p.covered && <span className="tag">{tr("Pro 预警", "Pro alerts")}</span>}
                  </span>
                  <span className="position-nums muted">
                    {tr("抵押", "Collateral")} {usd(p.collateralUsd)} · {tr("债务", "Debt")} {usd(p.debtUsd)}
                    {p.liqEthPrice ? <> · {tr("ETH 清算价", "ETH liq. price")} {usd(p.liqEthPrice)}</> : null}
                  </span>
                </span>
                <span className={`hf ${riskClass(p.healthFactor, threshold)}`}>{p.healthFactor.toFixed(3)}</span>
              </div>
            ),
          )}
        </div>
      )}
    </div>
  );
}

function ThresholdControl({ a, onSaved }: { a: WatchedAddress; onSaved: () => void }) {
  const [editing, setEditing] = useState(false);
  const [input, setInput] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const hint = tr(
    "默认 1.2，健康因子低于这个数就发预警，之后每 1 小时提醒一次；低于 1.05 时每 15 分钟提醒一次。",
    "1.2 by default. You're alerted when the health factor drops below it, then reminded every hour, or every 15 minutes once it's under 1.05.",
  );

  function open() {
    setInput(String(a.threshold));
    setError("");
    setEditing(true);
  }

  async function save(e: FormEvent) {
    e.preventDefault();
    const n = Number(input);
    if (!(n > 1 && n <= 10)) {
      setError(tr("预警线要大于 1，且不超过 10", "The alert line must be above 1 and at most 10"));
      return;
    }
    setBusy(true);
    setError("");
    try {
      await api.setAddressThreshold(a.address, n);
      setEditing(false);
      onSaved();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : tr("保存失败", "Could not save"));
    } finally {
      setBusy(false);
    }
  }

  if (!editing) {
    return (
      <span className="threshold">
        {tr("预警线：", "Alert line:")}
        <button type="button" className="threshold-value" onClick={open} title={hint}>
          {a.threshold}
        </button>
      </span>
    );
  }

  return (
    <form className="threshold editing" onSubmit={save} title={error || hint}>
      {tr("预警线：", "Alert line:")}
      <input
        type="number"
        step="0.05"
        min="1.05"
        max="10"
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={(e) => e.key === "Escape" && setEditing(false)}
        className={error ? "invalid" : ""}
        disabled={busy}
        autoFocus
      />
      <button className="threshold-link strong" disabled={busy}>
        {tr("保存", "Save")}
      </button>
      <button type="button" className="threshold-link" disabled={busy} onClick={() => setEditing(false)}>
        {tr("取消", "Cancel")}
      </button>
      {error && <span className="threshold-error">{error}</span>}
    </form>
  );
}

type Props = {
  me: Me;
  addresses: WatchedAddress[];
  onChange: () => void;
};

export default function AddressList({ me, addresses, onChange }: Props) {
  const [input, setInput] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const full = me.addressCount >= me.addressLimit;

  async function add(e: FormEvent) {
    e.preventDefault();
    if (!input.trim()) return;
    setBusy(true);
    setError("");
    try {
      await api.addAddress(input.trim());
      setInput("");
      onChange();
      window.setTimeout(onChange, 6000);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : tr("添加失败", "Could not add the address"));
    } finally {
      setBusy(false);
    }
  }

  async function remove(address: string) {
    const short = shortAddress(address);
    if (!window.confirm(tr(`删除 ${short}？删除后不再预警。`, `Delete ${short}? It will no longer be alerted.`))) return;
    try {
      await api.removeAddress(address);
      onChange();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : tr("删除失败", "Could not delete the address"));
    }
  }

  return (
    <section className="panel">
      <div className="panel-head">
        <h2>{tr("监控地址", "Watched addresses")}</h2>
        <span className="muted">
          {me.addressCount} / {me.addressLimit}
        </span>
      </div>

      <form className="add-row" onSubmit={add}>
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder={full ? tr("已达到地址上限", "Address limit reached") : tr("粘贴 0x 开头的钱包地址", "Paste a wallet address starting with 0x")}
          disabled={full || busy}
          spellCheck={false}
        />
        <button className="btn btn-primary" disabled={full || busy || !input.trim()}>
          {busy ? tr("添加中…", "Adding…") : tr("添加", "Add")}
        </button>
      </form>
      {error && <p className="form-error">{error}</p>}
      {full && !me.pro && (
        <p className="muted small-text">
          {tr(
            `Starter 最多 ${me.addressLimit} 个地址，升级 Pro 可监控 20 个。`,
            `Starter allows ${me.addressLimit} addresses. Upgrade to Pro to watch 20.`,
          )}
        </p>
      )}

      {addresses.length === 0 ? (
        <div className="empty">
          {tr("还没有地址。添加后几秒内会显示各协议的健康因子。", "No addresses yet. Health factors appear within seconds of adding one.")}
        </div>
      ) : (
        <div className="address-list">
          {addresses.map((a) => (
            <div className={`address-card ${a.active ? "" : "paused"}`} key={a.address}>
              <div className="address-top">
                <div className="address-id">
                  <span className="mono" title={a.address}>
                    {shortAddress(a.address)}
                  </span>
                  <CopyButton text={a.address} />
                  {!a.active && <span className="tag">{tr("已暂停 · 超出 Starter 额度", "Paused · over Starter limit")}</span>}
                </div>
                <div className="address-actions">
                  <ThresholdControl key={`${a.address}:${a.threshold}`} a={a} onSaved={onChange} />
                  <span className="muted small-text">{timeAgo(a.checkedAt)}</span>
                  <button className="btn btn-small btn-ghost danger-text" onClick={() => void remove(a.address)}>
                    {tr("删除", "Delete")}
                  </button>
                </div>
              </div>

              {a.checkedAt && a.positions.length === 0 && (
                <div className="no-debt">
                  <b>{tr("这个地址目前没有借款", "This address has no debt right now")}</b>
                  <p>
                    {tr(
                      "已检查 Aave V3、Compound V3、Morpho Blue（Ethereum · Arbitrum · Optimism），都没有借款，不会触发清算。",
                      "Checked Aave V3, Compound V3 and Morpho Blue (Ethereum · Arbitrum · Optimism). No debt anywhere, so nothing can be liquidated.",
                    )}
                  </p>
                  <p>
                    {tr(
                      "地址会继续监控，以后在这些协议借款，会自动开始预警，不用重新添加。",
                      "We keep watching it. Borrow on any of these protocols later and alerts start automatically, no need to add it again.",
                    )}
                  </p>
                  {!me.pro && (
                    <p className="no-debt-note">
                      {tr(
                        "Starter 只对 Aave V3 发预警，Compound、Morpho 的借款需要 Pro。",
                        "Starter alerts on Aave V3 only. Compound and Morpho loans need Pro.",
                      )}
                    </p>
                  )}
                </div>
              )}

              {a.positions.length > 0 && <AddressBody a={a} threshold={a.threshold} />}
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
