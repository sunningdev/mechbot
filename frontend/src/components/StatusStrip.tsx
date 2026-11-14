import PriceTicker from "./PriceTicker";

export default function StatusStrip() {
  return (
    <div className="status-strip">
      <div className="container status-inner">
        <PriceTicker />
        <span className="hide-md status-chains">Ethereum · Arbitrum · Base · Optimism</span>
      </div>
    </div>
  );
}
