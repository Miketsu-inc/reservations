import Card from "@components/Card";
import Loading from "@components/Loading";
import Select from "@components/Select";
import ServerError from "@components/ServerError";
import { useWindowSize } from "@lib/hooks";
import { fillStatisticsWithDate, invalidateLocalStorageAuth } from "@lib/lib";
import {
  keepPreviousData,
  queryOptions,
  useQuery,
} from "@tanstack/react-query";
import { createFileRoute, useRouteContext } from "@tanstack/react-router";
import { useState } from "react";
import AppointmentsList from "./-components/AppointmentsList";
import LowStockProductsAlert from "./-components/LowStockProductsAlert";
import RevenueChart from "./-components/RevenueChart";
import StatisticsCard from "./-components/StatisticsCard";

async function fetchDashboardData(period) {
  const date = new Date().toJSON();

  const response = await fetch(
    `/api/v1/merchants/dashboard?date=${date}&period=${period}`,
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
    if (result.data !== null) {
      return result.data;
    }
  }
}

function dashboardQueryOptions(period) {
  return queryOptions({
    queryKey: ["dashboard", period],
    queryFn: () => fetchDashboardData(period),
  });
}

export const Route = createFileRoute("/_authenticated/_sidepanel/dashboard")({
  component: DashboardPage,
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
  pendingComponent: Loading,
});

function DashboardPage() {
  const windowSize = useWindowSize();
  const [period, setPeriod] = useState(7);
  const { data, isLoading, isError, error, isFetching } = useQuery({
    ...dashboardQueryOptions(period),
    placeholderData: keepPreviousData,
  });
  const { queryClient } = useRouteContext({ from: Route.id });

  async function invalidateDashBoardData() {
    await queryClient.invalidateQueries({
      queryKey: ["dashboard"],
    });
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  if (isLoading) {
    return <Loading />;
  }

  return (
    <div className="flex h-full flex-col px-4 py-2 md:px-0 md:py-0 lg:h-[90svh]">
      <div className="flex flex-row items-center justify-between py-3">
        <p className="text-xl">Your dashboard</p>
        <Select
          styles="w-36!"
          options={[
            { value: 7, label: "Last 7 days" },
            { value: 30, label: "Last 30 days" },
          ]}
          value={period}
          onSelect={(option) => {
            if (period !== option.value) {
              setPeriod(option.value);
            }
          }}
          disabled={isFetching}
        />
      </div>
      <div className="flex h-full w-full flex-col gap-4 lg:flex-row lg:gap-6">
        <div className="flex h-full flex-1 flex-col gap-4 lg:max-w-1/2">
          <div className="flex h-fit flex-col gap-4">
            <div
              className="flex h-fit flex-row items-center justify-between gap-4"
            >
              <StatisticsCard
                title="Revenue"
                text={`${data.statistics.revenue_sum}`}
                percent={data.statistics.revenue_change}
                tooltip={windowSize !== "sm" && windowSize !== "md"}
                tooltipText="Calculated by adding up all your completed appointments for this period"
              />
              <StatisticsCard
                title="Appointments"
                text={data.statistics.appointments}
                percent={data.statistics.appointments_change}
                tooltip={windowSize !== "sm" && windowSize !== "md"}
                tooltipText="The amount of completed appointments in this period"
              />
              {windowSize === "lg" ||
              windowSize === "2xl" ||
              windowSize === "3xl" ? (
                <StatisticsCard
                  title="Cancellations"
                  text={data.statistics.cancellations}
                  percent={data.statistics.cancellations_change}
                  tooltip={windowSize !== "sm" && windowSize !== "md"}
                  tooltipText="The amount of cancelled appointments (by customers) in this period"
                />
              ) : (
                <></>
              )}
              {windowSize === "3xl" ? (
                <StatisticsCard
                  title="Average duration"
                  text={data.statistics.average_duration}
                  percent={data.statistics.average_duration_change}
                  tooltip={windowSize !== "sm" && windowSize !== "md"}
                  tooltipText="The average duration of services from your completed appointments in this period"
                />
              ) : (
                <></>
              )}
            </div>
            <Card styles="flex h-80 flex-col gap-2">
              <RevenueChart
                data={fillStatisticsWithDate(
                  data.statistics.revenue,
                  data.period_start,
                  data.period_end
                )}
              />
            </Card>
          </div>
          <div className="flex flex-1 flex-col gap-2">
            <LowStockProductsAlert
              products={data.low_stock_products}
              route={Route}
            />
          </div>
        </div>
        <div className="flex h-full flex-1 flex-col gap-4 lg:max-w-1/2">
          <p className="text-lg">Upcoming appointments</p>
          <div className="flex max-h-1/2 flex-col gap-2 rounded-lg">
            <AppointmentsList
              appointments={data.upcoming_appointments}
              visibleCount={1}
              onAccept={() => {}}
              onCancel={invalidateDashBoardData}
              route={Route}
            />
          </div>
          <p className="text-lg">Latest bookings</p>
          <div className="flex max-h-1/2 flex-col gap-2 rounded-lg">
            <AppointmentsList
              appointments={data.latest_bookings}
              visibleCount={3}
              onAccept={() => {}}
              onCancel={invalidateDashBoardData}
              route={Route}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
