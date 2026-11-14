import { useSyncExternalStore } from "react";

export type Theme = "dark" | "light";

const KEY = "mechbot-theme";
const media = typeof window !== "undefined" ? window.matchMedia("(prefers-color-scheme: light)") : null;
const listeners = new Set<() => void>();

function stored(): Theme | null {
  try {
    const v = localStorage.getItem(KEY);
    return v === "dark" || v === "light" ? v : null;
  } catch {
    return null;
  }
}

export function getTheme(): Theme {
  const attr = document.documentElement.dataset.theme;
  if (attr === "dark" || attr === "light") return attr;
  return stored() ?? (media?.matches ? "light" : "dark");
}

function apply(theme: Theme) {
  document.documentElement.dataset.theme = theme;
  const meta = document.querySelector("meta[name=theme-color]");
  meta?.setAttribute("content", theme === "light" ? "#f5f7fa" : "#18212c");
  listeners.forEach((fn) => fn());
}

export function setTheme(theme: Theme) {
  try {
    localStorage.setItem(KEY, theme);
  } catch {
    /* storage unavailable */
  }
  apply(theme);
}

media?.addEventListener("change", (e) => {
  if (!stored()) apply(e.matches ? "light" : "dark");
});

function subscribe(fn: () => void) {
  listeners.add(fn);
  return () => listeners.delete(fn);
}

export function useTheme(): Theme {
  return useSyncExternalStore(subscribe, getTheme, () => "dark");
}

export function themeAccent(theme: Theme) {
  return theme === "light" ? "#1d5ff5" : "#4c82fb";
}
