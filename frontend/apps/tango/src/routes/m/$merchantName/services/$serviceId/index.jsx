import { CalendarIcon, ClockIcon, MapPinIcon } from "@reservations/assets";
import { Button, Card, Loading, ServerError } from "@reservations/components";
import {
  businessHoursQueryOptions,
  formatDuration,
  invalidateLocalStorageAuth,
  useWindowSize,
} from "@reservations/lib";
import {
  queryOptions,
  useQueries,
  useSuspenseQuery,
} from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import MapboxMap from "../../-components/MapboxMap";
import DropDownBusinessHours from "./-components/DropDownBusinessHours";
import PhaseItem from "./-components/PhaseItem";

async function fetchServiceDetails(merchantName, serviceId, locationId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantName}/${locationId}/services/public/${serviceId}`,
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

function publicServiceDetailsQueryOptions(merchantName, serviceId, locationId) {
  return queryOptions({
    queryKey: ["public-service-details", merchantName, serviceId, locationId],
    queryFn: () => fetchServiceDetails(merchantName, serviceId, locationId),
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
      publicServiceDetailsQueryOptions(merchantName, serviceId, locationId)
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
  const windowSize = useWindowSize();
  const isWindowSmall =
    windowSize === "sm" || windowSize === "md" || windowSize === "lg";

  const businessHours = useSuspenseQuery(businessHoursQueryOptions());

  const queryResults = useQueries({
    queries: [
      publicServiceDetailsQueryOptions(merchantName, serviceId, locationId),
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
    <div className="flex justify-center">
      <div
        className="flex w-full flex-col justify-center p-3 lg:flex-row lg:gap-4
          lg:p-6"
      >
        <Card styles="flex-1 lg:max-w-2xl min-h-fit lg:p-6 p-4">
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
              <div className="flex flex-col gap-5">
                <span className="text-3xl font-bold">
                  {serviceDetails.name}
                </span>
                <div
                  className={`flex
                    ${windowSize === "sm" ? "flex-col gap-3" : "gap-6"}`}
                >
                  <div className="flex items-center justify-start gap-2">
                    <ClockIcon styles="size-4 fill-text_color" />
                    {formatDuration(serviceDetails.total_duration)}
                  </div>
                  <span>
                    {serviceDetails.price}
                    {serviceDetails?.price_note}
                  </span>
                </div>
              </div>
            </div>
            {serviceDetails?.description && (
              <div className="flex flex-col gap-2">
                <div className="text-lg font-semibold">Description</div>
                <p>{serviceDetails.description}</p>
              </div>
            )}
            {isWindowSmall && (
              <NextAvailable
                hasAvailableSlot={hasAvailableSlot}
                nextAvailable={nextAvailable}
                locationId={locationId}
                serviceDetails={serviceDetails}
                businessHours={businessHours.data}
                isWindowSmall={isWindowSmall}
              />
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
            <div className="flex flex-col gap-3">
              <h2 className="flex items-center gap-2 text-lg font-semibold">
                <MapPinIcon styles="size-5" />
                Location
              </h2>
              <div className="flex flex-col gap-2">
                <p>{serviceDetails.formatted_location}</p>
                <MapboxMap
                  coordinates={[
                    serviceDetails.geo_point.longitude,
                    serviceDetails.geo_point.latitude,
                  ]}
                  minHeight={300}
                  zoom={13}
                />
              </div>
            </div>
          </div>
        </Card>
        {!isWindowSmall && (
          <div className="sticky top-6 w-96 shrink-0 self-start">
            <NextAvailable
              hasAvailableSlot={hasAvailableSlot}
              nextAvailable={nextAvailable}
              locationId={locationId}
              serviceDetails={serviceDetails}
              businessHours={businessHours.data}
              isWindowSmall={isWindowSmall}
            />
          </div>
        )}
      </div>
    </div>
  );
}

function NextAvailable({
  hasAvailableSlot,
  nextAvailable,
  locationId,
  serviceDetails,
  businessHours,
  isWindowSmall,
}) {
  return (
    <div
      className="lg:border-border_color bg-layer_bg flex w-full flex-col
        items-center gap-5 rounded-xl md:shadow-sm lg:gap-6 lg:border lg:p-6"
    >
      <div className="flex w-full items-center gap-2">
        <CalendarIcon styles="size-6 mb-0.5 text-text_color" />
        <h2 className="text-text_color text-lg font-semibold">
          Next Available
        </h2>
      </div>

      <div
        className={`flex w-full flex-col gap-4 rounded-md border-2 p-4 ${
          hasAvailableSlot
            ? `border-green-200 bg-green-100/40 dark:border-green-800/50
              dark:bg-green-900/10`
            : `border-gray-200 bg-gray-100 dark:border-gray-800
              dark:bg-gray-900/20`
          }`}
      >
        {hasAvailableSlot ? (
          <>
            <div className="flex flex-col gap-3">
              <div className="flex items-baseline gap-2">
                <span
                  className="text-2xl font-bold text-green-700
                    dark:text-green-400"
                >
                  {nextAvailable.time}
                </span>
              </div>
              <div
                className="flex items-center gap-2 text-base text-gray-700
                  dark:text-gray-300"
              >
                <span className="font-medium">
                  {formatDate(nextAvailable.date)}
                </span>
              </div>
            </div>
          </>
        ) : (
          <div className="flex flex-col gap-3 text-center">
            <p className="text-gray-600 dark:text-gray-400">
              No available slots for the next 3 months. Check the calendar for
              future availability.
            </p>
          </div>
        )}
      </div>

      <Link
        from={Route.fullPath}
        to="../../booking"
        search={{
          locationId: locationId,
          serviceId: serviceDetails.id,
        }}
        className="w-full"
      >
        <Button
          variant="primary"
          styles="w-full py-3 px-4 font-semibold"
          buttonText={hasAvailableSlot ? "Book Now" : "Check Calendar"}
        />
      </Link>
      {!isWindowSmall && (
        <>
          <hr className="border-border_color w-full border-b" />
          <DropDownBusinessHours hoursData={businessHours} />
        </>
      )}
    </div>
  );
}
