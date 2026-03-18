import type { Category, Source } from "../types/memory";
import { enUS } from "./translations/en-US";
import { zhCN } from "./translations/zh-CN";

// ─── Language types ────────────────────────────────────────────────────────

export type Language = "en" | "zh";

export const translations: Record<Language, typeof enUS> = {
  en: enUS,
  zh: zhCN as unknown as typeof enUS
};

// ─── Type-safe dot-notation key ───────────────────────────────────────────

type Namespace = keyof typeof enUS;

export type TranslationKey = {
  [N in Namespace]: `${N}.${keyof (typeof enUS)[N] & string}`;
}[Namespace];

// ─── Domain label key maps ────────────────────────────────────────────────

export const categoryKeys: Record<Category, TranslationKey> = {
  preference: "common.catPreference",
  identity: "common.catIdentity",
  goal: "common.catGoal",
  context: "common.catContext"
};

export const sourceKeys: Record<Source, TranslationKey> = {
  chat: "common.srcChat",
  manual: "common.srcManual",
  system: "common.srcSystem"
};

// ─── Translation function ─────────────────────────────────────────────────

/**
 * Translate a dot-notation key for the given language.
 * Supports `{param}` interpolation via the optional `params` argument.
 *
 * @example
 *   translate("zh", "header.apply")              // → "应用"
 *   translate("en", "memory.total", { count: 5}) // → "5 memories"
 */
export function translate(
  lang: Language,
  key: TranslationKey,
  params?: Record<string, string | number>
): string {
  const parts = key.split(".");
  let value: unknown = translations[lang];

  for (const part of parts) {
    if (typeof value !== "object" || value === null) return key;
    value = (value as Record<string, unknown>)[part];
  }

  if (typeof value !== "string") return key;

  if (!params) return value;

  return Object.entries(params).reduce(
    (str, [k, v]) => str.replace(`{${k}}`, String(v)),
    value
  );
}
