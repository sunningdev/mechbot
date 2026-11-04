import { useCallback, useEffect, useState } from "react";
import { NavLink, Outlet, useNavigate, useOutletContext } from "react-router-dom";
import { api, ApiError, type Me } from "../api";
import Brand from "../components/Brand";
import LangToggle from "../components/LangToggle";
import StatusStrip from "../components/StatusStrip";
import { tr } from "../i18n";

export type ConsoleContext = { me: Me; refresh: () => Promise<void> };

export const useConsole = () => useOutletContext<ConsoleContext>();

const tabs = () => [
  { to: "/app", label: tr("控制台", "Console"), end: true },
  { to: "/app/map", label: tr("清算地图", "Liq. map"), end: false },
  { to: "/app/pro", label: tr("会员", "Membership"), end: false },
  { to: "/app/referral", label: tr("推荐", "Referral"), end: false },
];

export default function ConsoleLayout() {
  const navigate = useNavigate();
  const [me, setMe] = useState<Me | null>(null);

  const refresh = useCallback(async () => {
    try {
      setMe(await api.me());
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) navigate("/", { replace: true });
    }
  }, [navigate]);

  useEffect(() => {
    void refresh();
    const timer = window.setInterval(() => void refresh(), 30_000);
    return () => window.clearInterval(timer);
  }, [refresh]);

  async function logout() {
    await api.logout().catch(() => undefined);
    navigate("/", { replace: true });
  }

  if (!me) {
    return <div className="page center-screen muted">{tr("加载中…", "Loading…")}</div>;
  }

  const name = me.username ? `@${me.username}` : me.firstName || String(me.telegramId);

  return (
    <div className="page">
      <StatusStrip />
      <div className="topbar">
        <header className="nav container">
          <Brand />
          <nav className="nav-links">
            <div className="console-tabs">
              {tabs().map((t) => (
                <NavLink key={t.to} to={t.to} end={t.end} className={({ isActive }) => `console-tab ${isActive ? "on" : ""}`}>
                  {t.label}
                </NavLink>
              ))}
            </div>
            <span className="muted console-name">{name}</span>
            <LangToggle />
            <button className="btn btn-small btn-ghost" onClick={() => void logout()}>
              {tr("退出", "Log out")}
            </button>
          </nav>
        </header>
      </div>
      <Outlet context={{ me, refresh } satisfies ConsoleContext} />
    </div>
  );
}
