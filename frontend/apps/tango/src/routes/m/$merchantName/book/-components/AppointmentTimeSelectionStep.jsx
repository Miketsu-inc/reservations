import { Person } from "@hugeicons/core-free-icons";
import {
  Avatar,
  Icon,
  ServerError,
  SmallCalendar,
  Textarea,
} from "@reservations/components";
import {
  formatToDateString,
  GetDayPickerWindow,
  invalidateLocalStorageAuth,
} from "@reservations/lib";
import {
  keepPreviousData,
  queryOptions,
  useQuery,
} from "@tanstack/react-query";
import { useEffect, useState } from "react";
import "react-day-picker/style.css";
import AvailableTimeSection from "./AvailableTimeSection";
import { StepContentSkeleton } from "./StepContentSkeleton";

async function fetchHours(
  merchantName,
  locationId,
  serviceId,
  employeeId,
  start,
  end
) {
  const params = new URLSearchParams();
  params.append("start", new Date(start).toJSON());
  params.append("end", new Date(end).toJSON());
  if (employeeId && employeeId !== "no-pref")
    params.append("employee_id", employeeId);

  const queryString = params.toString();
  const response = await fetch(
    `/api/v1/public/merchants/${merchantName}/locations/${locationId}/services/${serviceId}/availability?${queryString}`,
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

function availableTimesQueryOptions(
  merchantName,
  locationId,
  serviceId,
  employeeId,
  start,
  end
) {
  return queryOptions({
    queryKey: [
      "available-times",
      merchantName,
      locationId,
      serviceId,
      employeeId,
      start,
      end,
    ],
    queryFn: () =>
      fetchHours(merchantName, locationId, serviceId, employeeId, start, end),
  });
}

async function fetchDisabledDays(
  merchantName,
  locationId,
  serviceId,
  employeeId
) {
  const params = new URLSearchParams();
  if (employeeId && employeeId !== "no-pref")
    params.append("employee_id", employeeId);

  const queryString = params.toString();
  const response = await fetch(
    `/api/v1/public/merchants/${merchantName}/locations/${locationId}/services/${serviceId}/availability/disabled-days?${queryString}`,
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

function disabledDaysQueryOptions(
  merchantName,
  locationId,
  serviceId,
  employeeId
) {
  return queryOptions({
    queryKey: [
      "disabled-days",
      merchantName,
      locationId,
      serviceId,
      employeeId,
    ],
    queryFn: () =>
      fetchDisabledDays(merchantName, locationId, serviceId, employeeId),
  });
}

export default function AppointmentTimeSelectionStep({
  merchantName,
  locationId,
  serviceId,
  employeeId,
  onSelect,
  employee,
}) {
  const [selectedDay, setSelectedDay] = useState(null);
  const [selectedHour, setSelectedHour] = useState(null);
  const [customerNote, setCustomerNote] = useState("");

  const [currentMonth, setCurrentMonth] = useState(new Date());
  const [windowRange, setWindowRange] = useState(() =>
    GetDayPickerWindow(new Date())
  );
  const [skipCount, setSkipCount] = useState(0);
  const [isInitialSkipDone, setIsInitialSkipDone] = useState(false);

  const {
    data: availableTimesWindow,
    isLoading: atIsLoading,
    isError: atIsError,
    error: atError,
  } = useQuery({
    ...availableTimesQueryOptions(
      merchantName,
      locationId,
      serviceId,
      employeeId,
      windowRange.start,
      windowRange.end
    ),
    placeholderData: keepPreviousData,
  });

  const {
    data: disabledDays,
    isLoading: dsIsLoading,
    isError: dsIsError,
    error: dsError,
  } = useQuery(
    disabledDaysQueryOptions(merchantName, locationId, serviceId, employeeId)
  );

  // TODO: solve this
  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (!availableTimesWindow || skipCount > 3 || isInitialSkipDone) return;

    const first = availableTimesWindow.find((d) => d.is_available);

    if (!first) {
      const nextMonth = new Date(currentMonth);
      nextMonth.setMonth(nextMonth.getMonth() + 1);
      setCurrentMonth(nextMonth);
      setSkipCount((prev) => prev + 1);
      setWindowRange(GetDayPickerWindow(nextMonth));
    } else {
      setSelectedDay(first);
      setIsInitialSkipDone(true);
    }
  }, [availableTimesWindow, currentMonth, isInitialSkipDone, skipCount]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const unavailableDaysAsDates =
    availableTimesWindow
      ?.filter((day) => !day.is_available)
      .map((day) => new Date(day.date)) || [];

  function handleMonthChange(month) {
    setCurrentMonth(month);
    setWindowRange(GetDayPickerWindow(month));
  }

  function dayChangeHandler(date) {
    const dayString = formatToDateString(date);
    setSelectedHour(null);
    const selected = availableTimesWindow.find((d) => d.date === dayString);
    setSelectedDay(selected);
    onSelect({
      date: selected,
      time: null,
      customer_note: customerNote,
    });
  }

  function selectedHourHandler(e) {
    setSelectedHour(e.target.value);
    onSelect({
      date: selectedDay.date,
      time: e.target.value,
      customer_note: customerNote,
    });
  }

  if (atIsError || dsIsError) return <ServerError error={atError || dsError} />;
  if (atIsLoading || dsIsLoading) return <StepContentSkeleton />;

  return (
    <div className="flex h-full w-full flex-col gap-10">
      <h1 className="text-3xl font-bold">Select Date & Time</h1>
      <div
        className="bg-layer_bg border-border_color flex w-fit items-center gap-2
          rounded-full border py-1.5 pr-3 pl-2"
      >
        {employeeId === "no-pref" ? (
          <>
            <div
              className="bg-primary/80 flex size-8 items-center justify-center
                rounded-full"
            >
              <Icon icon={Person} styles="size-5 text-white" />
            </div>
            <span className="text-sm font-medium">No Preference</span>
          </>
        ) : (
          <>
            <Avatar
              styles="size-8! text-[10px]! shrink-0 rounded-full!"
              img={employee?.avatar_url}
              initials={`${employee.first_name[0]}${employee.last_name[0]}`}
            />
            <span className="text-sm font-medium">
              {employee.first_name} {employee.last_name}
            </span>
          </>
        )}
      </div>
      <div
        className="bg-layer_bg border-border_color flex items-center
          justify-center self-center rounded-sm border px-4 py-2 shadow-md"
      >
        <SmallCalendar
          value={selectedDay?.date}
          onSelect={dayChangeHandler}
          disabled={[
            { dayOfWeek: disabledDays.closed_days },
            { before: disabledDays.min_date },
            { after: disabledDays.max_date },
            ...unavailableDaysAsDates,
          ]}
          disabledTodayStyling={true}
          firstDayOfWeek={"Monday"}
          month={currentMonth}
          onMonthChange={handleMonthChange}
          startMonth={new Date()}
          endMonth={new Date(disabledDays.max_date)}
          animate={true}
        />
      </div>

      <div className="flex w-full flex-1 flex-col gap-6">
        <div className="flex flex-col gap-3">
          <p className="text-lg font-medium">Morning</p>
          <AvailableTimeSection
            availableTimes={selectedDay?.morning || []}
            timeSection="morning"
            selectedHour={selectedHour}
            clickedHour={selectedHourHandler}
          />
          <p className="mt-4 text-lg font-medium">Afternoon</p>
          <AvailableTimeSection
            availableTimes={selectedDay?.afternoon || []}
            timeSection="afternoon"
            selectedHour={selectedHour}
            clickedHour={selectedHourHandler}
          />
        </div>

        <Textarea
          styles="p-2 min-h-24"
          id="customerNote"
          name="customerNote"
          labelText="Add a note to your booking (Optional)"
          placeholder="E.g., I have sensitive skin..."
          value={customerNote}
          inputData={(data) => {
            setCustomerNote(data.value);
            onSelect({
              date: selectedDay.date,
              time: selectedHour,
              customer_note: data.value,
            });
          }}
          required={false}
        />
      </div>
    </div>
  );
}
