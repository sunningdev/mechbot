import { useEffect, useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import QRCode from "qrcode";
import { api, ApiError, type AdminSetup, type AdminStats, type AdminUserRow } from "../api";
import { localDay } from "../format";
import { setTheme, useTheme } from "../theme";

type Phase = "loading" | "setup" | "login" | "ready" | "disabled";

export default function Admin() {
  const navigate = useNavigate();
  const [phase, setPhase] = useState<Phase>("loading");
  const [setup, setSetup] = useState<AdminSetup | null>(null);
  const [qrDataUrl, setQrDataUrl] = useState("");
  const [username, setUsername] = useState("");
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [users, setUsers] = useState<AdminUserRow[]>([]);
  const theme = useTheme();

  useEffect(() => {
    document.title = "Admin · MechBot";
    document.documentElement.lang = "en";
  }, []);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        await api.adminMe();
        if (cancelled) return;
        setPhase("ready");
        await loadData();
        return;
      } catch (e) {
        if (cancelled) return;
        if (e instanceof ApiError && e.status === 404) {
          navigate("/", { replace: true });
          return;
        }
        if (e instanceof ApiError && e.status === 503) {
          setPhase("disabled");
          return;
        }
      }
      try {
        const s = await api.adminSetup();
        if (cancelled) return;
        setSetup(s);
        setUsername(s.username || "sunzhenning");
        if (!s.bound && s.otpauthUrl) {
          const url = await QRCode.toDataURL(s.otpauthUrl, { width: 220, margin: 2 });
          if (!cancelled) setQrDataUrl(url);
          setPhase("setup");
        } else {
          setPhase("login");
        }
      } catch (e) {
        if (cancelled) return;
        if (e instanceof ApiError && e.status === 404) {
          navigate("/", { replace: true });
          return;
        }
        if (e instanceof ApiError && e.status === 503) setPhase("disabled");
        else setPhase("login");
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [navigate]);

  async function loadData() {
    const [s, u] = await Promise.all([api.adminStats(), api.adminUsers()]);
    setStats(s);
    setUsers(u.users ?? []);
  }

  async function onLogin(e: FormEvent) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      await api.adminAuthLogin(username.trim(), code.trim());
      setSetup((prev) => (prev ? { ...prev, bound: true } : prev));
      setQrDataUrl("");
      setPhase("ready");
      setCode("");
      setError("");
      await loadData();
    } catch (err) {
      if (err instanceof ApiError && (err.status === 401 || err.status === 403)) {
        setError("Illegal login");
      } else {
        setError(err instanceof Error ? err.message : "Sign-in failed");
      }
    } finally {
      setBusy(false);
    }
  }

  async function onLogout() {
    await api.adminLogout().catch(() => {});
    setStats(null);
    setUsers([]);
    setCode("");
    setError("");
    setPhase("login");
  }

  const nextTheme = theme === "dark" ? "light" : "dark";
  const themeLabel = theme === "dark" ? "Switch to light mode" : "Switch to dark mode";
  const showAuthForm = phase === "setup" || phase === "login";

  return (
    <div className="admin-page" lang="en">
      <header className="admin-top">
        <Link to="/" className="brand" aria-label="MechBot home">
          <img className="brand-logo" src="/logo.svg" alt="" />
          <span>MechBot</span>
        </Link>
        <div className="admin-top-actions">
          <button
            type="button"
            className="theme-toggle"
            aria-label={themeLabel}
            title={themeLabel}
            onClick={() => setTheme(nextTheme)}
          >
            {theme === "dark" ? (
              <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true">
                <circle cx="12" cy="12" r="4" />
                <path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" />
              </svg>
            ) : (
              <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true">
                <path d="M20 14.5A8 8 0 0 1 9.5 4a8 8 0 1 0 10.5 10.5z" />
              </svg>
            )}
          </button>
          {phase === "ready" && (
            <button type="button" className="btn ghost" onClick={onLogout}>
              Log out
            </button>
          )}
        </div>
      </header>

      <main className="admin-main">
        <h1>Admin</h1>

        {phase === "loading" && <p className="muted">Loading…</p>}
        {phase === "disabled" && (
          <p className="muted">
            Admin is not configured. Set a <strong>valid base32</strong> ADMIN_TOTP_SECRET (run{" "}
            <code>go run ./cmd/admintotp</code>), plus ADMIN_PATH, then restart the server.
          </p>
        )}

        {showAuthForm && (
          <form className="admin-login" onSubmit={onLogin} autoComplete="off">
            {phase === "setup" && (
              <div className="admin-setup">
                <h2>Bind Google Authenticator</h2>
                <p className="muted">
                  Scan this QR with Google Authenticator (or enter the secret manually). Then confirm with a
                  6-digit code. The QR is shown only until the first successful sign-in.
                </p>
                {qrDataUrl ? (
                  <img className="admin-qr" src={qrDataUrl} alt="Google Authenticator QR code" width={220} height={220} />
                ) : (
                  <p className="muted">Generating QR…</p>
                )}
                {setup?.secret && (
                  <label>
                    Manual secret
                    <input type="text" readOnly value={setup.secret} onFocus={(e) => e.target.select()} />
                  </label>
                )}
              </div>
            )}
            {phase === "login" && <p className="muted">Sign in with username and Google Authenticator code.</p>}
            <label>
              Username
              <input
                type="text"
                name="username"
                autoComplete="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                spellCheck={false}
              />
            </label>
            <label>
              Google Authenticator code
              <input
                type="text"
                name="totp"
                inputMode="numeric"
                autoComplete="one-time-code"
                value={code}
                onChange={(e) => setCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
                required
                minLength={6}
                maxLength={6}
                pattern="[0-9]{6}"
                placeholder="6-digit code"
              />
            </label>
            {error && <p className="admin-error">{error}</p>}
            <div className="admin-login-actions">
              <button type="submit" className="btn" disabled={busy || !username.trim() || code.length !== 6}>
                {phase === "setup" ? "Confirm & sign in" : "Sign in"}
              </button>
            </div>
          </form>
        )}

        {phase === "ready" && stats && (
          <>
            <section className="admin-section">
              <h2>Today</h2>
              <div className="admin-kpis">
                <div>
                  <strong>{stats.today.visits}</strong>
                  <span>Homepage visits</span>
                </div>
                <div>
                  <strong>{stats.today.uniqueIps}</strong>
                  <span>Unique IPs</span>
                </div>
                <div>
                  <strong>{users.length}</strong>
                  <span>Users</span>
                </div>
              </div>
            </section>

            <section className="admin-section">
              <h2>Last 30 days</h2>
              <div className="admin-table-wrap">
                <table className="admin-table">
                  <thead>
                    <tr>
                      <th>Day (UTC)</th>
                      <th>Visits</th>
                      <th>Unique IPs</th>
                    </tr>
                  </thead>
                  <tbody>
                    {[...stats.days].reverse().map((d) => (
                      <tr key={d.day}>
                        <td>{d.day}</td>
                        <td>{d.visits}</td>
                        <td>{d.uniqueIps}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </section>

            <section className="admin-section">
              <h2>Users</h2>
              <div className="admin-table-wrap">
                <table className="admin-table">
                  <thead>
                    <tr>
                      <th>User</th>
                      <th>Plan</th>
                      <th>Paid</th>
                      <th>Chains</th>
                      <th>Membership</th>
                      <th>Days left</th>
                      <th>Created</th>
                    </tr>
                  </thead>
                  <tbody>
                    {users.map((u) => (
                      <tr key={u.telegramId}>
                        <td>
                          <div className="admin-user">
                            <span>{u.username ? `@${u.username}` : u.firstName || "—"}</span>
                            <small className="muted">{u.telegramId}</small>
                          </div>
                        </td>
                        <td>
                          <span className={`admin-plan admin-plan-${u.plan}`}>{planLabel(u.plan)}</span>
                        </td>
                        <td>${u.paidUsd.toFixed(2)}</td>
                        <td>{u.chains?.length ? u.chains.join(", ") : "—"}</td>
                        <td>{membershipLabel(u.lastPlan)}</td>
                        <td>{u.plan === "starter" ? "—" : u.daysLeft}</td>
                        <td>{fmtDate(u.createdAt)}</td>
                      </tr>
                    ))}
                    {users.length === 0 && (
                      <tr>
                        <td colSpan={7} className="muted">
                          No users yet.
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </section>
          </>
        )}
      </main>
    </div>
  );
}

function planLabel(plan: string) {
  if (plan === "trial") return "Trial Pro";
  if (plan === "pro") return "Pro";
  return "Starter";
}

function membershipLabel(last?: string) {
  if (last === "month") return "Monthly";
  if (last === "quarter") return "Quarterly";
  if (last === "year") return "Yearly";
  if (last === "custom") return "Custom";
  return "—";
}

function fmtDate(v?: string) {
  if (!v) return "—";
  return localDay(v);
}
