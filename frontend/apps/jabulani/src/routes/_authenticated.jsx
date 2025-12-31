import { Loading, ServerError } from "@reservations/components";
import { isAuthenticated } from "@reservations/lib";
import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated")({
  beforeLoad: async () => {
    if (!(await isAuthenticated("/api/v1/auth/"))) {
      throw redirect({
        to: "/login",
      });
    }
  },
  pendingComponent: Loading,
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});
