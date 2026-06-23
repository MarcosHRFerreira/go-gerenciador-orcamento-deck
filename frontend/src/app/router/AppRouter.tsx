import { Navigate, Route, Routes } from "react-router-dom";
import { AppShell } from "../../components/layout/AppShell";
import { AdminRoute } from "../../features/auth/components/AdminRoute";
import { ProtectedRoute } from "../../features/auth/components/ProtectedRoute";
import { PublicOnlyRoute } from "../../features/auth/components/PublicOnlyRoute";
import { ChangePasswordPage } from "../../features/auth/pages/ChangePasswordPage";
import { LoginPage } from "../../features/auth/pages/LoginPage";
import { BudgetCreatePage } from "../../features/budgets/pages/BudgetCreatePage";
import BudgetDeliveryMonitorPage from "../../features/budgets/pages/BudgetDeliveryMonitorPage";
import { BudgetEditPage } from "../../features/budgets/pages/BudgetEditPage";
import { BudgetImportPage } from "../../features/budgets/pages/BudgetImportPage";
import { BudgetListPage } from "../../features/budgets/pages/BudgetListPage";
import { CommunicationPage } from "../../features/communication/pages/CommunicationPage";
import { DashboardPage } from "../../features/dashboard/pages/DashboardPage";
import EstimatorCreatePage from "../../features/estimators/pages/EstimatorCreatePage";
import EstimatorEditRoutePage from "../../features/estimators/pages/EstimatorEditRoutePage";
import EstimatorListPage from "../../features/estimators/pages/EstimatorListPage";
import ProjectCreatePage from "../../features/projects/pages/ProjectCreatePage";
import ProjectDetailPage from "../../features/projects/pages/ProjectDetailPage";
import ProjectEditPage from "../../features/projects/pages/ProjectEditPage";
import ProjectListPage from "../../features/projects/pages/ProjectListPage";
import SalespersonListPage from "../../features/salespeople/pages/SalespersonListPage";
import SystemTypeListPage from "../../features/system-types/pages/SystemTypeListPage";
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
          <Route
            path="/budgets/delivery-monitor"
            element={<BudgetDeliveryMonitorPage />}
          />
          <Route path="/budgets/:budgetId/edit" element={<BudgetEditPage />} />
          <Route path="/communication" element={<CommunicationPage />} />
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
            <Route path="/estimators/new" element={<EstimatorCreatePage />} />
            <Route
              path="/estimators/:estimatorId/edit"
              element={<EstimatorEditRoutePage />}
            />
            <Route path="/system-types" element={<SystemTypeListPage />} />
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
