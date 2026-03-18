import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { MemoryRouter } from "react-router-dom";

import { AppShell } from "../components/AppShell";
import { LanguageProvider } from "../context/LanguageContext";
import { UserProvider } from "../context/UserContext";

// ─── Mocks ────────────────────────────────────────────────────────────────

// Ant Design App context is required for message hooks; mock minimally.
vi.mock("antd", async (importOriginal) => {
  const antd = await importOriginal<typeof import("antd")>();
  return {
    ...antd,
    App: {
      ...antd.App,
      useApp: () => ({ message: { warning: vi.fn(), error: vi.fn(), success: vi.fn() } })
    }
  };
});

// react-router-dom Outlet renders nothing in these unit tests.
vi.mock("react-router-dom", async (importOriginal) => {
  const rrd = await importOriginal<typeof import("react-router-dom")>();
  return { ...rrd, Outlet: () => null };
});

// ─── Helper ───────────────────────────────────────────────────────────────

function renderAppShell() {
  return render(
    <MemoryRouter initialEntries={["/memories"]}>
      <LanguageProvider>
        <UserProvider>
          <AppShell />
        </UserProvider>
      </LanguageProvider>
    </MemoryRouter>
  );
}

// ─── Language toggle button ───────────────────────────────────────────────

describe("AppShell – language toggle button", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("renders the toggle button with '中文' label when language is English", () => {
    renderAppShell();
    // The button shows the target language name (what you switch TO)
    expect(screen.getByRole("button", { name: /中文/i })).toBeInTheDocument();
  });

  it("renders the toggle button with 'English' label when language is Chinese", () => {
    localStorage.setItem("memory-system.language", "zh");
    renderAppShell();
    expect(screen.getByRole("button", { name: /english/i })).toBeInTheDocument();
  });

  it("switches UI to Chinese after clicking the toggle button", async () => {
    const user = userEvent.setup();
    renderAppShell();

    // Initial state: English
    expect(screen.getByRole("button", { name: /中文/i })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /中文/i }));

    // Now in Chinese: nav items translated, toggle shows "English"
    expect(screen.getByRole("button", { name: /english/i })).toBeInTheDocument();
    expect(screen.getByText("记忆列表")).toBeInTheDocument();
    expect(screen.getByText("搜索")).toBeInTheDocument();
    expect(screen.getByText("摘要")).toBeInTheDocument();
  });

  it("switches UI back to English after two toggle clicks", async () => {
    const user = userEvent.setup();
    renderAppShell();

    await user.click(screen.getByRole("button", { name: /中文/i }));
    await user.click(screen.getByRole("button", { name: /english/i }));

    expect(screen.getByRole("button", { name: /中文/i })).toBeInTheDocument();
    expect(screen.getByText("Memories")).toBeInTheDocument();
  });
});

// ─── Header strings ───────────────────────────────────────────────────────

describe("AppShell – header text localisation", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("renders English app title by default", () => {
    renderAppShell();
    expect(screen.getByText("Memory System Console")).toBeInTheDocument();
  });

  it("renders Chinese app title after toggle", async () => {
    const user = userEvent.setup();
    renderAppShell();

    await user.click(screen.getByRole("button", { name: /中文/i }));

    expect(screen.getByText("记忆系统控制台")).toBeInTheDocument();
  });

  it("renders 'Apply' button in English mode", () => {
    renderAppShell();
    expect(screen.getByRole("button", { name: /^apply$/i })).toBeInTheDocument();
  });

  it("renders '应用' button in Chinese mode", async () => {
    const user = userEvent.setup();
    renderAppShell();

    await user.click(screen.getByRole("button", { name: /中文/i }));

    expect(screen.getByRole("button", { name: /应用/ })).toBeInTheDocument();
  });

  it("shows 'User not set' tag in English when no user is active", () => {
    renderAppShell();
    expect(screen.getByText("User not set")).toBeInTheDocument();
  });

  it("shows '未设置用户' tag in Chinese when no user is active", async () => {
    const user = userEvent.setup();
    renderAppShell();

    await user.click(screen.getByRole("button", { name: /中文/i }));

    expect(screen.getByText("未设置用户")).toBeInTheDocument();
  });
});

// ─── Navigation ───────────────────────────────────────────────────────────

describe("AppShell – navigation items", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("renders all three nav items in English", () => {
    renderAppShell();
    expect(screen.getByText("Memories")).toBeInTheDocument();
    expect(screen.getByText("Search")).toBeInTheDocument();
    expect(screen.getByText("Summary")).toBeInTheDocument();
  });

  it("renders all three nav items in Chinese after toggle", async () => {
    const user = userEvent.setup();
    renderAppShell();

    await user.click(screen.getByRole("button", { name: /中文/i }));

    expect(screen.getByText("记忆列表")).toBeInTheDocument();
    expect(screen.getByText("搜索")).toBeInTheDocument();
    expect(screen.getByText("摘要")).toBeInTheDocument();
  });
});
