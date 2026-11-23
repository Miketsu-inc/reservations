import { DatePicker, Loading, ServerError } from "@reservations/components";
import {
  dateStringToLocalDate,
  formatToDateString,
  invalidateLocalStorageAuth,
} from "@reservations/lib";
import {
  keepPreviousData,
  queryOptions,
  useQuery,
} from "@tanstack/react-query";
import { useState } from "react";
import "react-day-picker/style.css";
import AvailableTimeSection from "./AvailableTimeSection";
import DaySelector from "./DaySelector";
import EmployeePicker from "./EmployeePicker";
import { StepContentSkeleton } from "./StepContentSkeleton";
import TimezoneWarning from "./TimezoneWarning";

async function fetchAvailableDays(
  merchantName,
  locationId,
  serviceId,
  employeeId
) {
  const params = new URLSearchParams();
  if (employeeId !== "no-pref") params.append("employee_id", employeeId);

  const response = await fetch(
    `/api/v1/public/merchants/${merchantName}/locations/${locationId}/services/${serviceId}/availability/available-days?${params.toString()}`,
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

function availableDaysQueryOptions(
  merchantName,
  locationId,
  serviceId,
  employeeId
) {
  return queryOptions({
    queryKey: [
      "available-times",
      merchantName,
      locationId,
      serviceId,
      employeeId,
    ],
    queryFn: () =>
      fetchAvailableDays(merchantName, locationId, serviceId, employeeId),
  });
}

async function fetchDayTimes(
  merchantName,
  locationId,
  serviceId,
  employeeId,
  date
) {
  const params = new URLSearchParams();
  params.append("date", date);
  if (employeeId !== "no-pref") params.append("employee_id", employeeId);

  const response = await fetch(
    `/api/v1/public/merchants/${merchantName}/locations/${locationId}/services/${serviceId}/availability?${params.toString()}`,
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

function dayTimesQueryOptions(
  merchantName,
  locationId,
  serviceId,
  employeeId,
  date
) {
  return queryOptions({
    queryKey: [
      "day-times",
      merchantName,
      locationId,
      serviceId,
      employeeId,
      date,
    ],
    queryFn: () =>
      fetchDayTimes(merchantName, locationId, serviceId, employeeId, date),
    enabled: !!date,
  });
}

export default function AppointmentTimeSelectionStep({
  merchantName,
  locationId,
  serviceId,
  employeeId,
  onSelect,
  onEmployeeChange,
  employee,
}) {
  const [manualSelectedDay, setManualSelectedDay] = useState(null);
  const [selectedSlot, setSelectedSlot] = useState(null);

  function handleEmployeeChange(emp) {
    setSelectedSlot(null);
    onSelect({
      date: selectedDay,
      time: null,
      starts_at: null,
    });
    onEmployeeChange(emp);
  }

  const {
    data: availableDays,
    isLoading: daysIsLoading,
    isError: daysIsError,
    error: daysError,
  } = useQuery({
    ...availableDaysQueryOptions(
      merchantName,
      locationId,
      serviceId,
      employeeId
    ),
    placeholderData: keepPreviousData,
  });

  const firstAvailableDay = availableDays?.find((d) => d.is_available)?.date;
  const selectedDay = manualSelectedDay || firstAvailableDay;

  const {
    data: dayTimes,
    isLoading: timesIsLoading,
    isError: timesIsError,
    error: timesError,
  } = useQuery(
    dayTimesQueryOptions(
      merchantName,
      locationId,
      serviceId,
      employeeId,
      selectedDay
    )
  );

  function handleDaySelect(dateStr) {
    setSelectedSlot(null);
    setManualSelectedDay(dateStr);
    onSelect({
      date: dateStr,
      time: null,
      starts_at: null,
    });
  }

  function handleDatePickerSelect(pickerDate) {
    handleDaySelect(formatToDateString(pickerDate));
  }

  function selectedSlotHandler(slot) {
    setSelectedSlot(slot);
    onSelect({
      date: selectedDay,
      time: slot.time,
      starts_at: slot.starts_at,
    });
  }

  if (daysIsError) return <ServerError error={daysError.message} />;
  if (daysIsLoading) return <StepContentSkeleton />;

  const hasDays = availableDays?.length > 0;
  const firstDay = hasDays ? availableDays[0].date : undefined;
  const lastDay = hasDays
    ? availableDays[availableDays.length - 1].date
    : undefined;

  const hasNoOpenings =
    !selectedDay ||
    (!timesIsLoading &&
      dayTimes &&
      dayTimes.morning.length === 0 &&
      dayTimes.afternoon.length === 0);

  return (
    <div className="flex h-full w-full max-w-full flex-col">
      <h1 className="text-2xl font-semibold">Select Date & Time</h1>

      <TimezoneWarning merchantTimeZone={dayTimes?.time_zone} />

      <div className="mt-6 flex items-center justify-between">
        <EmployeePicker
          merchantName={merchantName}
          locationId={locationId}
          serviceId={serviceId}
          employeeId={employeeId}
          employee={employee}
          onSelectEmployee={handleEmployeeChange}
        />
        <div>
          <DatePicker
            styles="w-min"
            hideText={true}
            value={
              selectedDay
                ? (dateStringToLocalDate(selectedDay) ?? undefined)
                : undefined
            }
            firstDayOfWeek={"Monday"}
            clearAfterClose={true}
            disabled={!hasDays}
            onSelect={handleDatePickerSelect}
            disabledBefore={dateStringToLocalDate(firstDay)}
            disabledAfter={dateStringToLocalDate(lastDay)}
          />
        </div>
      </div>
      {hasDays && (
        <div
          className="bg-bg sticky top-14.75 z-10 flex w-full max-w-full min-w-0
            items-center gap-2 py-4 lg:top-17"
        >
          <DaySelector
            days={availableDays}
            selectedDate={selectedDay}
            onSelect={handleDaySelect}
          />
        </div>
      )}

      <div className="mt-4 flex w-full flex-1 flex-col gap-6">
        {timesIsError ? (
          <ServerError error={timesError.message} />
        ) : timesIsLoading ? (
          <Loading />
        ) : hasNoOpenings ? (
          <div
            className="flex flex-col items-center gap-1 px-4 py-10 text-center"
          >
            {selectedDay ? (
              <>
                <p className="font-medium">No availability for this day</p>
                <p className="text-fg-muted text-sm">
                  There's no place left on this day. Please choose another date.
                </p>
              </>
            ) : (
              <>
                <p className="font-medium">No available times</p>
                <p className="text-fg-muted text-sm">
                  There are no open times for this service right now. Please
                  check back later or choose another professional.
                </p>
              </>
            )}
          </div>
        ) : (
          <>
            <div className="flex flex-col gap-3">
              <p className="text-fg-muted text-sm font-semibold">Morning</p>
              <AvailableTimeSection
                availableTimes={dayTimes?.morning || []}
                timeSection="morning"
                selectedStartsAt={selectedSlot?.starts_at}
                onSelect={selectedSlotHandler}
              />
              <p className="text-fg-muted mt-4 text-sm font-semibold">
                Afternoon
              </p>
              <AvailableTimeSection
                availableTimes={dayTimes?.afternoon || []}
                timeSection="afternoon"
                selectedStartsAt={selectedSlot?.starts_at}
                onSelect={selectedSlotHandler}
              />
            </div>
          </>
        )}
      </div>
    </div>
  );
}
