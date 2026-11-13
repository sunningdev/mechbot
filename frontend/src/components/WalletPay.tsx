import "@rainbow-me/rainbowkit/styles.css";
import { connectorsForWallets, lightTheme, RainbowKitProvider, useConnectModal, type Locale } from "@rainbow-me/rainbowkit";
import {
  binanceWallet,
  coinbaseWallet,
  injectedWallet,
  metaMaskWallet,
  okxWallet,
  rabbyWallet,
  rainbowWallet,
  trustWallet,
  walletConnectWallet,
} from "@rainbow-me/rainbowkit/wallets";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useEffect, useMemo, useRef, useState } from "react";
import { BaseError, erc20Abi, formatUnits, parseUnits, UserRejectedRequestError, type Address, type Chain as ViemChain } from "viem";
import { createConfig, http, useAccount, useDisconnect, useSwitchChain, useWriteContract, WagmiProvider, type Config } from "wagmi";
import {
  arbitrum,
  arbitrumSepolia,
  base,
  baseSepolia,
  bsc,
  bscTestnet,
  mainnet,
  optimism,
  optimismSepolia,
  sepolia,
} from "wagmi/chains";
import { api, type Chain, type Coin, type WalletBalance } from "../api";
import { getLang, tr, type Lang } from "../i18n";
import { coinOrder, sortChains } from "../payorder";
import ChainLogo from "./ChainLogo";

export type WalletPayProps = {
  chains: Chain[];
  deposit: string;
  price: number;
  projectId: string;
  onSent: (hash: string, chain: Chain) => void;
};

const viemChains: Record<number, ViemChain> = Object.fromEntries(
  [mainnet, arbitrum, base, optimism, bsc, sepolia, arbitrumSepolia, baseSepolia, optimismSepolia, bscTestnet].map((c) => [c.id, c]),
);

const locales: Record<Lang, Locale> = {
  zh: "zh-CN",
  "zh-TW": "zh-TW",
  en: "en-US",
  ru: "ru-RU",
  vi: "vi-VN",
  es: "es-419",
  fr: "fr-FR",
  pt: "pt-BR",
};

const gasCoin = (key: string) => (key === "bsc" ? "BNB" : "ETH");

const queryClient = new QueryClient();
const configs = new Map<string, Config>();

// wagmi keeps connection state in its config, so one config per network set lives for the whole session.
function walletConfig(chains: ViemChain[], projectId: string) {
  const key = projectId + ":" + chains.map((c) => c.id).join(",");
  let config = configs.get(key);
  if (!config) {
    const wallets = projectId
      ? [metaMaskWallet, rabbyWallet, coinbaseWallet, okxWallet, binanceWallet, trustWallet, rainbowWallet, walletConnectWallet]
      : [injectedWallet, rabbyWallet, coinbaseWallet];
    const connectors = connectorsForWallets([{ groupName: tr("钱包", "Wallets"), wallets }], {
      appName: "MechBot",
      projectId: projectId || "none",
      appUrl: window.location.origin,
    });
    config = createConfig({
      chains: chains as [ViemChain, ...ViemChain[]],
      connectors,
      transports: Object.fromEntries(chains.map((c) => [c.id, http()])),
    });
    configs.set(key, config);
  }
  return config;
}

export default function WalletPay(props: WalletPayProps) {
  const evm = useMemo(() => sortChains(props.chains).filter((c) => c.chainId && viemChains[c.chainId] && c.coins?.length), [props.chains]);
  const chains = evm.map((c) => viemChains[c.chainId!]);
  const config = walletConfig(chains, props.projectId);
  return (
    <WagmiProvider config={config}>
      <QueryClientProvider client={queryClient}>
        <RainbowKitProvider
          locale={locales[getLang()]}
          modalSize="compact"
          initialChain={chains[0]}
          theme={lightTheme({ accentColor: "#1f1f1f", accentColorForeground: "#ffffff", borderRadius: "small" })}
        >
          <Pay {...props} chains={evm} />
        </RainbowKitProvider>
      </QueryClientProvider>
    </WagmiProvider>
  );
}

type Option = { chain: Chain; coin: Coin; amount: bigint; balance?: bigint; enough: boolean; gas: boolean };

function options(chains: Chain[], price: number, balances: WalletBalance[] | null): Option[] {
  const order = coinOrder();
  const out: Option[] = [];
  for (const chain of chains) {
    const b = balances?.find((x) => x.chain === chain.key && !x.error);
    const held = (c: Coin) => (b?.coins[c.address] !== undefined ? BigInt(b.coins[c.address]) : undefined);
    // Optimism accepts two USDT contracts; offer the one the wallet holds more of.
    const bySymbol = new Map<string, Coin>();
    for (const coin of chain.coins ?? []) {
      const prev = bySymbol.get(coin.symbol);
      if (!prev || (held(coin) ?? 0n) > (held(prev) ?? 0n)) bySymbol.set(coin.symbol, coin);
    }
    const coins = [...bySymbol.values()].sort((x, y) => order.indexOf(x.symbol) - order.indexOf(y.symbol));
    for (const coin of coins) {
      const amount = parseUnits(String(price), coin.decimals);
      const balance = held(coin);
      out.push({
        chain,
        coin,
        amount,
        balance,
        enough: balance !== undefined && balance >= amount,
        gas: b !== undefined && BigInt(b.native) > 0n,
      });
    }
  }
  return out;
}

const optionKey = (o: Option) => `${o.chain.key}:${o.coin.address}`;
const short = (a: string) => `${a.slice(0, 6)}…${a.slice(-4)}`;
const amountText = (v: bigint, decimals: number) => Number(formatUnits(v, decimals)).toLocaleString(undefined, { maximumFractionDigits: 2 });

function Pay({ chains, deposit, price, onSent }: WalletPayProps) {
  const { address, isConnected, chainId, status } = useAccount();
  const { openConnectModal } = useConnectModal();
  const { disconnect } = useDisconnect();
  const { switchChainAsync } = useSwitchChain();
  const { writeContractAsync } = useWriteContract();
  const [balances, setBalances] = useState<WalletBalance[] | null>(null);
  const [failed, setFailed] = useState(false);
  const [picked, setPicked] = useState("");
  const [choosing, setChoosing] = useState(false);
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState<{ ok: boolean; text: string } | null>(null);
  const opened = useRef(false);

  // Opens the wallet list once, after wagmi has finished restoring a wallet connected on an earlier visit.
  useEffect(() => {
    if (status === "disconnected" && openConnectModal && !opened.current) {
      opened.current = true;
      openConnectModal();
    }
  }, [status, openConnectModal]);

  useEffect(() => {
    setBalances(null);
    setFailed(false);
    setPicked("");
    if (!address) return;
    let live = true;
    api.walletBalances(address).then(
      (b) => live && setBalances(b),
      () => live && setFailed(true),
    );
    return () => {
      live = false;
    };
  }, [address]);

  const list = options(chains, price, balances);
  const auto = list.find((o) => o.enough && o.gas);
  const selected = list.find((o) => optionKey(o) === picked) ?? auto ?? list[0];
  const loading = isConnected && !balances && !failed;
  const nothing = balances !== null && !auto;

  async function pay() {
    if (!selected) return;
    setBusy(true);
    setNote({ ok: true, text: tr("请在钱包里确认…", "Confirm in your wallet…") });
    try {
      if (chainId !== selected.chain.chainId) await switchChainAsync({ chainId: selected.chain.chainId! });
      const hash = await writeContractAsync({
        address: selected.coin.address as Address,
        abi: erc20Abi,
        functionName: "transfer",
        args: [deposit as Address, selected.amount],
        chainId: selected.chain.chainId!,
      });
      setNote(null);
      onSent(hash, selected.chain);
    } catch (err) {
      const rejected = err instanceof BaseError && err.walk((e) => e instanceof UserRejectedRequestError) !== null;
      const reason = err instanceof BaseError ? err.shortMessage : String(err);
      setNote({ ok: false, text: rejected ? tr("已取消", "Cancelled") : `${tr("支付失败", "Payment failed")}: ${reason}` });
    } finally {
      setBusy(false);
    }
  }

  if (!isConnected || !address) {
    return (
      <div className="wallet-pay">
        <button className="btn btn-primary btn-block" onClick={() => openConnectModal?.()}>
          {tr("连接钱包", "Connect wallet")}
        </button>
      </div>
    );
  }

  return (
    <div className="wallet-pay">
      <div className="wallet-line small-text">
        <span className="muted">{tr("钱包", "Wallet")}</span>
        <span className="mono">{short(address)}</span>
        <button type="button" className="threshold-link" onClick={() => disconnect()}>
          {tr("断开", "Disconnect")}
        </button>
      </div>

      {loading && <p className="muted small-text">{tr("正在读取钱包余额…", "Reading wallet balances…")}</p>}
      {failed && <p className="muted small-text">{tr("读取余额失败，可以直接选择网络和币种。", "Couldn't read balances. Pick a network and coin yourself.")}</p>}
      {nothing && (
        <p className="form-error small-text">
          {tr(
            "这个钱包在各网络上都没有足够的 USDC 或 USDT。可以换一个钱包，或改用转账到专属地址。",
            "This wallet doesn't hold enough USDC or USDT on any network. Try another wallet, or send to your deposit address instead.",
          )}
        </p>
      )}

      {selected && !loading && (
        <>
          <button className="btn btn-primary btn-block" onClick={() => void pay()} disabled={busy}>
            {tr(
              `在 ${selected.chain.name} 上支付 ${price} ${selected.coin.symbol}`,
              `Pay ${price} ${selected.coin.symbol} on ${selected.chain.name}`,
            )}
          </button>
          {balances && !nothing && (!selected.enough || !selected.gas) && (
            <p className="form-error small-text">
              {!selected.enough
                ? tr(`${selected.coin.symbol} 余额不足`, `Not enough ${selected.coin.symbol}`)
                : tr(`缺少 ${gasCoin(selected.chain.key)} 手续费`, `No ${gasCoin(selected.chain.key)} for gas`)}
            </p>
          )}
          <button type="button" className="threshold-link wallet-change" onClick={() => setChoosing((v) => !v)}>
            {choosing ? tr("收起", "Hide") : tr("更换网络或币种", "Change network or coin")}
          </button>
        </>
      )}

      {(choosing || nothing) && !loading && (
        <ul className="wallet-options">
          {list.map((o) => (
            <li key={optionKey(o)}>
              <button
                type="button"
                className={selected && optionKey(o) === optionKey(selected) ? "on" : ""}
                onClick={() => {
                  setPicked(optionKey(o));
                  setChoosing(false);
                }}
              >
                <ChainLogo chain={o.chain.key} />
                <span>{o.chain.name}</span>
                <span className="coin-chip">{o.coin.symbol}</span>
                <span className={`wallet-bal ${o.enough && o.gas ? "" : "muted"}`}>
                  {o.balance === undefined
                    ? "—"
                    : !o.gas
                      ? tr(`缺少 ${gasCoin(o.chain.key)} 手续费`, `No ${gasCoin(o.chain.key)} for gas`)
                      : tr(`余额 ${amountText(o.balance, o.coin.decimals)}`, `Balance ${amountText(o.balance, o.coin.decimals)}`)}
                </span>
              </button>
            </li>
          ))}
        </ul>
      )}

      {note && <p className={`small-text ${note.ok ? "muted" : "form-error"}`}>{note.text}</p>}
      <p className="muted small-text wallet-safe">
        {tr("只发起一笔转账，不请求授权或签名，收款方是你的专属地址：", "This sends one transfer only, with no approvals or signatures, to your personal deposit address:")}{" "}
        <span className="mono">{short(deposit)}</span>
      </p>
    </div>
  );
}
