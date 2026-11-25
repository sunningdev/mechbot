import { useEffect } from "react";
import { useLocation } from "react-router-dom";
import { isAdminPathname } from "../adminPath";
import { useLang } from "../i18n";
import { applyPageSeo, seoMeta, setCanonical, setDescription, setHreflang, setJsonLd } from "../seo";

// Blog and legal pages set their own title; posts and legal pages also their own description.
const ownTitle = /^\/(blog|terms|privacy)(\/|$)/;
const ownDescription = /^\/(blog\/|terms$|privacy$)/;

export default function SeoHead() {
  const { pathname } = useLocation();
  const lang = useLang();

  useEffect(() => {
    let robots = document.querySelector("meta[name='robots']");

    if (isAdminPathname(pathname)) {
      document.title = "Admin · MechBot";
      document.documentElement.lang = "en";
      if (!robots) {
        robots = document.createElement("meta");
        robots.setAttribute("name", "robots");
        document.head.appendChild(robots);
      }
      robots.setAttribute("content", "noindex, nofollow");
      setJsonLd({});
      return;
    }

    const known =
      pathname === "/" ||
      pathname === "/blog" ||
      pathname.startsWith("/blog/") ||
      pathname === "/terms" ||
      pathname === "/privacy" ||
      pathname === "/app" ||
      pathname.startsWith("/app/");

    if (robots) {
      robots.setAttribute("content", known ? "index, follow" : "noindex, follow");
    } else if (!known) {
      robots = document.createElement("meta");
      robots.setAttribute("name", "robots");
      robots.setAttribute("content", "noindex, follow");
      document.head.appendChild(robots);
    }

    if (pathname === "/") {
      applyPageSeo({ lang, pathname: "/", home: true });
      return;
    }

    if (!known && !isAdminPathname(pathname)) {
      document.title = "404 · MechBot";
      setDescription("Page not found.");
      return;
    }

    const meta = seoMeta(lang);
    if (!ownTitle.test(pathname)) document.title = meta.title;
    if (!ownDescription.test(pathname)) setDescription(meta.description);
    const canonical = window.location.origin + pathname + (lang === "en" ? "" : `?lang=${lang}`);
    setCanonical(canonical);
    setHreflang(pathname);
  }, [pathname, lang]);

  return null;
}
