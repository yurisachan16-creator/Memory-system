import { Fragment } from "react";

interface HighlightTextProps {
  text: string;
  terms: string[];
}

function escapeRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

export function HighlightText({ text, terms }: HighlightTextProps) {
  const uniqueTerms = Array.from(
    new Set(
      terms
        .map((term) => term.trim())
        .filter(Boolean)
        .sort((left, right) => right.length - left.length)
    )
  );

  if (uniqueTerms.length === 0) {
    return <>{text}</>;
  }

  const pattern = new RegExp(`(${uniqueTerms.map(escapeRegExp).join("|")})`, "gi");
  const fragments = text.split(pattern);

  return (
    <>
      {fragments.map((fragment, index) => {
        const isMatch = uniqueTerms.some((term) => term.toLowerCase() === fragment.toLowerCase());
        return isMatch ? (
          <mark key={`${fragment}-${index}`} className="highlight-chip">
            {fragment}
          </mark>
        ) : (
          <Fragment key={`${fragment}-${index}`}>{fragment}</Fragment>
        );
      })}
    </>
  );
}
