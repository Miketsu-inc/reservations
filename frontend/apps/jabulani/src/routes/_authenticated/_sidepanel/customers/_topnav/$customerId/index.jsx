import { ServerError } from "@reservations/components";
import { createFileRoute } from "@tanstack/react-router";
import CustomerBookings from "./-components/CustomerBookings";
import CustomerProfile from "./-components/CustomerProfile";
import CustomerStats from "./-components/CustomerStats";
import {
  ALL_BOOKING_STATUSES,
  customerBookingsQueryOptions,
  customerProfileQueryOptions,
  customerStatsQueryOptions,
} from "./-components/queries";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/customers/_topnav/$customerId/"
)({
  component: CustomerDetailsPage,
  loader: ({
    params,
    context: {
      queryClient,
      authContext: { merchantId },
    },
  }) => {
    void queryClient.prefetchQuery(
      customerProfileQueryOptions(merchantId, params.customerId)
    );
    void queryClient.prefetchQuery(
      customerStatsQueryOptions(merchantId, params.customerId)
    );
    void queryClient.prefetchInfiniteQuery(
      customerBookingsQueryOptions(
        merchantId,
        params.customerId,
        ALL_BOOKING_STATUSES
      )
    );
  },
  errorComponent: ({ error }) => <ServerError error={error.message} />,
});

function CustomerDetailsPage() {
  const { customerId } = Route.useParams();

  return (
    <div className="flex justify-center">
      <div
        className="flex w-full max-w-4xl flex-col gap-10 px-3 py-4 sm:px-5
          sm:py-6"
      >
        <CustomerProfile customerId={customerId} />
        <CustomerStats customerId={customerId} />
        <CustomerBookings customerId={customerId} route={Route} />
      </div>
    </div>
  );
}
