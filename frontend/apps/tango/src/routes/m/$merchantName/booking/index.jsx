import { BackArrowIcon } from "@reservations/assets";
import {
  Button,
  Loading,
  ServerError,
  SmallCalendar,
  Textarea,
} from "@reservations/components";
import {
  formatToDateString,
  GetDayPickerWindow,
  getDisplayPrice,
  invalidateLocalStorageAuth,
} from "@reservations/lib";
import {
  keepPreviousData,
  queryOptions,
  useQuery,
} from "@tanstack/react-query";
import { createFileRoute, Link, useRouter } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import "react-day-picker/style.css";
import AvailableTimeSection from "./-components/AvailableTimeSection";

async function fetchHours(merchantName, locationId, serviceId, start, end) {
  start = new Date(start).toJSON();
  end = new Date(end).toJSON();

  const response = await fetch(
    `/api/v1/merchants/available-times?name=${merchantName}&locationId=${locationId}&serviceId=${serviceId}&start=${start}&end=${end}`,
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
    return result.data;
  }
}

function availableTimesQueryOptions(
  merchantName,
  locationId,
  serviceId,
  start,
  end
) {
  return queryOptions({
    queryKey: [
      "available-times",
      merchantName,
      locationId,
      serviceId,
      start,
      end,
    ],
    queryFn: () => fetchHours(merchantName, locationId, serviceId, start, end),
  });
}

async function fetchDisabledDays(merchantName, serviceId) {
  const response = await fetch(
    `/api/v1/merchants/disabled-days?name=${merchantName}&serviceId=${serviceId}`,
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
    return result.data;
  }
}

function disabledDaysQueryOptions(merchantName, serviceId) {
  return queryOptions({
    queryKey: ["disabled-days", merchantName, serviceId],
    queryFn: () => fetchDisabledDays(merchantName, serviceId),
  });
}

async function fetchSummaryInfo(merchantName, serviceId, locationId) {
  const response = await fetch(
    `/api/v1/merchants/summary-info?name=${merchantName}&serviceId=${serviceId}&locationId=${locationId}`,
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
    return result.data;
  }
}

function summaryInfoQueryOptions(merchantName, serviceId, locationId) {
  return queryOptions({
    queryKey: ["summary-info", merchantName, serviceId, locationId],
    queryFn: () => fetchSummaryInfo(merchantName, serviceId, locationId),
  });
}

export const Route = createFileRoute("/m/$merchantName/booking/")({
  component: SelectDateTime,
  loaderDeps: ({ search: { locationId, serviceId } }) => ({
    locationId,
    serviceId,
  }),
  loader: async ({
    params: { merchantName },
    context: { queryClient },
    deps: { serviceId, locationId },
  }) => {
    await queryClient.ensureQueryData(
      summaryInfoQueryOptions(merchantName, serviceId, locationId)
    );

    await queryClient.ensureQueryData(
      disabledDaysQueryOptions(merchantName, serviceId)
    );
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function SelectDateTime() {
  const { merchantName } = Route.useParams();
  const search = Route.useSearch();
  const router = useRouter();
  const [selectedHour, setSelectedHour] = useState();
  const [selectedDay, setSelectedDay] = useState(null);
  const [serverError, setServerError] = useState();
  const [isLoading, setIsLoading] = useState(false);
  const [customerNote, setCustomerNote] = useState("");
  const [currentMonth, setCurrentMonth] = useState(new Date());
  const [skipCount, setSkipCount] = useState(0);
  const [isInitialSkipDone, setIsInitialSkipDone] = useState(false);
  const [windowRange, setWindowRange] = useState(() =>
    GetDayPickerWindow(new Date())
  );

  const {
    data: availableTimesWindow,
    isLoading: atIsLoading,
    isError: atIsError,
    error: atError,
  } = useQuery({
    ...availableTimesQueryOptions(
      merchantName,
      search.locationId,
      search.serviceId,
      windowRange.start,
      windowRange.end
    ),
    placeholderData: keepPreviousData,
  });

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

  const {
    data: summaryInfo,
    isLoading: siIsLoading,
    isError: siIsError,
    error: siError,
  } = useQuery(
    summaryInfoQueryOptions(merchantName, search.serviceId, search.locationId)
  );

  const {
    data: disabledDays,
    isLoading: dsIsLoading,
    isError: dsIsError,
    error: dsError,
  } = useQuery(disabledDaysQueryOptions(merchantName, search.serviceId));

  const unavailableDays = availableTimesWindow
    ?.filter((day) => !day.is_available)
    .map((day) => day.date);
  const unavailableDaysAsDates = unavailableDays?.map((date) => new Date(date));

  function handleMonthChange(month) {
    setCurrentMonth(month);
    setWindowRange(GetDayPickerWindow(month));
  }

  if (atIsError || dsIsError || siIsError) {
    const error = atError || dsError || siError;
    return <ServerError error={error} />;
  }

  async function onSubmitHandler(e) {
    e.preventDefault();

    if (!selectedDay || !selectedHour) {
      setServerError("Please select a day and a time.");
      return;
    }

    const date = new Date(selectedDay.date);

    const [hours, minutes] = selectedHour.split(":").map(Number);
    date.setHours(hours, minutes, 0, 0);
    const timeStamp = date.toISOString();

    setIsLoading(true);

    try {
      const response = await fetch("/api/v1/bookings/new", {
        method: "POST",
        headers: {
          "Content-type": "application/json; charset=UTF-8",
        },
        body: JSON.stringify({
          merchant_name: merchantName,
          service_id: search.serviceId,
          location_id: search.locationId,
          timeStamp: timeStamp,
          customer_note: customerNote,
        }),
      });

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);

        if (response.status === 401) {
          router.navigate({
            from: Route.fullPath,
            to: "/login",
            search: {
              redirect: router.history.location.href,
            },
          });
        }

        const result = await response.json();
        setServerError(result.error.message);
      } else {
        router.navigate({
          from: Route.fullPath,
          to: "completed",
        });
      }
    } catch (err) {
      setServerError(err.message);
    } finally {
      setIsLoading(false);
    }
  }

  function dayChangeHandler(date) {
    const dayString = formatToDateString(date);
    setSelectedHour();
    const selected = availableTimesWindow.find((d) => d.date === dayString);
    setSelectedDay(selected);
  }

  function selectedHourHandler(e) {
    setSelectedHour(e.target.value);
  }

  if (siIsLoading) {
    return <Loading />;
  }

  return (
    <div className="bg-layer_bg mx-auto min-h-screen max-w-7xl px-8">
      <div className="py-5">
        <Link from={Route.fullPath} to="..">
          <button className="inline-flex cursor-pointer gap-1 hover:underline">
            <BackArrowIcon styles="size-6 stroke-gray-500" />
            Back
          </button>
        </Link>
        <ServerError error={serverError} />
        <form method="POST" onSubmit={onSubmitHandler}>
          <div className="flex flex-col pt-5 md:flex-row md:gap-10 lg:pt-10">
            {!(atIsLoading || dsIsLoading) ? (
              <div className="flex flex-col gap-6 md:w-1/2">
                <p className="text-xl sm:py-5">Pick a date</p>
                <div
                  className="flex items-center justify-center self-center
                    bg-white shadow-lg dark:bg-neutral-950"
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
                    endMonth={disabledDays.max_date}
                    animate={true}
                  />
                </div>
                <hr className="border-gray-500" />
                <p className="text-xl sm:py-5">Pick a Time</p>
                <div className="flex flex-col gap-3">
                  <p className="text-lg font-bold">Morning</p>
                  <AvailableTimeSection
                    availableTimes={selectedDay?.morning || []}
                    timeSection="morning"
                    selectedHour={selectedHour}
                    clickedHour={selectedHourHandler}
                  />
                  <p className="text-lg font-bold">Afternoon</p>
                  <AvailableTimeSection
                    availableTimes={selectedDay?.afternoon || []}
                    timeSection="afternoon"
                    selectedHour={selectedHour}
                    clickedHour={selectedHourHandler}
                  />
                </div>
              </div>
            ) : (
              <div className="flex h-dvh items-center justify-center sm:w-1/2">
                <Loading />
              </div>
            )}
            <div className="pt-8 md:flex md:w-1/2 md:flex-col md:pt-0 md:pr-14">
              <div className="hidden md:flex md:flex-col md:gap-6">
                <p className="py-5 text-xl">Summary</p>
                <div className="text-lg *:grid *:grid-cols-2">
                  <div>
                    <p>Merchant:</p>
                    <p>{merchantName}</p>
                  </div>
                  <div>
                    <p>Service:</p>
                    <p>
                      {summaryInfo.name} -{" "}
                      {getDisplayPrice(
                        summaryInfo.price,
                        summaryInfo.price_type
                      )}
                    </p>
                  </div>
                  <div>
                    <p>Location:</p>
                    <p>{summaryInfo.formatted_location}</p>
                  </div>
                  <div className={`${selectedDay?.date ? "" : "invisible"}`}>
                    <p>date:</p>
                    <p>{selectedDay?.date}</p>
                  </div>
                  <div className={`${selectedHour ? "" : "invisible"}`}>
                    <p>Time:</p>
                    <p>{selectedHour}</p>
                  </div>
                </div>
              </div>

              <Textarea
                styles="p-2 max-h-20 min-h-20 md:max-h-32 md:min-h-32"
                id="customerNote"
                name="customerNote"
                labelText="Add a note to your booking"
                required={false}
                placeholder="Add your note here..."
                value={customerNote}
                inputData={(data) => setCustomerNote(data.value)}
              />
              <div
                className="bg-hvr_gray dark:bg-layer_bg fixed right-0 bottom-0
                  left-0 px-8 py-3 md:static md:bg-transparent md:px-0 md:pt-10
                  dark:md:bg-transparent"
              >
                <Button
                  variant="primary"
                  type="submit"
                  disabled={selectedDay && selectedHour ? false : true}
                  isLoading={isLoading}
                  buttonText="Reserve"
                  styles="w-full py-2"
                ></Button>
              </div>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}
