import { useEffect } from "react";
import { useLang } from "../i18n";
import { legalDoc, type LegalKind } from "../legal";
import { setDescription } from "../seo";
import { BlogShell, renderBlock } from "./Blog";

export default function Legal({ kind }: { kind: LegalKind }) {
  useLang();
  const doc = legalDoc(kind);

  useEffect(() => window.scrollTo(0, 0), [kind]);

  useEffect(() => {
    document.title = `${doc.title} · MechBot`;
    setDescription(doc.summary);
  }, [doc.title, doc.summary]);

  return (
    <BlogShell>
      <article className="post">
        <h1>{doc.title}</h1>
        <p className="post-lead">{doc.summary}</p>
        <p className="muted">{doc.updated}</p>
        <div className="post-body">{doc.body.map(renderBlock)}</div>
      </article>
    </BlogShell>
  );
}
