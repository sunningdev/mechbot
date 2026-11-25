declare global {
  interface Window {
    __MECHBOT_ADMIN_PATH__?: string;
  }
}

const blocked = new Set(["admin", "api", "app", "blog", "terms", "privacy", "assets", "static", "favicon.ico", "robots.txt"]);

function sanitize(raw: string): string {
  const p = raw.trim().replace(/^\/+|\/+$/g, "");
  if (!p || p.includes("/") || p.length < 8 || p.length > 64) return "";
  if (!/^[a-zA-Z0-9_-]+$/.test(p)) return "";
  if (blocked.has(p.toLowerCase())) return "";
  return p;
}

/**
 * Admin URL slug. Sources (first match wins):
 * 1. Injected by Go when it serves the admin page HTML
 * 2. VITE_ADMIN_PATH for local `npm run dev`
 * 3. Current path — needed when Nginx serves the SPA and Go never injects the slug
 */
export function adminPath(): string {
  const fromWindow = typeof window !== "undefined" ? window.__MECHBOT_ADMIN_PATH__ : undefined;
  if (fromWindow) {
    const p = sanitize(fromWindow);
    if (p) return p;
  }
  const fromEnv = import.meta.env.VITE_ADMIN_PATH;
  if (fromEnv) {
    const p = sanitize(String(fromEnv));
    if (p) return p;
  }
  if (typeof window !== "undefined") {
    const slug = window.location.pathname.replace(/^\/+|\/+$/g, "").split("/")[0] ?? "";
    return sanitize(slug);
  }
  return "";
}

export function adminPagePath(): string {
  const p = adminPath();
  return p ? `/${p}` : "";
}

export function adminApiBase(): string {
  const p = adminPath();
  return p ? `/api/${p}` : "";
}

export function isAdminPathname(pathname: string): boolean {
  const p = adminPagePath();
  return Boolean(p && (pathname === p || pathname.startsWith(p + "/")));
}
