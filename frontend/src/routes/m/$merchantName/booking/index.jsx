import Button from "@components/Button";
import ServerError from "@components/ServerError";
import SmallCalendar from "@components/SmallCalendar";
import Textarea from "@components/Textarea";
import BackArrowIcon from "@icons/BackArrowIcon";
import { formatToDateString } from "@lib/datetime";
import { invalidateLocalStorageAuth } from "@lib/lib";
import { createFileRoute, Link, useRouter } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import "react-day-picker/style.css";
import AvailableTimeSection from "./-components/AvailableTimeSection";

async function fetchHours(merchantName, locationId, serviceId, day) {
  const response = await fetch(
    `/api/v1/merchants/available-times?name=${merchantName}&locationId=${locationId}&serviceId=${serviceId}&day=${day}`,
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

async function fetchClosedDays(merchantName) {
  const response = await fetch(
    `/api/v1/merchants/business-hours/closed?name=${merchantName}`,
    {
      method: "GET",
      headers: {
        Accept: "application/json",
        "constent-type": "application/json",
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

export const Route = createFileRoute("/m/$merchantName/booking/")({
  component: SelectDateTime,
  loaderDeps: ({ search: { locationId, serviceId, day } }) => ({
    locationId,
    serviceId,
    day,
  }),
  loader: async ({ params, deps: { locationId, serviceId, day } }) => {
    const availableTimes = await fetchHours(
      params.merchantName,
      locationId,
      serviceId,
      day
    );
    const closedDays = await fetchClosedDays(params.merchantName);

    return {
      availableTimes,
      closedDays,
    };
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function SelectDateTime() {
  const { merchantName } = Route.useParams();
  const { locationId, serviceId, day } = Route.useSearch();
  const loaderData = Route.useLoaderData();
  const router = useRouter();
  const [selectedHour, setSelectedHour] = useState();
  const [serverError, setServerError] = useState();
  const [isLoading, setIsLoading] = useState(false);
  const [availableTimes, setAvailableTimes] = useState({
    morning: [],
    afternoon: [],
  });
  const [customerNote, setCustomerNote] = useState("");
  const closedDays = loaderData.closedDays;

  useEffect(() => {
    if (loaderData.availableTimes) {
      setAvailableTimes(loaderData.availableTimes);
    }
  }, [loaderData]);

  async function onSubmitHandler(e) {
    e.preventDefault();
    const date = new Date(day);

    const [hours, minutes] = selectedHour.split(":").map(Number);
    date.setHours(hours, minutes, 0, 0);
    const timeStamp = date.toISOString();

    setIsLoading(true);

    try {
      const response = await fetch("/api/v1/appointments/new", {
        method: "POST",
        headers: {
          "Content-type": "application/json; charset=UTF-8",
        },
        body: JSON.stringify({
          merchant_name: merchantName,
          service_id: serviceId,
          location_id: locationId,
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
    setSelectedHour();
    router.navigate({
      search: (prev) => ({ ...prev, day: formatToDateString(date) }),
    });
  }

  function selectedHourHandler(e) {
    setSelectedHour(e.target.value);
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
            <div className="flex flex-col gap-6 md:w-1/2">
              <p className="text-xl sm:py-5">Pick a date</p>
              <div className="flex items-center justify-center self-center bg-white shadow-lg dark:bg-neutral-950">
                <SmallCalendar
                  value={day}
                  onSelect={dayChangeHandler}
                  disabled={[{ dayOfWeek: closedDays }, { before: new Date() }]}
                  disabledTodayStyling={true}
                />
              </div>
              <hr className="border-gray-500" />
              <p className="text-xl sm:py-5">Pick a Time</p>
              <div className="flex flex-col gap-3">
                <p className="text-lg font-bold">Morning</p>
                <AvailableTimeSection
                  availableTimes={availableTimes.morning}
                  timeSection="morning"
                  selectedHour={selectedHour}
                  clickedHour={selectedHourHandler}
                />
                <p className="text-lg font-bold">Afternoon</p>
                <AvailableTimeSection
                  availableTimes={availableTimes.afternoon}
                  timeSection="afternoon"
                  selectedHour={selectedHour}
                  clickedHour={selectedHourHandler}
                />
              </div>
            </div>
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
                    <p>{serviceId}</p>
                  </div>
                  <div>
                    <p>Location:</p>
                    <p>{locationId}</p>
                  </div>
                  <div className={`${day ? "" : "invisible"}`}>
                    <p>Date:</p>
                    <p>{day}</p>
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
                labelText="Add a note to your appointment"
                required={false}
                placeholder="Add your note here..."
                value={customerNote}
                inputData={(data) => setCustomerNote(data.value)}
              />

              <div className="bg-hvr_gray dark:bg-layer_bg fixed right-0 bottom-0 left-0 px-8 py-3 md:static md:bg-transparent md:px-0 md:pt-10 dark:md:bg-transparent">
                <Button
                  variant="primary"
                  type="submit"
                  disabled={day && selectedHour ? false : true}
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
