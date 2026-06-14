import { Navigate, Route, Routes } from "react-router-dom";
import { AppShell } from "../../components/layout/AppShell";
import { ProtectedRoute } from "../../features/auth/components/ProtectedRoute";
import { PublicOnlyRoute } from "../../features/auth/components/PublicOnlyRoute";
import { LoginPage } from "../../features/auth/pages/LoginPage";
import { BudgetCreatePage } from "../../features/budgets/pages/BudgetCreatePage";
import { BudgetEditPage } from "../../features/budgets/pages/BudgetEditPage";
import { BudgetListPage } from "../../features/budgets/pages/BudgetListPage";
import { DashboardPage } from "../../features/dashboard/pages/DashboardPage";

export function AppRouter() {
  return (
    <Routes>
      <Route element={<PublicOnlyRoute />}>
        <Route path="/login" element={<LoginPage />} />
      </Route>
      <Route element={<ProtectedRoute />}>
        <Route element={<AppShell />}>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/budgets" element={<BudgetListPage />} />
          <Route path="/budgets/new" element={<BudgetCreatePage />} />
          <Route path="/budgets/:budgetId/edit" element={<BudgetEditPage />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
