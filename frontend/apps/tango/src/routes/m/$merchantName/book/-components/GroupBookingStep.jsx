import {
  Clock01Icon,
  Tick02Icon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";
import {
  Avatar,
  DatePicker,
  Icon,
  ServerError,
  Textarea,
} from "@reservations/components";
import {
  dateStringToLocalDate,
  formatDuration,
  getDisplayPrice,
  invalidateLocalStorageAuth,
} from "@reservations/lib";
import {
  keepPreviousData,
  queryOptions,
  useQuery,
} from "@tanstack/react-query";
import { useMemo, useState } from "react";
import "react-day-picker/style.css";
import EmployeePicker from "./EmployeePicker";
import { StepContentSkeleton } from "./StepContentSkeleton";
import TimezoneWarning from "./TimezoneWarning";

async function fetchAvailableGroupBookings(
  merchantName,
  locationId,
  serviceId,
  employeeId
) {
  const params = new URLSearchParams();
  if (employeeId !== "no-pref" && employeeId) {
    params.append("employee_id", employeeId);
  }

  const response = await fetch(
    `/api/v1/public/merchants/${merchantName}/locations/${locationId}/services/${serviceId}/availability/group?${params.toString()}`,
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
  }
  return result.data;
}

function availableGroupBookingsQueryOptions(
  merchantName,
  locationId,
  serviceId,
  employeeId
) {
  return queryOptions({
    queryKey: [
      "available-group-bookings",
      merchantName,
      locationId,
      serviceId,
      employeeId,
    ],
    queryFn: () =>
      fetchAvailableGroupBookings(
        merchantName,
        locationId,
        serviceId,
        employeeId
      ),
    enabled: Boolean(merchantName && locationId && serviceId),
  });
}

export default function GroupBookingStep({
  merchantName,
  locationId,
  serviceId,
  employeeId,
  onSelect,
  onEmployeeChange,
  employee,
}) {
  const [selectedSessionId, setSelectedSessionId] = useState(null);
  const [customerNote, setCustomerNote] = useState("");

  const {
    data: availability,
    isLoading,
    isError,
    error,
  } = useQuery({
    ...availableGroupBookingsQueryOptions(
      merchantName,
      locationId,
      serviceId,
      employeeId
    ),
    placeholderData: keepPreviousData,
  });

  const groupBookings = availability?.bookings ?? [];
  const groupedBookings = useMemo(() => {
    const byDay = Map.groupBy(availability?.bookings ?? [], (s) => s.date);

    return [...byDay].map(([dateKey, sessions]) => {
      const date = dateStringToLocalDate(sessions[0].date);
      return {
        dateKey,
        dayNum: date.getDate(),
        weekdayShort: date.toLocaleDateString("en-US", { weekday: "short" }),
        monthShort: date.toLocaleDateString("en-US", { month: "short" }),
        sessions,
      };
    });
  }, [availability]);

  function handleEmployeeChange(emp) {
    setSelectedSessionId(null);
    onSelect({
      date: null,
      time: null,
      starts_at: null,
      bookingId: null,
      customer_note: customerNote,
    });
    onEmployeeChange(emp);
  }

  function handleSessionSelect(session) {
    if (selectedSessionId === session.id) {
      setSelectedSessionId(null);
      onSelect({
        date: null,
        time: null,
        starts_at: null,
        bookingId: null,
        customer_note: customerNote,
      });
    } else {
      setSelectedSessionId(session.id);
      onSelect({
        date: session.date,
        time: session.time,
        starts_at: session.from_date,
        bookingId: session.id,
        customer_note: customerNote,
        employee: {
          id: session.employee_id,
          first_name: session.employee_first_name,
          last_name: session.employee_last_name,
        },
      });
    }
  }

  if (isError)
    return <ServerError error={error?.message || "Failed to load events"} />;
  if (isLoading) return <StepContentSkeleton />;

  return (
    <div className="flex h-full w-full max-w-full flex-col gap-6">
      <h1 className="text-3xl font-bold">Select an Event</h1>

      <TimezoneWarning merchantTimeZone={availability?.time_zone} />

      <div className="flex w-full items-center justify-between">
        <EmployeePicker
          merchantName={merchantName}
          serviceId={serviceId}
          employeeId={employeeId}
          employee={employee}
          onSelectEmployee={handleEmployeeChange}
        />
        <div>
          <DatePicker
            styles="w-min"
            hideText={true}
            firstDayOfWeek="Monday"
            clearAfterClose={true}
          />
        </div>
      </div>
      <div className="flex w-full flex-1 flex-col gap-6">
        {groupBookings.length === 0 ? (
          <div
            className="flex flex-col items-center gap-2 px-4 py-16 text-center"
          >
            <Icon icon={UserGroupIcon} styles="size-12 text-gray-400" />
            <p className="text-lg font-medium">No upcoming sessions</p>
            <p className="max-w-sm text-sm text-gray-500">
              There are currently no scheduled sessions available for this
              service. Please check back later or choose another service.
            </p>
          </div>
        ) : (
          <div className="flex flex-col gap-6 sm:gap-8">
            {groupedBookings.map((group) => {
              const groupHasSelected = group.sessions.some(
                (s) => s.id === selectedSessionId
              );

              return (
                <div
                  key={group.dateKey}
                  className="relative flex w-full flex-col gap-2.5"
                >
                  <div
                    className="flex items-baseline gap-2 select-none xl:hidden"
                  >
                    <span
                      className={`text-2xl leading-none font-bold
                        ${groupHasSelected ? "text-primary" : "text-text_color"}`}
                    >
                      {group.dayNum}
                    </span>
                    <span className="text-text_color/60 text-sm font-medium">
                      {group.weekdayShort}, {group.monthShort}
                    </span>
                  </div>

                  <div
                    className="hidden xl:absolute xl:top-2 xl:-left-20 xl:flex
                      xl:w-16 xl:flex-col xl:items-start xl:text-left"
                  >
                    <span
                      className={`block text-3xl leading-none font-bold
                        tracking-tight transition-colors ${
                          groupHasSelected ? "text-primary" : "text-text_color"
                        }`}
                    >
                      {group.dayNum}
                    </span>
                    <span
                      className="text-text_color/60 mt-1 block text-sm
                        leading-tight font-medium"
                    >
                      {group.weekdayShort}, {group.monthShort}
                    </span>
                  </div>

                  <ul className="flex w-full min-w-0 flex-1 flex-col gap-3">
                    {group.sessions.map((session) => (
                      <SessionCard
                        key={session.id}
                        session={session}
                        isSelected={selectedSessionId === session.id}
                        onSelect={handleSessionSelect}
                      />
                    ))}
                  </ul>
                </div>
              );
            })}
          </div>
        )}

        <Textarea
          styles="p-2 min-h-24"
          id="customerNote"
          name="customerNote"
          labelText="Add a note to your booking (Optional)"
          placeholder="E.g., Any special requirements or requests..."
          value={customerNote}
          inputData={(data) => {
            setCustomerNote(data.value);
            const currentSession = groupBookings.find(
              (s) => s.id === selectedSessionId
            );
            if (currentSession) {
              onSelect({
                date: currentSession.date,
                time: currentSession.time,
                starts_at: currentSession.from_date,
                bookingId: currentSession.id,
                customer_note: data.value,
                employee: {
                  id: currentSession.employee_id,
                  first_name: currentSession.employee_first_name,
                  last_name: currentSession.employee_last_name,
                },
              });
            }
          }}
          required={false}
        />
      </div>
    </div>
  );
}

function SessionCard({ session, isSelected, onSelect }) {
  const spotsLeft = session.max_participants - session.current_participants;

  const durationMinutes = Math.max(
    0,
    Math.round(
      (new Date(session.to_date).getTime() -
        new Date(session.from_date).getTime()) /
        60000
    )
  );

  return (
    <li
      role="radio"
      aria-checked={isSelected}
      onClick={() => onSelect(session)}
      className={`bg-layer_bg border-border_color dark:hover:bg-layer_bg/80 flex
        w-full cursor-pointer flex-col justify-between gap-3.5 rounded-md border
        p-5 hover:bg-gray-100/40 hover:shadow-sm ${
          isSelected
            ? " border-primary "
            : "hover:border-gray-400 dark:hover:border-zinc-700"
        }`}
    >
      <div className="flex items-center justify-between gap-3">
        <div className="flex min-w-0 flex-wrap items-center gap-2.5">
          <span className="text-text_color text-xl font-semibold">
            {session.time} - {session.end_time}
          </span>

          <span
            className="text-text_color/70 rounded-full bg-gray-200/80 px-3 py-1
              text-xs font-semibold dark:bg-gray-400/25"
          >
            {spotsLeft} spots left
          </span>
        </div>

        <div
          className={`flex size-7 shrink-0 items-center justify-center
            rounded-full border transition-all ${
              isSelected
                ? "border-primary bg-primary text-white shadow-xs"
                : "border-gray-300 dark:border-gray-600"
            }`}
        >
          {isSelected && <Icon icon={Tick02Icon} styles="size-4.5" />}
        </div>
      </div>

      <div className="flex items-center justify-between gap-3">
        <div className="flex min-w-0 items-center gap-2.5">
          <Avatar
            styles="size-7! text-[10px] shrink-0 rounded-full!"
            initials={`${session.employee_first_name[0]}${session.employee_last_name[0]}`}
          />
          <span className="text-text_color truncate text-sm">
            {session.employee_first_name} {session.employee_last_name}
          </span>
          <span className="text-text_color/30 text-xs">•</span>

          <div
            className="text-text_color flex shrink-0 items-center gap-1 text-sm"
          >
            <Icon icon={Clock01Icon} styles="size-4" />
            <span>{formatDuration(durationMinutes)}</span>
          </div>
        </div>

        <span className="text-text_color shrink-0 font-medium">
          {getDisplayPrice(session.price, session.price_type)}
        </span>
      </div>
    </li>
  );
}
