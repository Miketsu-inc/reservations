import { createFileRoute, redirect } from "@tanstack/react-router";
import Loading from "../components/Loading";
import { isAuthenticated } from "../lib/lib";

export const Route = createFileRoute("/_authenticated")({
  beforeLoad: async () => {
    if (!(await isAuthenticated("/api/v1/auth/user"))) {
      throw redirect({
        to: "/login",
      });
    }
  },
  pendingComponent: Loading,
});
