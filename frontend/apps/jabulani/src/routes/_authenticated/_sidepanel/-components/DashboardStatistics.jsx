import { Loading, ServerError } from "@reservations/components";
import { invalidateLocalStorageAuth, useWindowSize } from "@reservations/lib";
import {
  keepPreviousData,
  queryOptions,
  useQuery,
} from "@tanstack/react-query";
import StatisticsCard from "./StatisticsCard";

const DASHBOARD_CACHE_TIME = 5 * 60 * 1000;

async function fetchDashboardStatistics(merchantId, period) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/dashboard/statistics?period=${period}`,
    {
      method: "GET",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    }
  );

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

function dashboardStatisticsQueryOptions(merchantId, period) {
  return queryOptions({
    queryKey: [merchantId, "dashboard-statistics", period],
    queryFn: () => fetchDashboardStatistics(merchantId, period),
    placeholderData: keepPreviousData,
    staleTime: 30_000,
    gcTime: DASHBOARD_CACHE_TIME,
  });
}

export default function DashboardStatistics({ merchantId, period }) {
  const { isWindowSmall, windowSize } = useWindowSize();
  const { data, isLoading, isError, error } = useQuery(
    dashboardStatisticsQueryOptions(merchantId, period)
  );

  if (isError) {
    return <ServerError error={error.message} />;
  }

  if (isLoading) {
    return <Loading />;
  }

  return (
    <div className="flex h-fit flex-row items-center justify-between gap-4">
      <StatisticsCard
        title="Revenue"
        text={`${data.revenue_sum}`}
        percent={data.revenue_change}
        tooltip={!isWindowSmall}
        tooltipText="Calculated by adding up all your completed bookings for this period"
      />
      <StatisticsCard
        title="Bookings"
        text={data.bookings}
        percent={data.bookings_change}
        tooltip={!isWindowSmall}
        tooltipText="The amount of completed bookings in this period"
      />
      {(windowSize === "lg" || windowSize === "2xl") && (
        <StatisticsCard
          title="Cancellations"
          text={data.cancellations}
          percent={data.cancellations_change}
          tooltip={!isWindowSmall}
          tooltipText="The amount of cancelled bookings (by customers) in this period"
        />
      )}
      {windowSize === "2xl" && (
        <StatisticsCard
          title="Average duration"
          text={data.average_duration}
          percent={data.average_duration_change}
          tooltip={!isWindowSmall}
          tooltipText="The average duration of services from your completed bookings in this period"
        />
      )}
    </div>
  );
}
