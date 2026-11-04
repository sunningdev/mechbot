import { Link } from "react-router-dom";
import { tr } from "../i18n";

export default function Brand() {
  return (
    <Link to="/" className="brand" aria-label={tr("MechBot 首页", "MechBot home")}>
      <img className="brand-logo" src="/logo.svg" alt="" />
      <span>MechBot</span>
    </Link>
  );
}
