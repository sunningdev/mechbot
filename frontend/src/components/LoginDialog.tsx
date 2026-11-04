import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api, storedRef, type LoginStart } from "../api";
import { tr } from "../i18n";

type Props = { onClose: () => void };

const svg = {
  width: 22,
  height: 22,
  viewBox: "0 0 24 24",
  fill: "none",
  stroke: "currentColor",
  strokeWidth: 2,
  strokeLinecap: "round" as const,
  strokeLinejoin: "round" as const,
  "aria-hidden": true,
};

const PlaneIcon = () => (
  <svg {...svg} width={26} height={26}>
    <path d="M22 2 11 13" />
    <path d="M22 2 15 22l-4-9-9-4 20-7z" />
  </svg>
);

const ShieldIcon = () => (
  <svg {...svg}>
    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
  </svg>
);

const BoltIcon = () => (
  <svg {...svg}>
    <path d="M13 2 3 14h9l-1 8 10-12h-9l1-8z" />
  </svg>
);

const CheckIcon = () => (
  <svg {...svg}>
    <circle cx="12" cy="12" r="10" />
    <path d="m8 12 3 3 5-6" />
  </svg>
);

const features = () => [
  { icon: <ShieldIcon />, text: tr("机器人里确认，安全登录", "Confirm in the bot for a secure login") },
  { icon: <BoltIcon />, text: tr("即时访问，无需注册和邮箱", "Instant access, no sign-up or email") },
  { icon: <CheckIcon />, text: tr("无需记住密码", "No password to remember") },
];

export default function LoginDialog({ onClose }: Props) {
  const navigate = useNavigate();
  const [login, setLogin] = useState<LoginStart | null>(null);
  const [state, setState] = useState<"intro" | "loading" | "waiting" | "expired" | "error">("intro");

  const start = useCallback(async () => {
    setState("loading");
    try {
      setLogin(await api.loginStart(storedRef()));
      setState("waiting");
    } catch {
      setState("error");
    }
  }, []);

  useEffect(() => {
    if (state !== "waiting" || !login) return;
    const timer = window.setInterval(async () => {
      try {
        const res = await api.loginPoll(login.token);
        if (res.status === "ok") {
          window.clearInterval(timer);
          navigate("/app");
        } else if (res.status === "expired") {
          window.clearInterval(timer);
          setState("expired");
        }
      } catch {
        /* keep polling through brief network errors */
      }
    }, 2000);
    return () => window.clearInterval(timer);
  }, [state, login, navigate]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && onClose();
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  return (
    <div className="overlay" onClick={onClose}>
      <div className="dialog" role="dialog" aria-modal="true" onClick={(e) => e.stopPropagation()}>
        <button className="dialog-close" onClick={onClose} aria-label={tr("关闭", "Close")}>
          ×
        </button>
        <h2 className="dialog-title">
          <PlaneIcon />
          {tr("使用 Telegram 登录", "Log in with Telegram")}
        </h2>

        {state === "intro" && (
          <>
            <p className="muted">{tr("通过 Telegram 账户快速、安全地登录", "Fast, secure login with your Telegram account")}</p>
            <ul className="feature-list">
              {features().map((f) => (
                <li key={f.text}>
                  {f.icon}
                  {f.text}
                </li>
              ))}
            </ul>
            <button className="btn btn-primary btn-block" onClick={() => void start()}>
              {tr("使用 Telegram 继续", "Continue with Telegram")} <span aria-hidden="true">→</span>
            </button>
          </>
        )}

        {state === "loading" && <p className="muted">{tr("正在生成登录链接…", "Creating login link…")}</p>}

        {state === "waiting" && login && (
          <>
            <p className="muted">
              {tr("在 Telegram 里确认，验证码要和这里一致。", "Confirm in Telegram. The code there must match this one.")}
            </p>
            <div className="code">{login.code}</div>
            <a className="btn btn-primary btn-block" href={login.botUrl} target="_blank" rel="noreferrer">
              {tr("打开 Telegram 确认", "Open Telegram to confirm")}
            </a>
            <p className="hint">
              <span className="pulse" /> {tr("等待确认，确认后自动进入控制台", "Waiting for confirmation, then you'll enter the console")}
            </p>
          </>
        )}

        {(state === "expired" || state === "error") && (
          <>
            <p className="muted">
              {state === "expired"
                ? tr("登录链接已过期。", "The login link has expired.")
                : tr("暂时连不上服务器。", "Can't reach the server right now.")}
            </p>
            <button className="btn btn-primary btn-block" onClick={() => void start()}>
              {tr("重新生成", "Try again")}
            </button>
          </>
        )}
      </div>
    </div>
  );
}
