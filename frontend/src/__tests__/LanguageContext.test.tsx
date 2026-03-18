import { render, screen, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, beforeEach } from "vitest";

import { LanguageProvider, useLanguage } from "../context/LanguageContext";

// ─── Test helper component ────────────────────────────────────────────────

function LanguageDisplay() {
  const { lang, toggleLang, t } = useLanguage();
  return (
    <div>
      <span data-testid="lang">{lang}</span>
      <span data-testid="apply-label">{t("header.apply")}</span>
      <span data-testid="title">{t("app.title")}</span>
      <button onClick={toggleLang}>toggle</button>
    </div>
  );
}

function renderWithProvider() {
  return render(
    <LanguageProvider>
      <LanguageDisplay />
    </LanguageProvider>
  );
}

// ─── Tests ────────────────────────────────────────────────────────────────

describe("LanguageProvider", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("defaults to English when localStorage is empty", () => {
    renderWithProvider();
    expect(screen.getByTestId("lang").textContent).toBe("en");
    expect(screen.getByTestId("apply-label").textContent).toBe("Apply");
  });

  it("reads initial language from localStorage", () => {
    localStorage.setItem("memory-system.language", "zh");
    renderWithProvider();
    expect(screen.getByTestId("lang").textContent).toBe("zh");
    expect(screen.getByTestId("apply-label").textContent).toBe("应用");
  });

  it("toggles from English to Chinese on button click", async () => {
    const user = userEvent.setup();
    renderWithProvider();

    expect(screen.getByTestId("lang").textContent).toBe("en");

    await user.click(screen.getByRole("button", { name: /toggle/i }));

    expect(screen.getByTestId("lang").textContent).toBe("zh");
    expect(screen.getByTestId("apply-label").textContent).toBe("应用");
  });

  it("toggles from Chinese back to English", async () => {
    const user = userEvent.setup();
    localStorage.setItem("memory-system.language", "zh");
    renderWithProvider();

    await user.click(screen.getByRole("button", { name: /toggle/i }));

    expect(screen.getByTestId("lang").textContent).toBe("en");
    expect(screen.getByTestId("apply-label").textContent).toBe("Apply");
  });

  it("persists language selection to localStorage", async () => {
    const user = userEvent.setup();
    renderWithProvider();

    await user.click(screen.getByRole("button", { name: /toggle/i }));

    expect(localStorage.getItem("memory-system.language")).toBe("zh");

    await user.click(screen.getByRole("button", { name: /toggle/i }));

    expect(localStorage.getItem("memory-system.language")).toBe("en");
  });

  it("t() translates keys correctly in English mode", () => {
    renderWithProvider();
    expect(screen.getByTestId("title").textContent).toBe("Memory System Console");
  });

  it("t() translates keys correctly after toggle to Chinese", async () => {
    const user = userEvent.setup();
    renderWithProvider();

    await user.click(screen.getByRole("button", { name: /toggle/i }));

    expect(screen.getByTestId("title").textContent).toBe("记忆系统控制台");
  });
});

// ─── Parameter interpolation via t() ─────────────────────────────────────

describe("useLanguage().t() with params", () => {
  function ParamDisplay({ id }: { id: string }) {
    const { t } = useLanguage();
    return <span data-testid="result">{t("header.activeUser", { id })}</span>;
  }

  it("interpolates params in English", () => {
    render(
      <LanguageProvider>
        <ParamDisplay id="user-42" />
      </LanguageProvider>
    );
    expect(screen.getByTestId("result").textContent).toBe("Active user: user-42");
  });

  it("interpolates params in Chinese after toggle", async () => {
    const user = userEvent.setup();

    function App() {
      const { toggleLang, t } = useLanguage();
      return (
        <>
          <span data-testid="result">{t("header.activeUser", { id: "u1" })}</span>
          <button onClick={toggleLang}>toggle</button>
        </>
      );
    }

    render(
      <LanguageProvider>
        <App />
      </LanguageProvider>
    );

    await user.click(screen.getByRole("button", { name: /toggle/i }));

    expect(screen.getByTestId("result").textContent).toBe("当前用户：u1");
  });
});

// ─── Error boundary ───────────────────────────────────────────────────────

describe("useLanguage() outside provider", () => {
  it("throws an error when used outside LanguageProvider", () => {
    function Naked() {
      useLanguage();
      return null;
    }

    // Suppress expected console.error from React's error boundary
    const spy = vi.spyOn(console, "error").mockImplementation(() => {});
    expect(() => render(<Naked />)).toThrow("useLanguage must be used within <LanguageProvider>");
    spy.mockRestore();
  });
});
