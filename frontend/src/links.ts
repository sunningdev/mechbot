// Outbound links that come from the API are only rendered when they point at a known https host.
export const APP_HOSTS = ["app.aave.com", "app.morpho.org", "app.compound.finance"];
export const EXPLORER_HOSTS = ["etherscan.io", "arbiscan.io", "basescan.org", "bscscan.com", "tronscan.org"];
export const TELEGRAM_HOSTS = ["t.me"];

export function safeUrl(raw: string | undefined, hosts: string[]): string | undefined {
  if (!raw) return undefined;
  try {
    const u = new URL(raw);
    if (u.protocol !== "https:") return undefined;
    const host = u.hostname.toLowerCase();
    return hosts.some((h) => host === h || host.endsWith("." + h)) ? u.href : undefined;
  } catch {
    return undefined;
  }
}
