import { useEffect } from "react";
import { useLocation } from "react-router-dom";
import { useLang } from "../i18n";
import { seoMeta, setCanonical, setDescription } from "../seo";

// Blog pages set their own title (and posts their own description).
export default function SeoHead() {
  const { pathname } = useLocation();
  const lang = useLang();

  useEffect(() => {
    const meta = seoMeta(lang);
    if (!pathname.startsWith("/blog")) document.title = meta.title;
    if (!pathname.startsWith("/blog/")) setDescription(meta.description);
    setCanonical(window.location.origin + pathname + (lang === "en" ? "" : `?lang=${lang}`));
  }, [pathname, lang]);

  return null;
}
