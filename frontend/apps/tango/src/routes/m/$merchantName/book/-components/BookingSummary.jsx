import {
  ArrowRight02Icon,
  Calendar,
  Clock,
  Note01Icon,
  StarIcon,
} from "@hugeicons/core-free-icons";
import { Button, Card, Icon } from "@reservations/components";
import {
  formatDuration,
  formatTimeRange,
  getDisplayPrice,
} from "@reservations/lib";

const formatDate = (dateString) => {
  const date = new Date(`${dateString}`);

  const formattedDate = new Intl.DateTimeFormat("en-US", {
    month: "long",
    day: "numeric",
    weekday: "long",
  }).format(date);

  return formattedDate;
};

export default function BookingSummary({
  fetchedSummary,
  selectedSummary,
  isLoading,
  onContinue,
  canContinue,
  isSubmitting,
  currentStep,
}) {
  const duration =
    fetchedSummary?.total_duration || selectedSummary.service?.total_duration;

  if (isLoading) {
    return (
      <Card
        styles="sticky top-24 flex h-[calc(100vh-8rem)] w-full flex-col
          justify-between p-8 animate-pulse"
      >
        <div>
          <div className="flex items-start gap-3">
            <div
              className="size-18 shrink-0 rounded-lg bg-gray-200
                dark:bg-gray-800"
            ></div>
            <div className="flex w-full flex-col gap-2 py-1">
              <div className="h-5 w-3/4 rounded bg-gray-200 dark:bg-gray-800"></div>
              <div className="h-4 w-1/2 rounded bg-gray-100 dark:bg-gray-800"></div>
              <div className="h-3 w-full rounded bg-gray-100 dark:bg-gray-800"></div>
            </div>
          </div>
          <hr className="border-border_color my-5" />
          <div className="flex items-start justify-between">
            <div className="flex w-full flex-col gap-2">
              <div className="h-5 w-2/3 rounded bg-gray-200 dark:bg-gray-800"></div>
              <div className="h-4 w-1/3 rounded bg-gray-100 dark:bg-gray-800"></div>
              <div className="h-4 w-1/2 rounded bg-gray-100 dark:bg-gray-800"></div>
            </div>
            <div
              className="h-5 w-16 shrink-0 rounded bg-gray-200 dark:bg-gray-800"
            ></div>
          </div>
        </div>
        <div
          className="mt-8 h-10 w-full rounded-md bg-gray-200 dark:bg-gray-800"
        ></div>
      </Card>
    );
  }

  return (
    <Card
      styles="flex flex-col justify-between max-w-110 p-8 sticky top-24
        h-[calc(100vh-8rem)]"
    >
      <div>
        <div className="flex items-start gap-3">
          <div className="flex size-18 shrink-0 overflow-hidden rounded-lg">
            <img
              className="size-full object-cover"
              src="https://dummyimage.com/70x70/d156c3/000000.jpg"
              alt="service photo"
            />
          </div>
          <div className="flex flex-col gap-0.5">
            <h2 className="text-lg font-medium">
              {fetchedSummary?.merchant_name}
            </h2>
            <div className="flex items-center gap-1 text-sm">
              <Icon
                icon={StarIcon}
                styles="size-5 text-yellow-500 fill-yellow-500"
              />
              <span className="font-medium">4.6</span>
              <span className="ml-1 cursor-pointer text-gray-500">
                (9 reviews)
              </span>
            </div>
            <p className="dark:text-text_color/80 line-clamp-1 w-full text-sm">
              {fetchedSummary?.formatted_location}
            </p>
          </div>
        </div>

        <hr className="border-border_color my-5" />
        {selectedSummary.time?.date && selectedSummary.time?.time && (
          <>
            <div className="flex justify-between gap-2">
              <div className="flex items-center gap-2">
                <Icon icon={Calendar} styles="size-5" />
                <span className="">
                  {formatDate(selectedSummary.time.date)}
                </span>
              </div>
              <div className="flex items-center gap-2">
                <Icon icon={Clock} styles="size-5" />
                <span className="">
                  {formatTimeRange(selectedSummary.time.time, duration)}
                </span>
              </div>
            </div>
            <hr className="border-border_color my-5" />
          </>
        )}

        {fetchedSummary?.service_name || selectedSummary?.service ? (
          <div className="flex justify-between">
            <div className="flex flex-col gap-1">
              <span className="text-[17px] font-medium">
                {fetchedSummary.service_name || selectedSummary.service.name}
              </span>
              <span className="text-gray-600 dark:text-gray-300">
                {fetchedSummary?.service_name
                  ? formatDuration(fetchedSummary.total_duration)
                  : formatDuration(selectedSummary.service.total_duration)}
              </span>
              {(fetchedSummary?.employee_first_name ||
                selectedSummary?.employee) && (
                <span className="text-gray-600 dark:text-gray-300">
                  With:{" "}
                  {fetchedSummary?.employee_first_name ||
                    selectedSummary.employee?.first_name}
                </span>
              )}
            </div>
            <span className="font-medium">
              {fetchedSummary?.price_type
                ? getDisplayPrice(
                    fetchedSummary?.price,
                    fetchedSummary?.price_type
                  )
                : getDisplayPrice(
                    selectedSummary.service?.price,
                    selectedSummary.service?.price_type
                  )}
            </span>
          </div>
        ) : (
          <p className="text-gray-400">Select a service to continue...</p>
        )}
        {selectedSummary.time?.customer_note && (
          <>
            <hr className="border-border_color my-5" />
            <div
              className="flex items-start justify-start gap-2 text-gray-600
                dark:text-gray-300"
            >
              <Icon icon={Note01Icon} styles="size-5 shrink-0 mt-0.5" />
              <p className="line-clamp-3 font-medium">
                Your Note:{" "}
                <span className="text-sm font-normal">
                  {selectedSummary.time?.customer_note}
                </span>
              </p>
            </div>
          </>
        )}
      </div>

      <Button
        styles="w-full py-2"
        buttonText={currentStep === "time" ? "Book" : "Continue"}
        variant="primary"
        disabled={!canContinue}
        onClick={onContinue}
        isLoading={isSubmitting}
      >
        <Icon icon={ArrowRight02Icon} styles="size-5" />
      </Button>
    </Card>
  );
}
