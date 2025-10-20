import { ClockIcon, MapPinIcon, WarningIcon } from "@reservations/assets";
import { Button, Loading, ServerError } from "@reservations/components";
import { preferencesQueryOptions, timeStringFromDate } from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";

async function fetchBookingInfo(bookingId) {
  const response = await fetch(`/api/v1/bookings/public/${bookingId}`, {
    method: "GET",
  });

  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  } else {
    return result.data;
  }
}

function publicBookingInfoQueryOptions(bookingId) {
  return queryOptions({
    queryKey: ["public-booking-info", bookingId],
    queryFn: () => fetchBookingInfo(bookingId),
  });
}

function formatDate(dateString) {
  const date = new Date(dateString);
  return {
    day: date.getDate(),
    month: date.toLocaleDateString("default", { month: "short" }),
    weekday: date.toLocaleDateString("default", { weekday: "short" }),
  };
}

export const Route = createFileRoute("/m/$merchantName/cancel/$bookingId/")({
  component: CancelPage,
  loader: async ({ params: { bookingId }, context: { queryClient } }) => {
    await queryClient.ensureQueryData(publicBookingInfoQueryOptions(bookingId));
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

const cancelDeadlineLabels = {
  0: "",
  30: "30 minutes",
  60: "1 hour",
  120: "2 hours",
  180: "3 hours",
  240: "4 hours",
  300: "5 hours",
  360: "6 hours",
  720: "12 hours",
  1440: "1 day",
  2880: "2 days",
  5760: "3 days",
  20160: "2 weeks",
};

function CancelPage() {
  const [cancelling, setCancelling] = useState(false);
  const [serverError, setServerError] = useState("");
  const router = useRouter();
  const { merchantName, bookingId } = Route.useParams();

  const {
    data: bookingData,
    isLoading,
    isError,
    error,
  } = useQuery(publicBookingInfoQueryOptions(bookingId));

  const { data: preferences } = useQuery(preferencesQueryOptions());

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <ServerError error={error} />;
  }

  const dateInfo = formatDate(bookingData.from_date);
  const fromDate = new Date(bookingData.from_date);
  const cancelDeadline = new Date(
    fromDate.getTime() - bookingData.cancel_deadline * 60000
  ); // getTime returns in milisecond

  const now = new Date();
  const alreadyCancelled = bookingData.is_cancelled;
  const canBeCancelled = !alreadyCancelled && now < cancelDeadline;

  let cancelMessage = "";
  if (alreadyCancelled) {
    cancelMessage = "This booking has already been cancelled.";
  } else if (bookingData.cancel_deadline === 0 || fromDate <= now) {
    cancelMessage =
      "You cannot cancel this booking because it has already passed.";
  } else {
    const deadlineLabel = cancelDeadlineLabels[bookingData.cancel_deadline];
    cancelMessage = `You cannot cancel this booking less than ${deadlineLabel} before it starts.`;
  }

  async function handleCancel() {
    setCancelling(true);
    try {
      const response = await fetch(`/api/v1/bookings/public/${bookingId}`, {
        method: "DELETE",
        headers: {
          "Content-type": "application/json; charset=UTF-8",
        },
        body: JSON.stringify({
          booking_id: Number(bookingId),
          merchant_name: merchantName,
        }),
      });

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
    <div
      className="mx-auto flex w-full flex-col items-center px-3 sm:max-w-lg
        sm:gap-4 sm:p-0"
    >
      <div className={`${!canBeCancelled ? "pt-6" : "py-6"} text-center`}>
        <h1 className="text-2xl font-bold">Cancel Booking</h1>
        <p className="mt-1 text-sm text-gray-500">
          Review your booking details below
        </p>
      </div>
      {!canBeCancelled && (
        <div
          className="my-4 flex w-full items-start justify-start gap-3 rounded-lg
            border border-yellow-800 bg-yellow-400/25 px-3 py-3 text-yellow-900
            sm:mt-0 sm:mb-2 dark:border-yellow-800 dark:bg-yellow-700/15
            dark:text-yellow-500"
        >
          <WarningIcon styles="h-5 w-5 shrink-0" />

          <span className="text-sm">{cancelMessage}</span>
        </div>
      )}
      <div
        className="bg-layer_bg border-border_color w-full rounded-t-lg
          rounded-b-none border border-b-0 p-4 shadow-none sm:rounded-lg
          sm:border sm:shadow-sm"
      >
        <div className="flex items-center gap-4">
          <div className="text-center">
            <div className="text-3xl font-bold">{dateInfo.day}</div>
            <div className="text-sm font-medium">{dateInfo.month}</div>
            <div className="text-xs opacity-60">{dateInfo.weekday}</div>
          </div>

          <div className="flex-1">
            <h2 className="text-lg font-bold">{bookingData.service_name}</h2>
            <p className="mb-1 text-sm text-gray-600 dark:text-gray-400">
              {bookingData.merchant_name}
            </p>
            <div className="flex items-center gap-2">
              <ClockIcon styles="w-4 h-4 dark:fill-gray-400 fill-gray-500" />
              <span className="font-medium text-gray-500 dark:text-gray-400">
                {`${timeStringFromDate(
                  new Date(bookingData.from_date),
                  preferences?.time_format
                )} - ${timeStringFromDate(
                  new Date(bookingData.to_date),
                  preferences?.time_format
                )}`}
              </span>
            </div>
          </div>
          <div
            className="from-secondary to-primary h-16 w-16 rounded-xl
              bg-linear-to-br sm:h-20 sm:w-20"
          ></div>
        </div>
      </div>
      <div
        className={`${bookingData.price && "sm:grid-cols-2"} grid w-full
          grid-cols-1 gap-0 sm:gap-4 sm:pb-2`}
      >
        <div
          className="border-border_color bg-layer_bg rounded-none border
            border-y-0 p-4 pt-2 shadow-none sm:mb-0 sm:rounded-lg sm:border
            sm:p-4 sm:shadow-sm"
        >
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
            {bookingData.short_location}
          </p>
        </div>
        {bookingData.price && (
          <div
            className="border-border_color bg-layer_bg rounded-none border
              border-y-0 p-4 pt-2 shadow-none sm:mb-0 sm:rounded-lg sm:border
              sm:p-4 sm:shadow-sm"
          >
            <div
              className="text-text_color mb-1 flex items-center gap-2 text-xs
                font-semibold"
            >
              <span className="text-green-600 dark:text-green-400">$</span>
              PRICE
            </div>

            <p className="mt-1 font-medium text-gray-500 dark:text-gray-400">
              {bookingData.price} {bookingData.price_note}
            </p>
          </div>
        )}
      </div>
      <div
        className="border-border_color bg-layer_bg flex w-full flex-col
          items-center gap-5 rounded-b-lg border border-t-0 py-4 sm:border-none
          sm:bg-transparent"
      >
        <div className="px-5 text-center text-sm text-gray-500">
          The cancellation is inmediate and cannot be undone.
        </div>
        <Button
          buttonText="Cancel booking"
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
