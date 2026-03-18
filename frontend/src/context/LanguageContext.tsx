import { createContext, useCallback, useContext, useMemo, useState } from "react";
import type { ReactNode } from "react";

import { type Language, type TranslationKey, translate } from "../i18n";

// ─── Context shape ────────────────────────────────────────────────────────

interface LanguageContextValue {
  /** Current active language. */
  lang: Language;
  /** Toggle between "en" and "zh". Persists to localStorage. */
  toggleLang: () => void;
  /** Translate a key, with optional `{param}` interpolation. */
  t: (key: TranslationKey, params?: Record<string, string | number>) => string;
}

const LanguageContext = createContext<LanguageContextValue | null>(null);

// ─── Storage helper ───────────────────────────────────────────────────────

const STORAGE_KEY = "memory-system.language";

function readStoredLang(): Language {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored === "zh" ? "zh" : "en";
  } catch {
    return "en";
  }
}

// ─── Provider ─────────────────────────────────────────────────────────────

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [lang, setLang] = useState<Language>(readStoredLang);

  const toggleLang = useCallback(() => {
    setLang((prev) => {
      const next: Language = prev === "en" ? "zh" : "en";
      try {
        localStorage.setItem(STORAGE_KEY, next);
      } catch {
        // localStorage unavailable (e.g. test environment without mock)
      }
      return next;
    });
  }, []);

  const t = useCallback(
    (key: TranslationKey, params?: Record<string, string | number>) =>
      translate(lang, key, params),
    [lang]
  );

  const value = useMemo(() => ({ lang, toggleLang, t }), [lang, toggleLang, t]);

  return (
    <LanguageContext.Provider value={value}>
      {children}
    </LanguageContext.Provider>
  );
}

// ─── Hook ─────────────────────────────────────────────────────────────────

export function useLanguage(): LanguageContextValue {
  const ctx = useContext(LanguageContext);
  if (!ctx) throw new Error("useLanguage must be used within <LanguageProvider>");
  return ctx;
}
