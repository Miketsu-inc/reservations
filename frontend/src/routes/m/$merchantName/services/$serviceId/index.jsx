import Button from "@components/Button";
import Card from "@components/Card";
import ServerError from "@components/ServerError";
import CalendarIcon from "@icons/CalendarIcon";
import ClockIcon from "@icons/ClockIcon";
import MapPinIcon from "@icons/MapPinIcon";
import { formatDuration } from "@lib/datetime";
import { invalidateLocalStorageAuth } from "@lib/lib";
import { createFileRoute, Link } from "@tanstack/react-router";
import PhaseItem from "./-components/PhaseItem";

async function fetchServiceDetails(serviceId, merchantName) {
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

export const Route = createFileRoute("/m/$merchantName/services/$serviceId/")({
  component: ServiceDetailsPage,
  loaderDeps: ({ search: { locationId } }) => ({
    locationId,
  }),
  loader: async ({ params, deps: { locationId } }) => {
    const serviceDetails = await fetchServiceDetails(
      params.serviceId,
      params.merchantName
    );

    const nextAvailable = await fetchNextAvailable(
      params.merchantName,
      params.serviceId,
      locationId
    );

    return {
      serviceDetails: serviceDetails,
      nextAvailable: nextAvailable,
    };
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
  const loaderData = Route.useLoaderData();
  const hasAvailableSlot =
    loaderData.nextAvailable.date && loaderData.nextAvailable.time;

  const sortedPhases = [...loaderData.serviceDetails.phases].sort(
    (a, b) => (a.sequence || 0) - (b.sequence || 0)
  );

  return (
    <div className="p-4">
      <Card styles="md:mx-auto min-h-screen max-w-7xl md:p-6 p-4">
        <div className="grid w-full grid-cols-1 gap-8 lg:grid-cols-2 lg:gap-16">
          <div className="flex w-full flex-col gap-8">
            <div className="flex items-center gap-6">
              <div className="shrink-0 overflow-hidden rounded-lg xl:size-[120px]">
                <img
                  className="size-full object-cover"
                  src="https://dummyimage.com/120x120/d156c3/000000.jpg"
                  alt="service photo"
                ></img>
              </div>
              <div className="flex flex-col gap-4">
                <span className="text-2xl font-bold">
                  {loaderData.serviceDetails.name}
                </span>
                <div className="flex flex-col gap-2">
                  {loaderData.serviceDetails.price.number}{" "}
                  {loaderData.serviceDetails.price.currency}{" "}
                  {loaderData.serviceDetails?.price_note}
                  <div className="flex items-center justify-start gap-3">
                    <ClockIcon styles="size-4 fill-text_color" />
                    {formatDuration(loaderData.serviceDetails.total_duration)}
                  </div>
                </div>
              </div>
            </div>
            {loaderData.serviceDetails?.description && (
              <div className="flex flex-col gap-2">
                <div className="text-lg font-semibold">Description</div>
                <p>{loaderData.serviceDetails.description}</p>
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
                      index={index}
                      isLast={index === sortedPhases.length - 1}
                      loaderData={loaderData}
                    />
                  ))}
                </div>
              </div>
            )}
          </div>
          <div className="flex flex-col gap-6 md:mt-3 md:gap-8">
            <div className="border-primary bg-primary/20 rounded-md border-2 border-dashed p-4">
              <div className="flex flex-col gap-6 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex flex-col gap-2 sm:gap-4">
                  <div className="flex items-center gap-2">
                    <CalendarIcon styles="size-5 mb-0.5 text-text_color" />
                    <span className="text-text_color text-sm font-medium tracking-wide">
                      NEXT AVAILABLE
                    </span>
                  </div>
                  {hasAvailableSlot ? (
                    <div className="text-text_color flex items-center gap-4 text-base font-semibold">
                      <div className="flex items-center gap-2">
                        <span>{formatDate(loaderData.nextAvailable.date)}</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <ClockIcon styles="size-4 stroke-text_color" />
                        <span>{loaderData.nextAvailable.time}</span>
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
                    locationId: 15,
                    serviceId: loaderData.serviceDetails.id,
                    day:
                      loaderData.nextAvailable.date ||
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
