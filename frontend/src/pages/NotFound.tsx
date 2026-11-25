import { Link } from "react-router-dom";
import Brand from "../components/Brand";
import { tr } from "../i18n";

/** Soft 404: stay on the requested URL (no redirect) so crawlers don't see a redirect loop/home bounce. */
export default function NotFound() {
  return (
    <div className="page center-screen">
      <div className="not-found">
        <Brand />
        <h1>404</h1>
        <p className="muted">{tr("页面不存在。", "This page doesn’t exist.")}</p>
        <Link className="btn btn-primary" to="/">
          {tr("返回首页", "Back to home")}
        </Link>
      </div>
    </div>
  );
}
