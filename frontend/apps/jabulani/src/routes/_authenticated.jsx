import { Loading, ServerError } from "@reservations/components";
import { AuthProvider } from "@reservations/jabulani/lib";
import { meQueryOptions } from "@reservations/lib";
import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated")({
  beforeLoad: async ({ context: { queryClient }, location }) => {
    try {
      const me = await queryClient.ensureQueryData(meQueryOptions());

      const stored = localStorage.getItem("activeMerchantId");
      const membership =
        me.memberships.find((e) => e.merchant_id === stored) ??
        me.memberships[0];

      let authContext = null;

      if (membership) {
        authContext = {
          merchantId: membership?.merchant_id,
          merchantName: membership?.merchant_name,
          merchantUrlName: membership?.merchant_url_name,
          locationId: membership?.location_id,
          employeeId: membership?.employee_id,
          role: membership?.role,
        };

        localStorage.setItem("activeMerchantId", membership.merchant_id);
      } else if (location.pathname !== "/signup") {
        throw redirect({ to: "/signup" });
      }

      return {
        authContext: authContext,
      };
    } catch (error) {
      if (error.status === 401) {
        throw redirect({
          to: "/login",
          search: { redirect: location.href },
        });
      }

      throw error;
    }
  },
  pendingComponent: Loading,
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
  component: AuthOutlet,
});

function AuthOutlet() {
  const { authContext } = Route.useRouteContext();

  return (
    <AuthProvider authContext={authContext}>
      <Outlet />
    </AuthProvider>
  );
}
