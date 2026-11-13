import { Fragment, lazy, Suspense, useCallback, useEffect, useRef, useState, type ReactNode } from "react";
import QRCode from "qrcode";
import { api, ApiError, type Billing, type Chain, type Me, type Plan } from "../api";
import { dateTime } from "../format";
import { planName, tr } from "../i18n";
import { EXPLORER_HOSTS, safeUrl } from "../links";
import { coinOrder, sortChains, tokenLabel, usdtFirst } from "../payorder";
import ChainLogo from "./ChainLogo";
import ChainSelect from "./ChainSelect";
import CopyButton from "./CopyButton";

// The wallet libraries are large, so they load only when the user chooses to pay from a wallet.
const WalletPay = lazy(() => import("./WalletPay"));

type Mode = "wallet" | "address";

// After a wallet transfer, the chain is re-scanned this often until the deposit is credited.
const SENT_POLL_MS = 25_000;
const SENT_POLLS = 12;

type Props = { me: Me; onPaid: () => void };

// Replaces {0}, {1}… in a translated sentence with rich nodes.
const fill = (text: string, ...values: ReactNode[]) =>
  text.split(/(\{\d\})/).map((part, i) => {
    const slot = part.match(/^\{(\d)\}$/);
    return <Fragment key={i}>{slot ? values[Number(slot[1])] : part}</Fragment>;
  });

const credited = (b: Billing) => b.deposits.filter((d) => d.status === "credited").length;

// Saving versus buying the shortest plan month by month: 90 days = 3 months, 365 days = 12 months.
function saving(plans: Plan[], p: Plan) {
  const base = plans.reduce<Plan | null>((min, x) => (!min || x.days < min.days ? x : min), null);
  if (!base || p.days <= base.days) return 0;
  const full = Math.round(p.days / base.days) * base.price;
  return Math.max(0, Math.round(((full - p.price) / full) * 100));
}

// The first network in the user's order that also takes the monthly plan, so TRON never hides it by default.
function defaultChain(chains: Chain[]) {
  const lowest = Math.min(...chains.map((c) => c.minPrice));
  return sortChains(chains).find((c) => c.minPrice === lowest)?.key ?? "";
}

export default function BillingPanel({ me, onPaid }: Props) {
  const [billing, setBilling] = useState<Billing | null>(null);
  const [chainKey, setChainKey] = useState("");
  const [qr, setQr] = useState("");
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);
  const [busy, setBusy] = useState(false);
  const [mode, setMode] = useState<Mode>(usdtFirst() ? "address" : "wallet");
  const [planKey, setPlanKey] = useState("");
  const [walletStarted, setWalletStarted] = useState(false);
  const [projectId, setProjectId] = useState("");
  const [sent, setSent] = useState<{ hash: string; chain: Chain } | null>(null);
  const paidCount = useRef<number | null>(null);
  const onPaidRef = useRef(onPaid);
  onPaidRef.current = onPaid;

  const apply = useCallback((b: Billing) => {
    setBilling(b);
    setChainKey((cur) => (b.chains.some((c) => c.key === cur) ? cur : defaultChain(b.chains)));
    const n = credited(b);
    if (paidCount.current !== null && n > paidCount.current) {
      setMsg({ ok: true, text: tr("已到账，Pro 已开通。", "Payment received. Pro is active.") });
      setSent(null);
      onPaidRef.current();
    }
    paidCount.current = n;
  }, []);

  useEffect(() => {
    const load = () => api.billing().then(apply, () => undefined);
    void load();
    const timer = window.setInterval(load, 30_000);
    return () => window.clearInterval(timer);
  }, [apply]);

  useEffect(() => {
    api.config().then((c) => setProjectId(c.walletConnectId ?? ""), () => undefined);
  }, []);

  useEffect(() => {
    if (!sent) return;
    let polls = 0;
    const timer = window.setInterval(() => {
      if (++polls > SENT_POLLS) window.clearInterval(timer);
      api.checkDeposit().then(apply, () => undefined);
    }, SENT_POLL_MS);
    return () => window.clearInterval(timer);
  }, [sent, apply]);

  const chain = billing?.chains.find((c) => c.key === chainKey);
  const address = chain ? (chain.family === "tron" ? billing?.address?.tron : billing?.address?.evm) : "";

  useEffect(() => {
    if (!address) {
      setQr("");
      return;
    }
    QRCode.toDataURL(address, { margin: 1, width: 264, color: { dark: "#111111", light: "#f6f5f1" } }).then(
      setQr,
      () => setQr(""),
    );
  }, [address]);

  async function check() {
    setBusy(true);
    setMsg(null);
    try {
      const before = paidCount.current;
      const b = await api.checkDeposit();
      apply(b);
      if (before !== null && credited(b) === before) {
        setMsg({
          ok: false,
          text: tr(
            "还没查到到账。链上确认需要 1–3 分钟，稍后再点一次。",
            "No payment found yet. On-chain confirmation takes 1–3 minutes; try again shortly.",
          ),
        });
      }
    } catch (err) {
      setMsg({ ok: false, text: err instanceof ApiError ? err.message : tr("检查失败", "Check failed") });
    } finally {
      setBusy(false);
    }
  }

  if (!billing) return null;
  const minPrice = billing.plans[0]?.price ?? 0;
  const walletReady = Boolean(billing.address?.evm) && billing.chains.some((c) => c.chainId && c.coins?.length);
  const wallet = mode === "wallet" && walletReady;
  const plan = billing.plans.find((p) => p.key === planKey) ?? billing.plans[0];

  return (
    <section className="panel">
      <div className="panel-head">
        <h2>{me.pro ? tr("Pro 续费", "Renew Pro") : tr("升级 Pro", "Upgrade to Pro")}</h2>
        {me.pro && (
          <span className="muted small-text">
            {tr("到期", "Expires")} {dateTime(me.proExpireAt)}
          </span>
        )}
      </div>

      {!billing.enabled && <p className="muted">{tr("USDT 充值暂未开放。", "USDT payments are not open yet.")}</p>}

      {billing.enabled && billing.testnet && (
        <p className="testnet-note">
          <b>{tr("测试网模式", "Testnet mode")}</b>
          {tr(
            "收款只认测试网的测试币，不要转真 USDT，转了无法找回。",
            "Payments only accept testnet tokens. Do not send real USDT; it cannot be recovered.",
          )}
        </p>
      )}

      {billing.enabled && (
        <div className="pay-box">
          <div className="tier-row">
            {billing.plans.filter((p) => wallet || p.price >= (chain?.minPrice ?? 0)).map((p) => {
              const save = saving(billing.plans, p);
              const body = (
                <>
                  <strong>${p.price}</strong>
                  <span className="tier-days">
                    <span className="muted small-text">{tr(`${p.days} 天`, `${p.days} days`)}</span>
                    <span className="chain-tag">{planName(p.key, p.name.replace(/^Pro\s*/, ""))}</span>
                  </span>
                  {save > 0 && <span className="tier-save">{tr(`省 ${save}%`, `Save ${save}%`)}</span>}
                </>
              );
              return wallet ? (
                <button
                  key={p.key}
                  type="button"
                  className={`tier tier-pick ${p.key === plan?.key ? "on" : ""}`}
                  aria-pressed={p.key === plan?.key}
                  onClick={() => setPlanKey(p.key)}
                >
                  {body}
                </button>
              ) : (
                <div key={p.key} className="tier">
                  {body}
                </div>
              );
            })}
          </div>

          {wallet && plan && billing.address?.evm && (
            <>
              {walletStarted ? (
                <Suspense fallback={<p className="muted small-text">{tr("正在加载钱包…", "Loading wallet…")}</p>}>
                  <WalletPay
                    chains={billing.chains}
                    deposit={billing.address.evm}
                    price={plan.price}
                    projectId={projectId}
                    onSent={(hash, c) => {
                      setMsg(null);
                      setSent({ hash, chain: c });
                    }}
                  />
                </Suspense>
              ) : (
                <button className="btn btn-primary btn-block" onClick={() => setWalletStarted(true)}>
                  {tr("连接钱包支付", "Connect wallet to pay")}
                </button>
              )}
              {sent && (
                <p className="form-ok small-text">
                  {tr("已发送，等待链上确认，通常 1–3 分钟。", "Sent. Waiting for on-chain confirmation, usually 1–3 minutes.")}{" "}
                  {safeUrl(sent.chain.explorer && sent.chain.explorer + sent.hash, EXPLORER_HOSTS) && (
                    <a href={safeUrl(sent.chain.explorer + sent.hash, EXPLORER_HOSTS)} target="_blank" rel="noreferrer">
                      {tr("查看交易", "View transaction")}
                    </a>
                  )}
                </p>
              )}
              <button type="button" className="threshold-link pay-switch" onClick={() => setMode("address")}>
                {tr("或 转账到专属地址 →", "Or send to your deposit address →")}
              </button>
            </>
          )}

          {!wallet && (
          <>
          <p className="pay-how small-text">
            {fill(
              tr(
                "用 {0} 或 {1} 付款：选你钱包里有币、手续费低的网络，再转到下面的专属地址。",
                "Pay with {0} or {1}: pick a network where your wallet holds one and fees are low, then send to your address below.",
              ),
              ...coinOrder().map((c) => (
                <b key={c} className="coin-hl">
                  {c}
                </b>
              )),
            )}
          </p>
          <ChainSelect chains={sortChains(billing.chains)} value={chainKey} onChange={setChainKey} />

          {chain && address && (
            <div className="deposit">
              {qr && (
                <img className="deposit-qr" src={qr} alt={tr(`${chain.name} 充值地址二维码`, `${chain.name} deposit address QR code`)} />
              )}
              <div className="deposit-info">
                <span className="deposit-chain">
                  <ChainLogo chain={chain.key} size={18} />
                  <span className="muted small-text">
                    {tr(`${chain.name} · 你的专属 ${tokenLabel(chain.token)} 充值地址`, `${chain.name} · your personal ${tokenLabel(chain.token)} deposit address`)}
                  </span>
                </span>
                <span className="mono break deposit-addr">{address}</span>
                <CopyButton text={address} />
              </div>
            </div>
          )}

          {chain && (
            <ul className="deposit-notes muted small-text">
              <li>
                {fill(
                  tr(
                    "只转 {0}，网络选 {1}，转错币种或网络无法找回。",
                    "Send only {0} on the {1} network. Wrong tokens or networks cannot be recovered.",
                  ),
                  <b>{tokenLabel(chain.token)}</b>,
                  <b>{chain.name}</b>,
                )}
              </li>
              {chain.minPrice > minPrice ? (
                <li>
                  {fill(
                    tr("{0} 最低季付 {1}；月付请选 {2}。", "{0} starts at quarterly, {1}. For monthly, use {2}."),
                    chain.name,
                    <b>${chain.minPrice}</b>,
                    sortChains(billing.chains)
                      .filter((c) => c.minPrice <= minPrice)
                      .map((c) => c.name)
                      .join(tr("、", ", ")),
                  )}
                </li>
              ) : (
                <li>
                  {tr(
                    `到账金额对应套餐自动开通，超出部分不累计；低于 $${minPrice} 不开通。`,
                    `The plan matching the amount received activates automatically; any excess is not carried over. Below $${minPrice} nothing activates.`,
                  )}
                </li>
              )}
              <li>
                {tr(
                  "交易所提币会扣手续费，请确保到账金额足额。地址长期有效，续费转同一个地址即可。",
                  "Exchange withdrawals deduct fees, so make sure the full amount arrives. This address is permanent; renew by sending to it again.",
                )}
              </li>
            </ul>
          )}

          <button className="btn btn-ghost btn-block" onClick={() => void check()} disabled={busy}>
            {busy ? tr("检查中…", "Checking…") : tr("我已转账，立即检查", "I've paid, check now")}
          </button>
          {walletReady && (
            <button type="button" className="threshold-link pay-switch" onClick={() => setMode("wallet")}>
              {tr("或 连接钱包支付 →", "Or pay from a connected wallet →")}
            </button>
          )}
          </>
          )}
        </div>
      )}

      {msg && <p className={msg.ok ? "form-ok" : "form-error"}>{msg.text}</p>}
    </section>
  );
}
