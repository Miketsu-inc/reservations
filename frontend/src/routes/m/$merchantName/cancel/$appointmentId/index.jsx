import Button from "@components/Button";
import ServerError from "@components/ServerError";
import ClockIcon from "@icons/ClockIcon";
import MapPinIcon from "@icons/MapPinIcon";
import WarningIcon from "@icons/WarningIcon";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";

async function fetchAppointmentInfo(appointmentId) {
  const response = await fetch(`/api/v1/appointments/public/${appointmentId}`, {
    method: "GET",
  });

  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  } else {
    return result.data;
  }
}

function formatDate(dateString) {
  const date = new Date(dateString);
  return {
    day: date.getDate(),
    month: date.toLocaleDateString("default", { month: "short" }),
    weekday: date.toLocaleDateString("default", { weekday: "short" }),
  };
}

function formatTime(dateString) {
  const date = new Date(dateString);
  return date.toLocaleTimeString("default", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
}

export const Route = createFileRoute("/m/$merchantName/cancel/$appointmentId/")(
  {
    component: CancelPage,
    loader: async ({ params }) => {
      return fetchAppointmentInfo(params.appointmentId);
    },
    errorComponent: ({ error }) => {
      return <ServerError error={error.message} />;
    },
  }
);

function CancelPage() {
  const loaderData = Route.useLoaderData();
  const [cancelling, setCancelling] = useState(false);
  const [serverError, setServerError] = useState("");
  const router = useRouter();
  const dateInfo = formatDate(loaderData.from_date);
  const params = Route.useParams();

  const now = new Date();
  const fromDate = new Date(loaderData.from_date);
  const isInPast = fromDate <= now;
  const alreadyCancelled =
    loaderData.cancelled_by_user || loaderData.cancelled_by_merchant;
  const canBeCancelled = !isInPast && !alreadyCancelled;

  async function handleCancel() {
    setCancelling(true);
    try {
      const response = await fetch(
        `/api/v1/appointments/public/${params.appointmentId}`,
        {
          method: "DELETE",
          headers: {
            "Content-type": "application/json; charset=UTF-8",
          },
          body: JSON.stringify({
            appointment_id: Number(params.appointmentId),
            merchant_name: params.merchantName,
          }),
        }
      );

      if (!response.ok) {
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
      setCancelling(false);
    }
  }

  return (
    <div className="mx-auto flex w-full flex-col items-center px-3 sm:max-w-lg sm:gap-4 sm:p-0">
      <div className={`${!canBeCancelled ? "pt-6" : "py-6"} text-center`}>
        <h1 className="text-2xl font-bold">Cancel Appointment</h1>
        <p className="mt-1 text-sm text-gray-500">
          Review your appointment details below
        </p>
      </div>
      {!canBeCancelled && (
        <div className="my-4 flex w-full items-start justify-start gap-3 rounded-lg border border-yellow-800 bg-yellow-400/25 px-3 py-3 text-yellow-900 sm:mt-0 sm:mb-2 dark:border-yellow-800 dark:bg-yellow-700/15 dark:text-yellow-500">
          <WarningIcon styles="h-5 w-5 shrink-0" />
          <span className="text-sm">
            {loaderData.cancelled_by_user || loaderData.cancelled_by_merchant
              ? "This appointment has already been cancelled."
              : "You cannot cancel this appointment because it has already passed."}
          </span>
        </div>
      )}
      <div className="bg-layer_bg border-border_color w-full rounded-t-lg rounded-b-none border border-b-0 p-4 shadow-none sm:rounded-lg sm:border sm:shadow-sm">
        <div className="flex items-center gap-4">
          <div className="text-center">
            <div className="text-3xl font-bold">{dateInfo.day}</div>
            <div className="text-sm font-medium">{dateInfo.month}</div>
            <div className="text-xs opacity-60">{dateInfo.weekday}</div>
          </div>

          <div className="flex-1">
            <h2 className="text-lg font-bold">{loaderData.service_name}</h2>
            <p className="mb-1 text-sm text-gray-600 dark:text-gray-400">
              {loaderData.merchant_name}
            </p>
            <div className="flex items-center gap-2">
              <ClockIcon styles="w-4 h-4 dark:fill-gray-400 fill-gray-500" />
              <span className="font-medium text-gray-500 dark:text-gray-400">
                {formatTime(loaderData.from_date)} -{" "}
                {formatTime(loaderData.to_date)}
              </span>
            </div>
          </div>
          <div className="from-secondary to-primary h-16 w-16 rounded-xl bg-linear-to-br sm:h-20 sm:w-20"></div>
        </div>
      </div>
      <div
        className={`${loaderData.price && "sm:grid-cols-2"} grid w-full grid-cols-1 gap-0 sm:gap-4 sm:pb-2`}
      >
        <div className="border-border_color bg-layer_bg rounded-none border border-y-0 p-4 pt-2 shadow-none sm:mb-0 sm:rounded-lg sm:border sm:p-4 sm:shadow-sm">
          <div className="mb-1 flex items-center gap-2">
            <MapPinIcon styles="w-4 h-4 dark:text-blue-400 text-blue-600" />
            <span className="text-text_color text-xs font-semibold">
              LOCATION
            </span>
          </div>
          <p
            className={
              "mt-2 text-sm font-medium text-gray-500 dark:text-gray-400"
            }
          >
            {loaderData.short_location}
          </p>
        </div>
        {loaderData.price && (
          <div className="border-border_color bg-layer_bg rounded-none border border-y-0 p-4 pt-2 shadow-none sm:mb-0 sm:rounded-lg sm:border sm:p-4 sm:shadow-sm">
            <div className="text-text_color mb-1 flex items-center gap-2 text-xs font-semibold">
              <span className="text-green-600 dark:text-green-400">$</span>
              PRICE
            </div>

            <p className="mt-1 font-medium text-gray-500 dark:text-gray-400">
              {loaderData.price} {loaderData.price_note}
            </p>
          </div>
        )}
      </div>
      <div className="border-border_color bg-layer_bg flex w-full flex-col items-center gap-5 rounded-b-lg border border-t-0 py-4 sm:border-none sm:bg-transparent">
        <div className="px-5 text-center text-sm text-gray-500">
          The cancellation is inmediate and cannot be undone.
        </div>
        <Button
          buttonText="Cancel appointment"
          onClick={handleCancel}
          isLoading={cancelling}
          disabled={!canBeCancelled}
          variant="danger"
          styles="px-2 py-1 w-min truncate"
        />
      </div>
      {serverError && <ServerError error={serverError} />}
    </div>
  );
}
