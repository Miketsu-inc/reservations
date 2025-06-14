import Card from "@components/Card";
import BackArrowIcon from "@icons/BackArrowIcon";
import CalendarIcon from "@icons/CalendarIcon";
import ClockIcon from "@icons/ClockIcon";
import TickIcon from "@icons/TickIcon";
import TrashBinIcon from "@icons/TrashBinIcon";
import { formatToDateString, timeStringFromDate } from "@lib/datetime";
import { Link } from "@tanstack/react-router";
import { useState } from "react";

export default function AppointmentsList({
  appointments,
  visibleCount,
  onCancel,
  onAccept,
  route,
}) {
  const visibleAppointments = appointments.slice(0, visibleCount);

  return (
    <div className="h-full">
      {visibleAppointments.length > 0 ? (
        <div className="space-y-4">
          {visibleAppointments.map((appointment) => (
            <AppointmentCard
              key={appointment.id}
              appointment={appointment}
              onCancel={(app) => onCancel(app)}
              onAccept={(app) => onAccept(app)}
              route={route}
            />
          ))}
        </div>
      ) : (
        <div
          className="bg-layer_bg flex flex-col items-center justify-center rounded-lg p-4 text-center
            shadow-sm"
        >
          <div className="mb-3 rounded-full bg-gray-300 p-3 dark:bg-gray-700">
            <CalendarIcon styles="size-8 stroke-gray-500 dark:stroke-gray-400" />
          </div>
          <p className="mb-1">No appointments yet</p>
          <p className="mb-3 text-sm text-gray-500 dark:text-gray-400">
            When customers schedule appointments, they will appear here.
          </p>
        </div>
      )}
    </div>
  );
}

function monthDateFormat(date) {
  return date.toLocaleDateString([], {
    weekday: "short",
    month: "short",
    day: "numeric",
  });
}

function AppointmentCard({ appointment, route, onCancel, onAccept }) {
  const [showNote, setShowNote] = useState(false);

  return (
    <Card styles="py-2">
      <div className="flex h-fit flex-row items-center">
        <div
          className="flex w-full flex-col lg:flex-row lg:items-center lg:justify-between lg:pr-3
            xl:pr-6"
        >
          <div className="flex flex-col gap-2 py-1">
            <span className="dark:font-semibold">{`${appointment.last_name} ${appointment.first_name}`}</span>
            <div className="flex flex-row items-center gap-3">
              <span className="text-sm">{`${monthDateFormat(new Date(appointment.from_date))}`}</span>
              <div className="flex flex-row items-center gap-2">
                <ClockIcon styles="size-3 fill-gray-500 dark:fill-gray-400" />
                <span className="text-sm">{`${timeStringFromDate(new Date(appointment.from_date))} - ${timeStringFromDate(new Date(appointment.to_date))}`}</span>
              </div>
            </div>
          </div>
          <span
            className="w-fit rounded-full px-2 py-1 text-xs"
            style={{
              backgroundColor: `${appointment.service_color}20`,
              color: appointment.service_color,
            }}
          >
            {appointment.service_name}
          </span>
        </div>
        <div className="flex flex-row items-center gap-1">
          <Link
            from={route.fullPath}
            to="/calendar"
            params={{
              start: formatToDateString(new Date(appointment.from_date)),
            }}
          >
            <CalendarIcon styles="size-5 stroke-text_color" />
          </Link>
          <button
            className="cursor-pointer ps-1"
            onClick={() => onCancel(appointment)}
          >
            <TrashBinIcon styles="size-5" />
          </button>
          <button
            className="cursor-pointer"
            onClick={() => onAccept(appointment)}
          >
            <TickIcon styles="size-6 stroke-text_color" />
          </button>
        </div>
      </div>
      {appointment.user_note && (
        <div className="pt-3 md:pt-2">
          <button
            className="flex cursor-pointer flex-row items-center gap-2 text-gray-500 dark:text-gray-400"
            onClick={() => setShowNote(!showNote)}
          >
            <BackArrowIcon
              styles={`${showNote ? "rotate-90" : "-rotate-90"} transition-transform duration-300
              size-3 stroke-gray-500 dark:stroke-gray-400`}
            />
            <span className="text-xs">{showNote ? "Hide" : "View"} note</span>
          </button>
          <div
            className={`${showNote ? "mt-2 max-h-8 opacity-100 md:max-h-4" : "max-h-0 opacity-0"}
            overflow-hidden transition-all duration-300`}
          >
            <p className="text-xs text-gray-500 dark:text-gray-400">
              {appointment.user_note}
            </p>
          </div>
        </div>
      )}
    </Card>
  );
}
