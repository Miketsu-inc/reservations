import Card from "@components/Card";
import Select from "@components/Select";
import ServerError from "@components/ServerError";
import { useWindowSize } from "@lib/hooks";
import { fillStatisticsWithDate, invalidateLocalSotrageAuth } from "@lib/lib";
import { createFileRoute, useRouter } from "@tanstack/react-router";
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
    }
  );

  const result = await response.json();

  if (!response.ok) {
    invalidateLocalSotrageAuth(response.status);
    throw result.error;
  } else {
    if (result.data !== null) {
      return result.data;
    }
  }
}

export const Route = createFileRoute("/_authenticated/_sidepanel/dashboard")({
  component: DashboardPage,
  validateSearch: (search) => {
    if (search.period && (search.period === 7 || search.period === 30)) {
      return search;
    } else {
      return {
        period: 7,
      };
    }
  },
  loaderDeps: ({ search: { period } }) => ({ period }),
  loader: async ({ deps: { period } }) => {
    const dashboardData = await fetchDashboardData(period);

    return {
      crumb: "Dashboard",
      data: dashboardData,
    };
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function DashboardPage() {
  const windowSize = useWindowSize();
  const search = Route.useSearch();
  const loaderData = Route.useLoaderData();
  const router = useRouter();

  return (
    <div className="flex h-full flex-col px-4 py-2 md:px-0 md:py-0 lg:h-[90svh]">
      <div className="flex flex-row items-center justify-between py-3">
        <p className="text-xl">Your dashboard</p>
        <Select
          styles="w-36"
          options={[
            { value: 7, label: "Last 7 days" },
            { value: 30, label: "Last 30 days" },
          ]}
          value={search.period}
          onSelect={(option) => {
            if (search.period !== option.value) {
              router.navigate({
                search: () => ({ period: option.value }),
                replace: true,
              });
            }
          }}
        />
      </div>
      <div className="flex h-full w-full flex-col gap-4 lg:flex-row lg:gap-6">
        <div className="flex h-full flex-1 flex-col gap-4 lg:max-w-1/2">
          <div className="flex h-fit flex-col gap-4">
            <div className="flex h-fit flex-row items-center justify-between gap-4">
              <StatisticsCard
                title="Revenue"
                text={`${loaderData.data.statistics.revenue_sum} HUF`}
                percent={loaderData.data.statistics.revenue_change}
              />
              <StatisticsCard
                title="Appointments"
                text={loaderData.data.statistics.appointments}
                percent={loaderData.data.statistics.appointments_change}
              />
              {windowSize === "lg" ||
              windowSize === "2xl" ||
              windowSize === "3xl" ? (
                <StatisticsCard
                  title="Cancellations"
                  text={loaderData.data.statistics.cancellations}
                  percent={loaderData.data.statistics.cancellations_change}
                />
              ) : (
                <></>
              )}
              {windowSize === "3xl" ? (
                <StatisticsCard
                  title="Average duration"
                  text={loaderData.data.statistics.average_duration}
                  percent={loaderData.data.statistics.average_duration_change}
                />
              ) : (
                <></>
              )}
            </div>
            <Card styles="flex h-80 flex-col gap-2">
              <RevenueChart
                data={fillStatisticsWithDate(
                  loaderData.data.statistics.revenue,
                  loaderData.data.period_start,
                  loaderData.data.period_end
                )}
              />
            </Card>
          </div>
          <div className="flex flex-1 flex-col gap-2">
            <LowStockProductsAlert
              products={loaderData.data.low_stock_products}
              route={Route}
            />
          </div>
        </div>
        <div className="flex h-full flex-1 flex-col gap-4 lg:max-w-1/2">
          <p className="text-lg">Upcoming appointments</p>
          <div className="flex max-h-1/2 flex-col gap-2 rounded-lg">
            <AppointmentsList
              appointments={loaderData.data.upcoming_appointments}
              visibleCount={1}
              onAccept={() => {}}
              onCancel={router.invalidate}
              route={Route}
            />
          </div>
          <p className="text-lg">Latest bookings</p>
          <div className="flex max-h-1/2 flex-col gap-2 rounded-lg">
            <AppointmentsList
              appointments={loaderData.data.latest_bookings}
              visibleCount={3}
              onAccept={() => {}}
              onCancel={router.invalidate}
              route={Route}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
