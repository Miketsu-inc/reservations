import { Card, Select, ServerError } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import DashboardBookingsList from "./-components/DashboardBookingsList";
import DashboardStatistics from "./-components/DashboardStatistics";
import LowStockProductsAlert from "./-components/LowStockProductsAlert";
import RevenueChart from "./-components/RevenueChart";

export const Route = createFileRoute("/_authenticated/_sidepanel/dashboard")({
  component: DashboardPage,
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function DashboardPage() {
  const [period, setPeriod] = useState(7);
  const { merchantId } = useAuth();

  return (
    <div className="flex h-full flex-col px-4 pt-4 lg:h-[90svh]">
      <div className="flex flex-row items-center justify-between pb-3">
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
        />
      </div>
      <div className="flex h-full w-full flex-col gap-4 lg:flex-row lg:gap-6">
        <div className="flex h-full flex-1 flex-col gap-4 lg:max-w-1/2">
          <div className="flex h-fit flex-col gap-4">
            <DashboardStatistics merchantId={merchantId} period={period} />
            <Card styles="flex h-80 flex-col gap-2">
              <RevenueChart merchantId={merchantId} period={period} />
            </Card>
          </div>
          <div className="flex flex-1 flex-col gap-2">
            <LowStockProductsAlert merchantId={merchantId} route={Route} />
          </div>
        </div>
        <div className="flex h-full flex-1 flex-col gap-4 lg:max-w-1/2">
          <p className="text-lg">Upcoming bookings</p>
          <div className="flex max-h-1/2 flex-col gap-2 rounded-lg">
            <DashboardBookingsList
              merchantId={merchantId}
              view="upcoming"
              route={Route}
            />
          </div>
          <p className="text-lg">Latest bookings</p>
          <div className="flex max-h-1/2 flex-col gap-2 rounded-lg">
            <DashboardBookingsList
              merchantId={merchantId}
              view="latest"
              route={Route}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
