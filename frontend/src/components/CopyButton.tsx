import { useState } from "react";
import { copy } from "../format";
import { tr } from "../i18n";

export default function CopyButton({ text, label }: { text: string; label?: string }) {
  const [done, setDone] = useState(false);
  return (
    <button
      className="btn btn-small btn-ghost"
      onClick={async () => {
        if (await copy(text)) {
          setDone(true);
          window.setTimeout(() => setDone(false), 1500);
        }
      }}
    >
      {done ? tr("已复制", "Copied") : (label ?? tr("复制", "Copy"))}
    </button>
  );
}
