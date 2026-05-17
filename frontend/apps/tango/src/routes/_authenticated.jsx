import { Loading, ServerError } from "@reservations/components";
import { meQueryOptions } from "@reservations/lib";
import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated")({
  beforeLoad: async ({ context: { queryClient } }) => {
    try {
      await queryClient.ensureQueryData(meQueryOptions());
    } catch (error) {
      if (error.staus === 401) {
        throw redirect({
          to: "/login",
          search: { redirect: location.href },
        });
      }
    }
  },
  component: AuthComponent,
  pendingComponent: Loading,
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function AuthComponent() {
  return <Outlet />;
}
