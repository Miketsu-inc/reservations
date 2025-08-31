import Button from "@components/Button";
import Card from "@components/Card";
import Loading from "@components/Loading";
import ServerError from "@components/ServerError";
import CalendarIcon from "@icons/CalendarIcon";
import ClockIcon from "@icons/ClockIcon";
import MapPinIcon from "@icons/MapPinIcon";
import { formatDuration } from "@lib/datetime";
import { invalidateLocalStorageAuth } from "@lib/lib";
import { queryOptions, useQueries } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import PhaseItem from "./-components/PhaseItem";

async function fetchServiceDetails(merchantName, serviceId) {
  const response = await fetch(
    `/api/v1/merchants/services/public/${serviceId}/${merchantName}`,
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

function publicServiceDetailsQueryOptions(merchantName, serviceId) {
  return queryOptions({
    queryKey: ["public-service-details", merchantName, serviceId],
    queryFn: () => fetchServiceDetails(merchantName, serviceId),
  });
}

async function fetchNextAvailable(merchantName, serviceId, locationId) {
  const response = await fetch(
    `/api/v1/merchants/next-available?name=${merchantName}&locationId=${locationId}&serviceId=${serviceId}`,
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

function nextAvailableQueryOptions(merchantName, serviceId, locationId) {
  return queryOptions({
    queryKey: ["next-available", merchantName, serviceId, locationId],
    queryFn: () => fetchNextAvailable(merchantName, serviceId, locationId),
  });
}

export const Route = createFileRoute("/m/$merchantName/services/$serviceId/")({
  component: ServiceDetailsPage,
  loaderDeps: ({ search: { locationId } }) => ({
    locationId,
  }),
  loader: async ({
    params: { merchantName, serviceId },
    deps: { locationId },
    context: { queryClient },
  }) => {
    await queryClient.ensureQueryData(
      publicServiceDetailsQueryOptions(merchantName, serviceId)
    );
    await queryClient.ensureQueryData(
      nextAvailableQueryOptions(merchantName, serviceId, locationId)
    );
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

const formatDate = (dateString) => {
  if (!dateString) return "";
  const date = new Date(dateString);
  return date.toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
};

function ServiceDetailsPage() {
  const { locationId } = Route.useSearch({ from: Route.id });
  const { merchantName, serviceId } = Route.useParams({ from: Route.id });

  const queryResults = useQueries({
    queries: [
      publicServiceDetailsQueryOptions(merchantName, serviceId),
      nextAvailableQueryOptions(merchantName, serviceId, locationId),
    ],
  });

  if (queryResults.some((r) => r.isLoading)) {
    return <Loading />;
  }

  if (queryResults.some((r) => r.isError)) {
    const error = queryResults.find((r) => r.error);
    return <ServerError error={error} />;
  }

  const serviceDetails = queryResults[0].data;
  const nextAvailable = queryResults[1].data;

  const hasAvailableSlot = nextAvailable.date && nextAvailable.time;

  const sortedPhases = [...serviceDetails.phases].sort(
    (a, b) => (a.sequence || 0) - (b.sequence || 0)
  );

  return (
    <div className="p-4">
      <Card styles="md:mx-auto min-h-screen max-w-7xl md:p-6 p-4">
        <div className="grid w-full grid-cols-1 gap-8 lg:grid-cols-2 lg:gap-16">
          <div className="flex w-full flex-col gap-8">
            <div className="flex items-center gap-6">
              <div
                className="shrink-0 overflow-hidden rounded-lg xl:size-[120px]"
              >
                <img
                  className="size-full object-cover"
                  src="https://dummyimage.com/120x120/d156c3/000000.jpg"
                  alt="service photo"
                ></img>
              </div>
              <div className="flex flex-col gap-4">
                <span className="text-2xl font-bold">
                  {serviceDetails.name}
                </span>
                <div className="flex flex-col gap-2">
                  {serviceDetails.price.number} {serviceDetails.price.currency}{" "}
                  {serviceDetails?.price_note}
                  <div className="flex items-center justify-start gap-3">
                    <ClockIcon styles="size-4 fill-text_color" />
                    {formatDuration(serviceDetails.total_duration)}
                  </div>
                </div>
              </div>
            </div>
            {serviceDetails?.description && (
              <div className="flex flex-col gap-2">
                <div className="text-lg font-semibold">Description</div>
                <p>{serviceDetails.description}</p>
              </div>
            )}

            {sortedPhases.length > 1 && (
              <div className="flex h-min w-full flex-col gap-5">
                <div className="text-lg font-semibold">
                  Phases of the service
                </div>
                <div className="flex flex-col gap-9">
                  {sortedPhases.map((phase, index) => (
                    <PhaseItem
                      key={phase.id}
                      phase={phase}
                      isLast={index === sortedPhases.length - 1}
                    />
                  ))}
                </div>
              </div>
            )}
          </div>
          <div className="flex flex-col gap-6 md:mt-3 md:gap-8">
            <div
              className="border-primary bg-primary/20 rounded-md border-2
                border-dashed p-4"
            >
              <div
                className="flex flex-col gap-6 sm:flex-row sm:items-center
                  sm:justify-between"
              >
                <div className="flex flex-col gap-2 sm:gap-4">
                  <div className="flex items-center gap-2">
                    <CalendarIcon styles="size-5 mb-0.5 text-text_color" />
                    <span
                      className="text-text_color text-sm font-medium
                        tracking-wide"
                    >
                      NEXT AVAILABLE
                    </span>
                  </div>
                  {hasAvailableSlot ? (
                    <div
                      className="text-text_color flex items-center gap-4
                        text-base font-semibold"
                    >
                      <div className="flex items-center gap-2">
                        <span>{formatDate(nextAvailable.date)}</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <ClockIcon styles="size-4 stroke-text_color" />
                        <span>{nextAvailable.time}</span>
                      </div>
                    </div>
                  ) : (
                    <p className="text-text_color text-base font-semibold">
                      Fully booked for the next 3 months
                    </p>
                  )}
                </div>
                <Link
                  from={Route.fullPath}
                  to="../../booking"
                  search={{
                    locationId: locationId,
                    serviceId: serviceDetails.id,
                    day:
                      nextAvailable.date ||
                      new Date().toISOString().split("T")[0],
                  }}
                >
                  <Button
                    variant="primary"
                    styles="py-2 px-4 w-full sm:w-fit text-nowrap"
                    buttonText={hasAvailableSlot ? "Reserve" : "Check Calendar"}
                  />
                </Link>
              </div>
            </div>
            <div className="flex h-64 flex-col gap-3">
              <h2 className="flex items-center gap-2 text-lg font-semibold">
                <MapPinIcon styles="size-5" />
                Location
              </h2>
              <div className="border-text_color h-full rounded-md border"></div>
            </div>
          </div>
        </div>
      </Card>
    </div>
  );
}
