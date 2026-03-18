import "@testing-library/jest-dom";

// Minimal localStorage mock for test environments that lack it
const store: Record<string, string> = {};

Object.defineProperty(globalThis, "localStorage", {
  value: {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, value: string) => { store[key] = value; },
    removeItem: (key: string) => { delete store[key]; },
    clear: () => { Object.keys(store).forEach((k) => delete store[k]); }
  },
  writable: true
});

// Clear storage between tests
beforeEach(() => {
  localStorage.clear();
});
