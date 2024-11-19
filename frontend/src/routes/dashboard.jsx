import { createFileRoute, redirect } from "@tanstack/react-router";
import Loading from "../components/Loading";
import { isAuthenticated } from "../lib/lib";
import DashboardPage from "../pages/dashboard/DashboardPage";

export const Route = createFileRoute("/dashboard")({
  component: DashboardPage,
  beforeLoad: async () => {
    if (!(await isAuthenticated("api/v1/auth/user"))) {
      throw redirect({
        to: "/login",
      });
    }
  },
  pendingComponent: Loading,
});
