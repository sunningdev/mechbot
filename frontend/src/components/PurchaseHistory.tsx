import { useEffect, useState } from "react";
import { api, ApiError, type Deposit, type DepositPage } from "../api";
import { dateTime } from "../format";
import { planName, tr } from "../i18n";
import ChainLogo from "./ChainLogo";

type Props = { reload: number };

const planByDays: Record<number, string> = { 30: "month", 90: "quarter", 365: "year" };

function PlanCell({ d }: { d: Deposit }) {
  if (d.status !== "credited") return <span className="status-unmatched">{tr("金额不足，未开通", "Amount too low, not activated")}</span>;
  const key = planByDays[d.days];
  return (
    <>
      {key ? `Pro ${planName(key)}` : "Pro"} <span className="status-credited">{tr(`+${d.days} 天`, `+${d.days} days`)}</span>
    </>
  );
}

export default function PurchaseHistory({ reload }: Props) {
  const [page, setPage] = useState(1);
  const [data, setData] = useState<DepositPage | null>(null);
  const [error, setError] = useState("");

  useEffect(() => setPage(1), [reload]);

  useEffect(() => {
    let live = true;
    api.billingHistory(page).then(
      (d) => {
        if (!live) return;
        setData(d);
        setError("");
      },
      (err) => live && setError(err instanceof ApiError ? err.message : tr("读取失败", "Could not load")),
    );
    return () => {
      live = false;
    };
  }, [page, reload]);

  const pages = data ? Math.max(1, Math.ceil(data.total / data.pageSize)) : 1;

  return (
    <section className="panel">
      <div className="panel-head">
        <h2>{tr("购买记录", "Purchase history")}</h2>
        {data && data.total > 0 && <span className="muted small-text">{tr(`共 ${data.total} 条`, `${data.total} total`)}</span>}
      </div>

      {error && <p className="form-error">{error}</p>}
      {data && data.total === 0 && <p className="muted">{tr("还没有购买记录。", "No purchases yet.")}</p>}

      {data && data.items.length > 0 && (
        <div className="history-scroll">
          <table className="history">
            <thead>
              <tr>
                <th>{tr("时间", "Time")}</th>
                <th>{tr("方案", "Plan")}</th>
                <th>{tr("金额", "Amount")}</th>
                <th>{tr("网络", "Network")}</th>
                <th>{tr("交易", "Transaction")}</th>
              </tr>
            </thead>
            <tbody>
              {data.items.map((d) => {
                const short = `${d.txHash.slice(0, 8)}…${d.txHash.slice(-6)}`;
                return (
                  <tr key={`${d.chain}-${d.txHash}`}>
                    <td className="mono">{dateTime(d.createdAt)}</td>
                    <td>
                      <PlanCell d={d} />
                    </td>
                    <td className="mono">{d.amount} USDT</td>
                    <td>
                      <span className="history-chain">
                        <ChainLogo chain={d.chain} size={16} />
                        {d.chainName ?? d.chain}
                      </span>
                    </td>
                    <td className="mono">
                      {d.txUrl ? (
                        <a href={d.txUrl} target="_blank" rel="noreferrer">
                          {short} ↗
                        </a>
                      ) : (
                        short
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {pages > 1 && (
        <nav className="pager" aria-label={tr("分页", "Pagination")}>
          <button className="btn btn-small btn-ghost" disabled={page <= 1} onClick={() => setPage(page - 1)}>
            ← {tr("上一页", "Previous")}
          </button>
          <span className="mono">
            {page} / {pages}
          </span>
          <button className="btn btn-small btn-ghost" disabled={page >= pages} onClick={() => setPage(page + 1)}>
            {tr("下一页", "Next")} →
          </button>
        </nav>
      )}
    </section>
  );
}
