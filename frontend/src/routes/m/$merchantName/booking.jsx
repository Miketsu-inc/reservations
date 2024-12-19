import Button from "@components/Button";
import ServerError from "@components/ServerError";
import BackArrowIcon from "@icons/BackArrowIcon";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { createFileRoute, Link, useRouter } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { DayPicker } from "react-day-picker";
import "react-day-picker/style.css";
import AvailableTimeSection from "./-components/AvailableTimeSection";

async function fetchHours(merchantName, locationId, serviceId, day) {
  const response = await fetch(
    `/api/v1/merchants/times?name=${merchantName}&locationId=${locationId}&serviceId=${serviceId}&day=${day}`,
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
    invalidateLocalSotrageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

export const Route = createFileRoute("/m/$merchantName/booking")({
  component: SelectDateTime,
  loaderDeps: ({ search: { locationId, serviceId, day } }) => ({
    locationId,
    serviceId,
    day,
  }),
  loader: ({ params, deps: { locationId, serviceId, day } }) => {
    return fetchHours(params.merchantName, locationId, serviceId, day);
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

const defaultAvailableTimes = {
  morning: [],
  afternoon: [],
};

function SelectDateTime() {
  const { merchantName } = Route.useParams();
  const { locationId, serviceId, day } = Route.useSearch();
  const loaderData = Route.useLoaderData();
  const router = useRouter();
  const [selectedHour, setSelectedHour] = useState();
  const [serverError, setServerError] = useState();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [availableTimes, setAvailableTimes] = useState(defaultAvailableTimes);

  useEffect(() => {
    if (loaderData) {
      setAvailableTimes(loaderData);
    }
  }, [loaderData]);

  async function onSubmitHandler(e) {
    e.preventDefault();
    const date = new Date(day);

    const [hours, minutes] = selectedHour.split(":").map(Number);
    date.setUTCHours(hours, minutes, 0, 0);
    const timeStamp = date.toISOString();

    setIsSubmitting(true);

    try {
      const response = await fetch("/api/v1/appointments", {
        method: "POST",
        headers: {
          "Content-type": "application/json; charset=UTF-8",
        },
        body: JSON.stringify({
          merchant_name: merchantName,
          service_id: serviceId,
          location_id: locationId,
          timeStamp: timeStamp,
        }),
      });

      if (!response.ok) {
        invalidateLocalSotrageAuth(response.status);

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
      }
    } catch (err) {
      setServerError(err.message);
    } finally {
      setIsSubmitting(false);
    }
  }

  function dayChangeHandler(date) {
    setSelectedHour();
    router.navigate({
      search: (prev) => ({ ...prev, day: date.toISOString().split("T")[0] }),
    });
  }

  function selectedHourHandler(e) {
    setSelectedHour(e.target.value);
  }

  return (
    <div className="mx-auto min-h-screen max-w-screen-xl bg-layer_bg px-10">
      <div className="py-5">
        <Link from={Route.fullPath} to="..">
          <button className="inline-flex gap-1 hover:underline">
            <BackArrowIcon />
            Back
          </button>
        </Link>
        <ServerError error={serverError} />
        <form method="POST" onSubmit={onSubmitHandler}>
          <div className="flex flex-col pt-5 md:flex-row md:gap-10 lg:pt-10">
            <div className="flex flex-col gap-6 md:w-1/2">
              <p className="py-5 text-xl">Pick a date</p>
              <div className="self-center md:self-auto">
                <DayPicker
                  mode="single"
                  timeZone="UTC"
                  selected={day}
                  onSelect={dayChangeHandler}
                />
              </div>
              <hr className="border-gray-500" />
              <p className="py-5 text-xl">Pick a Time</p>
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
            <div className="pt-8 md:flex md:w-1/2 md:flex-col md:pt-0">
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
              <div className="md:pt-28">
                <Button
                  type="submit"
                  disabled={day && selectedHour ? false : true}
                  isLoading={isSubmitting}
                  buttonText="Reserve"
                  styles="w-full text-white dark:bg-transparent dark:border-2 border-secondary
                    dark:text-secondary dark:hover:border-hvr_secondary
                    dark:hover:text-hvr_secondary font-semibold border-primary hover:bg-hvr_primary
                    dark:focus:outline-none dark:focus:border-hvr_secondary
                    dark:focus:text-hvr_secondary"
                ></Button>
              </div>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}
