import PriceTicker from "./PriceTicker";

export default function StatusStrip() {
  return (
    <div className="status-strip">
      <div className="container status-inner">
        <span className="status-sys">
          <i className="signal-dot" /> SYS.STATUS NOMINAL
        </span>
        <span className="hide-md">INTERVAL 300S</span>
        <PriceTicker />
        <span className="hide-md">ETH · ARB · OP</span>
      </div>
    </div>
  );
}
