import ApproveIcon from "@icons/ApproveIcon";
import CalendarIcon from "@icons/CalendarIcon";
import XIcon from "@icons/XIcon";
import BookingDonutChart from "./BookingDonutChart";
export default function CustomerStats({ customer }) {
  const completed =
    customer.times_booked -
    (customer.times_upcoming + customer.times_cancelled);

  return (
    <div className="flex w-full flex-col gap-2 sm:flex-row sm:justify-start sm:gap-0">
      <div className="flex h-[180px] w-full justify-center sm:ml-10 sm:w-1/3">
        <BookingDonutChart
          cancelled={customer.times_cancelled}
          upcoming={customer.times_upcoming}
          completed={completed}
        />
      </div>
      <div className="flex w-full flex-col justify-center gap-2 sm:w-2/3">
        <div className="text-text_color flex items-center justify-center gap-4">
          <span className="text-lg font-bold">Total Bookings:</span>
          <span className="text-xl font-bold">{customer.times_booked}</span>
        </div>
        <div className="grid w-full grid-cols-3 gap-4 rounded-lg p-4">
          <StatElement value={completed} color="green-600" label="Completed">
            <ApproveIcon styles="size-7 stroke-green-600" />
          </StatElement>
          <StatElement
            value={customer.times_cancelled}
            color="red-600"
            label="Cancelled"
          >
            <div className="w-min rounded-full border-2 border-red-600">
              <XIcon styles="size-5 fill-red-600" />
            </div>
          </StatElement>
          <StatElement
            value={customer.times_upcoming}
            color="primary"
            label="Upcoming"
          >
            <CalendarIcon styles="size-6 mb-0.5 stroke-primary" />
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
        className={`flex items-center justify-center gap-2 text-2xl font-bold text-${color}`}
      >
        {children}
        {value}
      </div>
      <p className="text-text_color/70 mt-1 text-xs">{label}</p>
    </div>
  );
}
