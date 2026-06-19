import { Navigate, Route, Routes } from "react-router-dom";
import { AppShell } from "../../components/layout/AppShell";
import { AdminRoute } from "../../features/auth/components/AdminRoute";
import { ProtectedRoute } from "../../features/auth/components/ProtectedRoute";
import { PublicOnlyRoute } from "../../features/auth/components/PublicOnlyRoute";
import { ChangePasswordPage } from "../../features/auth/pages/ChangePasswordPage";
import { LoginPage } from "../../features/auth/pages/LoginPage";
import { BudgetCreatePage } from "../../features/budgets/pages/BudgetCreatePage";
import { BudgetEditPage } from "../../features/budgets/pages/BudgetEditPage";
import { BudgetImportPage } from "../../features/budgets/pages/BudgetImportPage";
import { BudgetListPage } from "../../features/budgets/pages/BudgetListPage";
import { DashboardPage } from "../../features/dashboard/pages/DashboardPage";
import EstimatorListPage from "../../features/estimators/pages/EstimatorListPage";
import ProjectCreatePage from "../../features/projects/pages/ProjectCreatePage";
import ProjectDetailPage from "../../features/projects/pages/ProjectDetailPage";
import ProjectEditPage from "../../features/projects/pages/ProjectEditPage";
import ProjectListPage from "../../features/projects/pages/ProjectListPage";
import SalespersonListPage from "../../features/salespeople/pages/SalespersonListPage";
import { UserCreatePage } from "../../features/users/pages/UserCreatePage";
import { UserEditPage } from "../../features/users/pages/UserEditPage";
import { UserListPage } from "../../features/users/pages/UserListPage";

export function AppRouter() {
  return (
    <Routes>
      <Route element={<PublicOnlyRoute />}>
        <Route path="/login" element={<LoginPage />} />
      </Route>
      <Route element={<ProtectedRoute />}>
        <Route path="/change-password" element={<ChangePasswordPage />} />
        <Route element={<AppShell />}>
          <Route path="/" element={<Navigate replace to="/budgets" />} />
          <Route path="/budgets" element={<BudgetListPage />} />
          <Route path="/budgets/:budgetId/edit" element={<BudgetEditPage />} />
          <Route path="/projects" element={<ProjectListPage />} />
          <Route path="/projects/new" element={<ProjectCreatePage />} />
          <Route
            path="/projects/:projectId/edit"
            element={<ProjectEditPage />}
          />
          <Route path="/projects/:projectId" element={<ProjectDetailPage />} />
          <Route element={<AdminRoute />}>
            <Route path="/dashboard" element={<DashboardPage />} />
            <Route path="/budgets/import" element={<BudgetImportPage />} />
            <Route path="/budgets/new" element={<BudgetCreatePage />} />
            <Route path="/salespeople" element={<SalespersonListPage />} />
            <Route path="/estimators" element={<EstimatorListPage />} />
            <Route path="/users" element={<UserListPage />} />
            <Route path="/users/new" element={<UserCreatePage />} />
            <Route path="/users/:userId/edit" element={<UserEditPage />} />
          </Route>
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/budgets" replace />} />
    </Routes>
  );
}
