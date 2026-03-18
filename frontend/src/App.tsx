import { Navigate, Route, Routes } from "react-router-dom";

import { AppShell } from "./components/AppShell";
import { MemoryPage } from "./pages/MemoryPage";
import { NotFoundPage } from "./pages/NotFoundPage";
import { SearchPage } from "./pages/SearchPage";
import { SummaryPage } from "./pages/SummaryPage";

export default function App() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<Navigate replace to="/memories" />} />
        <Route path="/memories" element={<MemoryPage />} />
        <Route path="/search" element={<SearchPage />} />
        <Route path="/summary" element={<SummaryPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  );
}
