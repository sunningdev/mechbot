import { useEffect, useState } from "react";

const SYMBOLS = ["BTC", "ETH", "SOL"] as const;
type Symbol = (typeof SYMBOLS)[number];
type Quote = { price: number; change: number };

const STREAMS = SYMBOLS.map((s) => `${s.toLowerCase()}usdt@miniTicker`).join("/");
// data-stream.binance.vision serves the same public market data where the main host is blocked.
const HOSTS = ["wss://stream.binance.com:9443", "wss://data-stream.binance.vision"];

function useBinanceQuotes() {
  const [quotes, setQuotes] = useState<Partial<Record<Symbol, Quote>>>({});

  useEffect(() => {
    let ws: WebSocket | undefined;
    let retry: number | undefined;
    let attempt = 0;
    let closed = false;

    const connect = () => {
      ws = new WebSocket(`${HOSTS[attempt % HOSTS.length]}/stream?streams=${STREAMS}`);
      ws.onopen = () => {
        attempt = 0;
      };
      ws.onmessage = (event) => {
        const msg = JSON.parse(event.data as string) as { data?: { s: string; c: string; o: string } };
        const d = msg.data;
        if (!d) return;
        const symbol = d.s.replace(/USDT$/, "") as Symbol;
        const price = Number(d.c);
        const open = Number(d.o);
        if (!Number.isFinite(price) || !open) return;
        setQuotes((q) => ({ ...q, [symbol]: { price, change: (price / open - 1) * 100 } }));
      };
      ws.onclose = () => {
        if (closed) return;
        attempt += 1;
        retry = window.setTimeout(connect, Math.min(30_000, 1000 * 2 ** Math.min(attempt, 5)));
      };
      ws.onerror = () => ws?.close();
    };

    connect();
    return () => {
      closed = true;
      window.clearTimeout(retry);
      ws?.close();
    };
  }, []);

  return quotes;
}

function formatPrice(n: number) {
  return n.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

export default function PriceTicker() {
  const quotes = useBinanceQuotes();

  return (
    <span className="ticker">
      {SYMBOLS.map((s) => {
        const q = quotes[s];
        const tone = !q ? "" : q.change >= 0 ? "up" : "down";
        return (
          <span className={`ticker-item ${tone} ${s === "SOL" ? "hide-sm" : ""}`} key={s}>
            {s} <b>{q ? formatPrice(q.price) : "—"}</b>
            {q && (
              <em>
                {q.change >= 0 ? "+" : ""}
                {q.change.toFixed(2)}%
              </em>
            )}
          </span>
        );
      })}
    </span>
  );
}
