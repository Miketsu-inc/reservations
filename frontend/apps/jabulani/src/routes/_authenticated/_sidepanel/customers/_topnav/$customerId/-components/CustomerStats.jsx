import {
  Calendar02Icon,
  Cancel01Icon,
  CheckmarkCircle02Icon,
} from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";
import BookingDonutChart from "./BookingDonutChart";

export default function CustomerStats({ customer }) {
  return (
    <div
      className="flex w-full flex-col gap-2 sm:flex-row sm:justify-start
        sm:gap-0"
    >
      {customer.times_booked !== 0 && (
        <div className="flex h-45 w-full justify-center sm:ml-10 sm:w-1/3">
          <BookingDonutChart
            cancelled={customer.times_cancelled_by_user}
            upcoming={customer.times_upcoming}
            completed={customer.times_completed}
          />
        </div>
      )}
      <div
        className={`flex w-full flex-col justify-center gap-2
          ${customer.times_booked !== 0 ? "sm:w-2/3" : "mt-2"}`}
      >
        <div className="text-text_color flex items-center justify-center gap-4">
          <span className="text-lg font-bold">Total Bookings:</span>
          <span className="text-xl font-bold">{customer.times_booked}</span>
        </div>
        <div className="grid w-full grid-cols-3 gap-4 rounded-lg p-4">
          <StatElement
            value={customer.times_completed}
            color="green-600"
            label="Completed"
          >
            <Icon icon={CheckmarkCircle02Icon} styles="size-7 text-green-600" />
          </StatElement>
          <StatElement
            value={customer.times_cancelled_by_user}
            color="red-600"
            label="Cancelled/No-show"
          >
            <div className="w-min rounded-full border-2 border-red-600">
              <Icon icon={Cancel01Icon} styles="size-5 text-red-600" />
            </div>
          </StatElement>
          <StatElement
            value={customer.times_upcoming}
            color="primary"
            label="Upcoming"
          >
            <Icon icon={Calendar02Icon} styles="size-6 mb-0.5 text-primary" />
          </StatElement>
        </div>
      </div>
    </div>
  );
}

function StatElement({ children, value, color, label }) {
  return (
    <div className="text-center">
      <div
        className={`flex items-center justify-center gap-2 text-2xl font-bold
          text-${color}`}
      >
        {children}
        {value}
      </div>
      <p className="text-text_color/70 mt-1 text-xs">{label}</p>
    </div>
  );
}
