import {
  ArrowLeft01Icon,
  Call02Icon,
  Clock01Icon,
  Facebook01Icon,
  InstagramIcon,
  Location01Icon,
  Mail01Icon,
  StarIcon,
  TiktokIcon,
  Youtube,
} from "@hugeicons/core-free-icons";
import {
  Avatar,
  Button,
  Icon,
  Loading,
  ServerError,
} from "@reservations/components";
import { useWindowSize } from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import BusinessHoursSection from "./-components/BusinessHoursSection";
import FavoriteButton from "./-components/FavoriteButton";
import Gallery from "./-components/Gallery";
import MapboxMap from "./-components/MapboxMap";
import ReservationSection from "./-components/ReservationSection";
import ServiceSection from "./-components/ServiceSection";
import TeamSection from "./-components/TeamSection";

async function fetchMerchantInfo(name) {
  const response = await fetch(`/api/v1/public/merchants/${name}`, {
    method: "GET",
  });

  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  } else {
    return result.data;
  }
}

function merchantInfoQueryOptions(name) {
  return queryOptions({
    queryKey: ["merchant-info", name],
    queryFn: () => fetchMerchantInfo(name),
  });
}

export const Route = createFileRoute("/m/$merchantName/")({
  component: MerchantPage,
  loader: async ({ params, context: { queryClient } }) => {
    await queryClient.ensureQueryData(
      merchantInfoQueryOptions(params.merchantName)
    );
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

const days = [
  "Sunday",
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
];

function MerchantPage() {
  const { merchantName } = Route.useParams({ from: Route.id });
  const { isWindowSmall } = useWindowSize();

  const {
    data: merchantInfo,
    isLoading,
    isError,
    error,
  } = useQuery(merchantInfoQueryOptions(merchantName));

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  const businessHoursStatus = merchantInfo.business_hours_status;

  let nextOpenStr = "";
  if (!businessHoursStatus.is_business_open) {
    const todayIndex = new Date().getDay();
    nextOpenStr =
      businessHoursStatus.next_open_day === todayIndex
        ? "later today"
        : `on ${days[businessHoursStatus.next_open_day]}`;
  }

  return (
    <div className="bg-bg relative flex min-h-screen w-full flex-col">
      <div
        className="bg-bg border-border sticky top-0 z-50 flex w-full
          items-center justify-between border-b p-3 shadow-sm lg:hidden"
      >
        <Link className="hover:bg-hover rounded-md p-1" to="/home">
          <Icon icon={ArrowLeft01Icon} styles="text-fg" />
        </Link>
        <span className="truncate px-2.5 text-lg font-semibold">
          {merchantInfo.merchant_name}
        </span>
        <FavoriteButton
          merchantId={merchantInfo.merchant_id}
          styles="hover:bg-hover cursor-pointer rounded-md p-1"
        />
      </div>

      <div className="mx-auto flex w-full max-w-360 flex-col lg:px-10 lg:pt-6">
        {!isWindowSmall && (
          <>
            <Link
              className="hover:bg-hover border-border bg-bg-raised shadow-card
                mb-5 flex h-fit w-fit rounded-full border p-2.5"
              to="/home"
            >
              <Icon icon={ArrowLeft01Icon} styles="text-fg size-5" />
            </Link>
            <div className="flex w-full items-start justify-between pb-4">
              <div className="flex flex-col gap-3">
                <div className="flex items-center gap-4">
                  {merchantInfo.logo_url && (
                    <Avatar
                      styles="size-14! shrink-0 rounded-xl!"
                      img={merchantInfo.logo_url}
                      alt={`${merchantInfo.merchant_name} logo`}
                      fallback={null}
                    />
                  )}
                  <h1 className="text-4xl font-semibold">
                    {merchantInfo.merchant_name}
                  </h1>
                </div>
                <div className="text-fg-muted flex items-center gap-3">
                  <div className="text-fg-warning flex items-center gap-1">
                    <Icon icon={StarIcon} styles="size-5 fill-warning" />
                    <span className="font-medium">4.6</span>
                    <span className="text-fg-primary ml-1 cursor-pointer">
                      (9 reviews)
                    </span>
                  </div>
                  <span>•</span>
                  <div className="flex items-center gap-2">
                    <Icon icon={Clock01Icon} styles="size-5" />
                    {businessHoursStatus.is_business_open ? (
                      <span className="flex items-center gap-1">
                        <span className="text-fg-success">Open</span>- closes at{" "}
                        {businessHoursStatus.close_time}
                      </span>
                    ) : (
                      <span className="flex items-center gap-1 text-nowrap">
                        <span className="text-fg-warning">Closed</span>- opens{" "}
                        {nextOpenStr}
                      </span>
                    )}
                  </div>
                  <span>•</span>
                  <div className="flex items-center gap-2">
                    <Icon icon={Location01Icon} styles="size-5" />
                    <div className="flex items-start gap-2">
                      <p>{merchantInfo.formatted_location}</p>

                      <a
                        href={`https://www.google.com/maps/dir/?api=1&destination=${merchantInfo.geo_point.latitude},${merchantInfo.geo_point.longitude}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-fg-primary hover:underline"
                      >
                        (Plan a ride)
                      </a>
                    </div>
                  </div>
                </div>
              </div>

              <FavoriteButton
                merchantId={merchantInfo.merchant_id}
                styles="bg-bg-raised border-border shadow-card hover:bg-hover
                  mt-1.5 h-fit cursor-pointer rounded-full border p-3
                  hover:scale-105"
              />
            </div>
          </>
        )}

        <Gallery
          images={merchantInfo.images}
          merchantName={merchantInfo.merchant_name}
          logoUrl={merchantInfo.logo_url}
        />

        <div
          className="relative flex flex-col pt-0 pb-24 lg:flex-row lg:gap-3
            lg:pt-14 xl:gap-18"
        >
          <div
            className="bg-bg relative z-10 -mt-6 flex w-full flex-col gap-10
              rounded-t-3xl px-6 pt-10 pb-12 lg:mt-0 lg:w-[65%] lg:gap-12
              lg:rounded-none lg:px-0 lg:pt-0"
          >
            <div className="flex flex-col gap-3.5 lg:hidden">
              <div className="flex items-center gap-3">
                {merchantInfo.logo_url && (
                  <Avatar
                    styles="size-12! shrink-0 rounded-xl!"
                    img={merchantInfo.logo_url}
                    alt={`${merchantInfo.merchant_name} logo`}
                    fallback={null}
                  />
                )}
                <h1 className="text-3xl font-bold">
                  {merchantInfo.merchant_name}
                </h1>
              </div>
              <div className="text-fg-muted flex flex-wrap items-center gap-3">
                <div className="text-fg-warning flex items-center gap-2">
                  <Icon icon={StarIcon} styles="size-5 fill-warning" />
                  <div className="flex items-center gap-1">
                    <span>4.6</span>
                    <button className="text-fg-primary" onClick={() => {}}>
                      (9)
                    </button>
                  </div>
                </div>
                <span>•</span>
                <div className="flex items-center gap-2">
                  <Icon icon={Clock01Icon} styles="size-5" />
                  {businessHoursStatus.is_business_open ? (
                    <span className="flex items-center gap-1">
                      <span className="text-fg-success font-medium">Open</span>-
                      closes at {businessHoursStatus.close_time}
                    </span>
                  ) : (
                    <span className="flex items-center gap-1 text-nowrap">
                      <span className="text-fg-warning font-medium">
                        Closed
                      </span>
                      - opens {nextOpenStr}
                    </span>
                  )}
                </div>
              </div>
              <div className="text-fg-muted flex items-center gap-2">
                <Icon icon={Location01Icon} styles="size-5" />
                <span>{merchantInfo.formatted_location}</span>
              </div>
            </div>

            <div className="flex w-full flex-col gap-2 lg:gap-4">
              <h2 className="text-xl font-semibold lg:text-2xl">
                Get to Know Us
              </h2>
              <p className="text-fg-muted">{merchantInfo.introduction}</p>
              {merchantInfo.announcement && (
                <div
                  className="border-primary bg-primary-subtle mt-2 rounded-r-lg
                    border-l-3 py-3 pr-4 pl-4"
                >
                  <p className="text-fg text-sm">{merchantInfo.announcement}</p>
                </div>
              )}
            </div>
            <ReservationSection name="Select a Service" show={true}>
              <ServiceSection
                router={Route}
                merchantInfo={merchantInfo}
                merchantName={merchantName}
              />
            </ReservationSection>
            <ReservationSection name="About us" show={merchantInfo.about_us}>
              <p className="text-fg-muted">{merchantInfo.about_us}</p>
            </ReservationSection>

            <ReservationSection name="Team" show={true}>
              <TeamSection merchantName={merchantName} />
            </ReservationSection>

            <ReservationSection
              name="Location"
              show={merchantInfo.formatted_location}
            >
              <div className="flex flex-col gap-4 lg:gap-6">
                <div className="flex flex-wrap items-start gap-1">
                  <p>{merchantInfo.formatted_location}</p>

                  <a
                    href={`https://www.google.com/maps/dir/?api=1&destination=${merchantInfo.geo_point.latitude},${merchantInfo.geo_point.longitude}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-fg-primary hover:underline"
                  >
                    (Plan a ride)
                  </a>
                </div>
                <div
                  className="border-border shadow-card overflow-hidden
                    rounded-2xl border"
                >
                  <MapboxMap
                    styles={isWindowSmall ? "h-64 w-full" : "h-87.5 w-full"}
                    coordinates={[
                      merchantInfo.geo_point.longitude,
                      merchantInfo.geo_point.latitude,
                    ]}
                    minHeight={isWindowSmall ? 200 : 350}
                    zoom={isWindowSmall ? 13 : 14}
                  />
                </div>

                <p className="text-fg-muted text-center text-sm">
                  {merchantInfo.parking_info}
                </p>
              </div>
            </ReservationSection>

            <div
              className="border-border bg-bg sticky bottom-0 z-10 flex w-full
                items-center justify-between border-y py-3 lg:hidden"
            >
              <span>{merchantName}</span>
              <Link
                from={Route.fullPath}
                to="book"
                search={{ locationId: merchantInfo.location_id }}
              >
                <Button buttonText="Reserve Now" styles="py-2 px-4 text-sm" />
              </Link>
            </div>

            <div className="flex w-full flex-col gap-10 lg:flex-row lg:gap-20">
              <ReservationSection name="Business Hours" show={true}>
                <div className="pt-2 lg:w-96">
                  <BusinessHoursSection
                    hoursData={merchantInfo.business_hours}
                  />
                </div>
              </ReservationSection>

              <ReservationSection
                name="Contact"
                show={merchantInfo.contact_email || merchantInfo.phone_number}
              >
                <div className="flex flex-col gap-6 pt-2 lg:gap-8">
                  <div
                    className="flex flex-col gap-4 md:flex-row md:flex-wrap
                      lg:gap-6"
                  >
                    {merchantInfo.contact_email && (
                      <div
                        className="text-fg-muted flex items-center gap-3
                          text-sm"
                      >
                        <Icon icon={Mail01Icon} styles="size-5" />
                        {merchantInfo.contact_email}
                      </div>
                    )}
                    {merchantInfo.phone_number && (
                      <div
                        className="text-fg-muted flex items-center gap-3
                          text-sm"
                      >
                        <Icon icon={Call02Icon} styles="size-5" />
                        {merchantInfo.phone_number}
                      </div>
                    )}
                  </div>

                  <div
                    className="border-border pt-2 lg:max-w-md lg:border-t
                      lg:pt-6"
                  >
                    <p className="text-fg mb-4 text-sm font-semibold">
                      Follow us
                    </p>
                    <div
                      className="*:hover:text-fg-primary *:text-fg-muted flex
                        items-center justify-evenly gap-6 *:size-6
                        *:cursor-pointer *:transition-colors sm:justify-start
                        lg:gap-8 lg:*:size-6"
                    >
                      <Icon icon={InstagramIcon} />
                      <Icon icon={Facebook01Icon} />
                      <Icon icon={Youtube} />
                      <Icon icon={TiktokIcon} />
                    </div>
                  </div>
                </div>
              </ReservationSection>
            </div>
          </div>

          <div className="hidden lg:block lg:w-[35%]">
            <div
              className="border-border bg-bg-raised shadow-card sticky top-14
                flex flex-col gap-5 rounded-2xl border p-6"
            >
              <div className="flex items-start justify-between gap-2 pb-4">
                <div className="flex flex-col gap-2">
                  <h2 className="text-2xl font-semibold tracking-tight">
                    {merchantInfo.merchant_name}
                  </h2>
                  <div className="flex items-center gap-1.5 text-sm font-medium">
                    <Icon
                      icon={StarIcon}
                      styles="size-4 fill-warning text-fg-warning"
                    />
                    <span>4.6</span>
                    <span className="text-fg-primary ml-1 font-normal">
                      (9 reviews)
                    </span>
                  </div>
                </div>
                <FavoriteButton
                  merchantId={merchantInfo.merchant_id}
                  styles="bg-bg-raised border-border hover:bg-hover h-fit
                    cursor-pointer rounded-full border p-2.5 hover:scale-105"
                />
              </div>
              <Link
                from={Route.fullPath}
                to="book"
                search={{ locationId: merchantInfo.location_id }}
              >
                <Button buttonText="Reserve Now" styles="w-full py-2.5" />
              </Link>

              <div
                className="border-border-subtle text-fg-muted mt-2 flex flex-col
                  gap-4 border-t pt-6 text-sm"
              >
                <div className="flex items-center gap-2">
                  <Icon icon={Clock01Icon} styles="size-4" />
                  {businessHoursStatus.is_business_open ? (
                    <span className="flex items-center gap-1">
                      <span className="text-fg-success font-medium">Open</span>-
                      closes at {businessHoursStatus.close_time}
                    </span>
                  ) : (
                    <span className="flex items-center gap-1 text-nowrap">
                      <span className="text-fg-warning font-medium">
                        Closed
                      </span>
                      - opens {nextOpenStr}
                    </span>
                  )}
                </div>
                <div className="flex items-start gap-1.5">
                  <Icon icon={Location01Icon} styles="size-4 mt-0.5" />
                  <span>{merchantInfo.formatted_location}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
