import { useEffect, type ReactNode } from "react";
import { Link, useParams } from "react-router-dom";
import { findPost, getPosts, type Block } from "../blog/posts";
import Brand from "../components/Brand";
import LangToggle from "../components/LangToggle";
import { tr, useLang } from "../i18n";

function BlogShell({ children }: { children: ReactNode }) {
  return (
    <div className="page">
      <header className="topbar">
        <div className="nav container">
          <Brand />
          <nav className="nav-links">
            <Link to="/blog">{tr("博客", "Blog")}</Link>
            <LangToggle />
            <Link className="btn btn-small btn-ghost" to="/">
              <span aria-hidden="true">←</span> {tr("返回首页", "Home")}
            </Link>
            <Link className="btn btn-small btn-primary hide-sm" to="/app">
              {tr("进入控制台", "Open console")}
            </Link>
          </nav>
        </div>
      </header>
      <main className="container blog">{children}</main>
      <footer className="footer">
        <div className="footer-bar">
          <div className="container footer-bar-inner">
            <span>© {new Date().getFullYear()} MechBot. {tr("保留所有权利。", "All rights reserved.")}</span>
            <nav className="footer-links">
              <Link to="/">{tr("首页", "Home")}</Link>
              <Link to="/blog">{tr("博客", "Blog")}</Link>
            </nav>
          </div>
        </div>
      </footer>
    </div>
  );
}

export function PostMeta({ date, tag, readMin }: { date: string; tag: string; readMin: number }) {
  return (
    <div className="post-meta">
      <span className="post-tag">{tag}</span>
      <time>{date}</time>
      <span>{tr(`${readMin} 分钟阅读`, `${readMin} min read`)}</span>
    </div>
  );
}

export function BlogList() {
  const lang = useLang();

  useEffect(() => window.scrollTo(0, 0), []);

  useEffect(() => {
    document.title = tr("博客 · MechBot", "Blog · MechBot");
  }, [lang]);

  return (
    <BlogShell>
      <div className="sec-head blog-head">
        <span className="sec-index sec-num">{tr("BLOG / 技术文章", "BLOG / ENGINEERING NOTES")}</span>
        <h2>{tr("把清算预警这件事讲清楚。", "Liquidation alerts, explained.")}</h2>
        <p className="muted">
          {tr("原理、工程实现和产品取舍，来自 MechBot 的一线经验。", "Principles, engineering and product trade-offs from building MechBot.")}
        </p>
      </div>
      <div className="post-list">
        {getPosts().map((p) => (
          <Link className="post-card frame" to={`/blog/${p.slug}`} key={p.slug}>
            <PostMeta date={p.date} tag={p.tag} readMin={p.readMin} />
            <h3>{p.title}</h3>
            <p className="muted">{p.summary}</p>
            <span className="post-more">
              {tr("阅读全文", "Read more")} <span aria-hidden="true">→</span>
            </span>
          </Link>
        ))}
      </div>
    </BlogShell>
  );
}

function renderBlock(b: Block, i: number) {
  switch (b.type) {
    case "h":
      return <h2 key={i}>{b.text}</h2>;
    case "list":
      return (
        <ul key={i}>
          {b.items.map((t) => (
            <li key={t}>{t}</li>
          ))}
        </ul>
      );
    case "code":
      return <pre key={i}>{b.text}</pre>;
    case "note":
      return (
        <p className="post-note" key={i}>
          {b.text}
        </p>
      );
    default:
      return <p key={i}>{b.text}</p>;
  }
}

export function BlogPost() {
  const { slug } = useParams();
  const post = findPost(slug);
  const title = post?.title;

  useEffect(() => window.scrollTo(0, 0), [slug]);

  useEffect(() => {
    document.title = title ? `${title} · MechBot` : tr("文章不存在 · MechBot", "Post not found · MechBot");
  }, [title]);

  if (!post) {
    return (
      <BlogShell>
        <div className="post-missing">
          <h2>{tr("文章不存在", "Post not found")}</h2>
          <Link className="text-link" to="/blog">
            {tr("← 返回博客", "← Back to blog")}
          </Link>
        </div>
      </BlogShell>
    );
  }

  const others = getPosts().filter((p) => p.slug !== post.slug);

  return (
    <BlogShell>
      <article className="post">
        <Link className="post-back" to="/blog">
          {tr("← 全部文章", "← All posts")}
        </Link>
        <PostMeta date={post.date} tag={post.tag} readMin={post.readMin} />
        <h1>{post.title}</h1>
        <p className="post-lead">{post.summary}</p>
        <div className="post-body">{post.body.map(renderBlock)}</div>
      </article>

      {others.length > 0 && (
        <aside className="post-others">
          <span className="sec-index sec-num">{tr("继续阅读", "Keep reading")}</span>
          <div className="post-list">
            {others.map((p) => (
              <Link className="post-card frame" to={`/blog/${p.slug}`} key={p.slug}>
                <PostMeta date={p.date} tag={p.tag} readMin={p.readMin} />
                <h3>{p.title}</h3>
              </Link>
            ))}
          </div>
        </aside>
      )}
    </BlogShell>
  );
}
