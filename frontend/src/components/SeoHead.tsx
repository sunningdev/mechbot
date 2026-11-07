import { useEffect } from "react";
import { useLocation } from "react-router-dom";
import { useLang } from "../i18n";
import { seoMeta, setCanonical, setDescription } from "../seo";

// Blog and legal pages set their own title; posts and legal pages also their own description.
const ownTitle = /^\/(blog|terms|privacy)(\/|$)/;
const ownDescription = /^\/(blog\/|terms$|privacy$)/;

export default function SeoHead() {
  const { pathname } = useLocation();
  const lang = useLang();

  useEffect(() => {
    const meta = seoMeta(lang);
    if (!ownTitle.test(pathname)) document.title = meta.title;
    if (!ownDescription.test(pathname)) setDescription(meta.description);
    setCanonical(window.location.origin + pathname + (lang === "en" ? "" : `?lang=${lang}`));
  }, [pathname, lang]);

  return null;
}
