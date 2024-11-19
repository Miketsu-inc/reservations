import { createFileRoute, redirect } from "@tanstack/react-router";
import Loading from "../components/Loading";
import { isAuthenticated } from "../lib/lib";
import SettingsPage from "../pages/settings/SettingsPage";

export const Route = createFileRoute("/settings")({
  component: SettingsPage,
  beforeLoad: async () => {
    if (!(await isAuthenticated("api/v1/auth/user"))) {
      throw redirect({
        to: "/login",
      });
    }
  },
  pendingComponent: Loading,
});
