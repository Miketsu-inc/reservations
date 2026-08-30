import { Person } from "@hugeicons/core-free-icons";
import {
  Avatar,
  DatePicker,
  Icon,
  Loading,
  ServerError,
  Textarea,
} from "@reservations/components";
import { invalidateLocalStorageAuth } from "@reservations/lib";
import {
  keepPreviousData,
  queryOptions,
  useQuery,
} from "@tanstack/react-query";
import { useState } from "react";
import "react-day-picker/style.css";
import AvailableTimeSection from "./AvailableTimeSection";
import DaySelector from "./DaySelector";
import { StepContentSkeleton } from "./StepContentSkeleton";

async function fetchAvailableDays(
  merchantName,
  locationId,
  serviceId,
  employeeId
) {
  const params = new URLSearchParams();
  if (employeeId && employeeId !== "no-pref")
    params.append("employee_id", employeeId);

  const response = await fetch(
    `/api/v1/public/merchants/${merchantName}/locations/${locationId}/services/${serviceId}/availability/day`,
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
  if (employeeId && employeeId !== "no-pref")
    params.append("employee_id", employeeId);

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
  employee,
}) {
  const [manualSelectedDay, setManualSelectedDay] = useState(null);
  const [selectedHour, setSelectedHour] = useState(null);
  const [customerNote, setCustomerNote] = useState("");

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
    setSelectedHour(null);
    setManualSelectedDay(dateStr);
    onSelect({ date: dateStr, time: null, customer_note: customerNote });
  }

  function handleDatePickerSelect(pickerDate) {
    const d = new Date(pickerDate);
    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");

    const formatted = `${year}-${month}-${day}`;

    handleDaySelect(formatted);
  }

  function selectedHourHandler(e) {
    setSelectedHour(e.target.value);
    onSelect({
      date: selectedDay,
      time: e.target.value,
      customer_note: customerNote,
    });
  }

  if (daysIsError) return <ServerError error={daysError} />;
  if (daysIsLoading) return <StepContentSkeleton />;

  const hasNoOpenings =
    !timesIsLoading &&
    dayTimes &&
    dayTimes.morning.length === 0 &&
    dayTimes.afternoon.length === 0;

  return (
    <div className="flex h-full w-full max-w-full flex-col">
      <h1 className="text-3xl font-bold">Select Date & Time</h1>

      <div className="mt-10 flex items-center justify-between">
        <div
          className="bg-layer_bg border-border_color flex w-fit items-center
            gap-2 rounded-full border py-1.5 pr-3 pl-2"
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
        <div>
          <DatePicker
            styles="w-min"
            hideText={true}
            value={
              selectedDay ? new Date(selectedDay + "T00:00:00") : undefined
            }
            firstDayOfWeek={"Monday"}
            clearAfterClose={true}
            onSelect={handleDatePickerSelect}
            disabledBefore={new Date(availableDays[0].date + "T00:00:00")}
            disabledAfter={
              new Date(
                availableDays[availableDays.length - 1].date + "T00:00:00"
              )
            }
          />
        </div>
      </div>
      <div
        className="lg:17 bg-bg_color sticky top-14.75 z-10 flex w-full
          max-w-full min-w-0 items-center gap-2 py-4"
      >
        <DaySelector
          days={availableDays}
          selectedDate={selectedDay}
          onSelect={handleDaySelect}
        />
      </div>

      <div className="mt-8 flex w-full flex-1 flex-col gap-6">
        {timesIsError ? (
          <ServerError error={timesError} />
        ) : timesIsLoading ? (
          <Loading />
        ) : hasNoOpenings ? (
          <div
            className="flex flex-col items-center gap-1 px-4 py-10 text-center"
          >
            <p className="font-medium">No availability for this day</p>
            <p className="text-sm text-gray-500">
              There's no place left on this day. Please choose another date.
            </p>
          </div>
        ) : (
          <>
            <div className="flex flex-col gap-3">
              <p className="text-lg font-medium">Morning</p>
              <AvailableTimeSection
                availableTimes={dayTimes?.morning || []}
                timeSection="morning"
                selectedHour={selectedHour}
                clickedHour={selectedHourHandler}
              />
              <p className="mt-4 text-lg font-medium">Afternoon</p>
              <AvailableTimeSection
                availableTimes={dayTimes?.afternoon || []}
                timeSection="afternoon"
                selectedHour={selectedHour}
                clickedHour={selectedHourHandler}
              />
            </div>
          </>
        )}
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
              date: selectedDay,
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
