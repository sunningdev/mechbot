import { translateServer } from "./i18n";

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message);
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    credentials: "same-origin",
    ...init,
    headers: init?.body ? { "Content-Type": "application/json" } : undefined,
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new ApiError(translateServer((data as { error?: string }).error ?? "请求失败"), res.status);
  }
  return data as T;
}

const send = <T,>(method: string, path: string, body: unknown = {}) =>
  request<T>(path, { method, body: JSON.stringify(body) });

export type Plan = { key: string; name: string; price: number; days: number };
export type Coin = { symbol: string; address: string; decimals: number };
export type Chain = {
  key: string;
  name: string;
  token: string;
  family: "evm" | "tron";
  minPrice: number;
  explorer?: string;
  chainId?: number;
  coins?: Coin[];
};

export type SiteConfig = {
  botUsername: string;
  botUrl: string;
  plans: Plan[];
  freeLimit: number;
  proLimit: number;
  walletConnectId?: string;
};

// Raw on-chain balances of a connected wallet, as decimal strings keyed by token contract.
export type WalletBalance = { chain: string; native: string; coins: Record<string, string>; error?: boolean };

export type LoginStart = { token: string; code: string; botUrl: string; expiresIn: number };

export type Me = {
  telegramId: number;
  username: string;
  firstName: string;
  pro: boolean;
  trial?: boolean;
  proExpireAt: string;
  proExpiredAt?: string;
  addressCount: number;
  addressLimit: number;
};

export type Position = {
  key: string;
  label: string;
  healthFactor: number;
  collateralUsd: number;
  debtUsd: number;
  error?: string;
  covered: boolean;
  ethPrice?: number;
  liqEthPrice?: number;
  borrowApy?: number;
  appUrl?: string;
  rescue?: Rescue;
  stress?: number[]; // health factor after ETH falls 10%, 20%, 30%
  wall?: { price: number; total: number; safeHf: number };
};

export type Rescue = {
  target: number;
  repayUsd: number;
  repayAmount?: number;
  repaySymbol?: string;
  addUsd?: number;
  addAmount?: number;
  addSymbol?: string;
};

export type WatchedAddress = {
  address: string;
  active: boolean;
  checkedAt?: string;
  positions: Position[];
  threshold: number;
};

export type Deposit = {
  chain: string;
  txHash: string;
  amount: string;
  token?: string;
  days: number;
  status: "credited" | "unmatched";
  createdAt: string;
  txUrl?: string;
  chainName?: string;
};

export type Billing = {
  enabled: boolean;
  testnet: boolean;
  plans: Plan[];
  chains: Chain[];
  address?: { evm: string; tron: string };
  deposits: Deposit[];
};

export type DepositPage = {
  items: Deposit[];
  total: number;
  page: number;
  pageSize: number;
};

export type Referral = {
  webLink: string;
  botLink: string;
  invited: number;
  paid: number;
  daysEarned: number;
};

export const api = {
  config: () => request<SiteConfig>("/api/config"),
  loginStart: (ref: string) => send<LoginStart>("POST", "/api/auth/start", { ref }),
  loginPoll: (token: string) =>
    request<{ status: "pending" | "ok" | "expired" }>(`/api/auth/poll?token=${encodeURIComponent(token)}`),
  logout: () => send<{ ok: boolean }>("POST", "/api/auth/logout"),
  me: () => request<Me>("/api/me"),
  addresses: () => request<WatchedAddress[]>("/api/addresses"),
  addAddress: (address: string) => send<{ address: string }>("POST", "/api/addresses", { address }),
  removeAddress: (address: string) =>
    send<{ ok: boolean }>("DELETE", `/api/addresses/${encodeURIComponent(address)}`),
  setAddressThreshold: (address: string, threshold: number) =>
    send<{ threshold: number }>("PUT", `/api/addresses/${encodeURIComponent(address)}/threshold`, { threshold }),
  setLang: (lang: string) => send<{ lang: string }>("PUT", "/api/me/lang", { lang }),
  billing: () => request<Billing>("/api/billing").then(normalizeBilling),
  checkDeposit: () => send<Billing>("POST", "/api/billing/check").then(normalizeBilling),
  walletBalances: (address: string) =>
    request<{ balances: WalletBalance[] }>(`/api/billing/wallet?address=${encodeURIComponent(address)}`).then((r) => r.balances ?? []),
  billingHistory: (page: number) => request<DepositPage>(`/api/billing/history?page=${page}`),
  referral: () => request<Referral>("/api/referral"),
  liqmap: () => request<LiqMap>("/api/liqmap"),
};

export type LiqBucket = { price: number; aave?: number; morpho?: number; total: number; count: number };

export type LiqItem = { protocol: "aave" | "morpho"; market: string; address: string; liq: number; coll: number; debt: number };

export type LiqSource = {
  key: string;
  label: string;
  state: "pending" | "syncing" | "ok" | "error" | "disabled";
  positions: number;
  progress: number;
  at?: number;
};

export type LiqMap = {
  ready: boolean;
  pro: boolean;
  at: number;
  ethPrice: number;
  bucket: number;
  delaySec: number;
  buckets: LiqBucket[];
  top: LiqItem[];
  sources: LiqSource[];
};

// A backend older than the frontend may omit fields; a missing list must not crash the whole console.
function normalizeBilling(b: Partial<Billing>): Billing {
  return {
    enabled: Boolean(b.enabled && b.address),
    testnet: Boolean(b.testnet),
    plans: b.plans ?? [],
    chains: (b.chains ?? []).map((c) => ({ ...c, minPrice: c.minPrice ?? 0 })),
    address: b.address,
    deposits: b.deposits ?? [],
  };
}

const REF_KEY = "mechbot_ref";

export function captureRef() {
  const ref = new URLSearchParams(window.location.search).get("ref");
  if (ref && /^\d+$/.test(ref)) localStorage.setItem(REF_KEY, ref);
}

export const storedRef = () => localStorage.getItem(REF_KEY) ?? "";
