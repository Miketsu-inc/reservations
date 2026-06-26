import {
  ArrowLeft01Icon,
  Call02Icon,
  Clock01Icon,
  Facebook01Icon,
  FavouriteIcon,
  InstagramIcon,
  Location01Icon,
  Mail01Icon,
  StarIcon,
  TiktokIcon,
  Youtube,
} from "@hugeicons/core-free-icons";
import { Button, Icon, Loading, ServerError } from "@reservations/components";
import { useWindowSize } from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import BusinessHoursSection from "./-components/BusinessHoursSection";
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
  const [isFavourite, setIsFavourite] = useState(false);
  const windowSize = useWindowSize();
  const isWindowSmall = windowSize === "sm" || windowSize === "md";

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
    return <ServerError error={error} />;
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
    <div
      className="bg-layer_bg lg:bg-bg_color relative flex min-h-screen w-full
        flex-col"
    >
      {isWindowSmall && (
        <div
          className="bg-layer_bg border-border_color sticky top-0 z-50 flex
            w-full items-center justify-between border-b-2 p-3 shadow-sm"
        >
          <Link
            className="hover:bg-hvr_gray rounded-md p-1"
            to={`http://app.reservations.local:3000/dashboard`}
            from={Route.fullPath}
          >
            <Icon icon={ArrowLeft01Icon} styles="text-text_color" />
          </Link>
          <span className="truncate px-2.5 text-lg font-semibold">
            {merchantInfo.merchant_name}
          </span>
          <button
            className="hover:bg-hvr_gray rounded-md p-1"
            onClick={() => {
              setIsFavourite(!isFavourite);
            }}
          >
            <Icon
              icon={FavouriteIcon}
              styles={` size-6
              ${isFavourite ? "text-red-600 fill-red-600" : "text-text_color"}`}
            />
          </button>
        </div>
      )}

      <div className="mx-auto flex w-full max-w-360 flex-col lg:px-10 lg:pt-6">
        {!isWindowSmall && (
          <>
            <Link
              className="hover:bg-hvr_gray/20 border-border_color bg-layer_bg
                mb-5 flex h-fit w-fit rounded-full border p-3 shadow-sm
                transition-colors"
              to={`http://app.reservations.local:3000/dashboard`}
              from=""
            >
              <Icon icon={ArrowLeft01Icon} styles="text-text_color size-6" />
            </Link>
            <div className="flex w-full items-start justify-between pb-4">
              <div className="flex flex-col gap-4">
                <h1 className="text-4xl font-bold">
                  {merchantInfo.merchant_name}
                </h1>
                <div className="flex items-center gap-3 text-base">
                  <div className="flex items-center gap-1 text-yellow-500">
                    <Icon icon={StarIcon} styles="size-5 fill-yellow-500" />
                    <span className="font-medium">4.6</span>
                    <span className="text-primary ml-1 cursor-pointer">
                      (9 reviews)
                    </span>
                  </div>
                  <span className="">•</span>
                  <div className="flex items-center gap-2">
                    <Icon icon={Clock01Icon} styles="size-5" />
                    {businessHoursStatus.is_business_open ? (
                      <span className="flex items-center gap-1">
                        <span className="text-green-600 dark:text-green-500">
                          Open
                        </span>
                        - closes at {businessHoursStatus.close_time}
                      </span>
                    ) : (
                      <span className="flex items-center gap-1 text-nowrap">
                        <span className="text-orange-700 dark:text-orange-500">
                          Closed
                        </span>
                        - opens {nextOpenStr}
                      </span>
                    )}
                  </div>
                  <span className="">•</span>
                  <div className="flex items-center gap-2">
                    <Icon icon={Location01Icon} styles="size-5" />
                    <div className="flex items-start gap-2">
                      <p className="lg:text-lg">
                        {merchantInfo.formatted_location}
                      </p>

                      <a
                        href={`https://www.google.com/maps/dir/?api=1&destination=${merchantInfo.geo_point.latitude},${merchantInfo.geo_point.longitude}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-primary hover:underline"
                      >
                        (Plan a ride)
                      </a>
                    </div>
                  </div>
                </div>
              </div>

              <button
                className="bg-layer_bg hover:bg-hvr_gray/20 border-border_color
                  mt-1.5 h-fit rounded-full border p-3 shadow-sm
                  transition-colors"
                onClick={() => {
                  setIsFavourite(!isFavourite);
                }}
              >
                <Icon
                  icon={FavouriteIcon}
                  styles={` size-6
                  ${isFavourite ? "text-red-600 fill-red-600" : "text-text_color"}`}
                />
              </button>
            </div>
          </>
        )}

        <div
          className="grid w-full grid-cols-1 gap-2 lg:h-125
            lg:grid-cols-[2fr_1fr] lg:overflow-hidden lg:rounded-2xl"
        >
          <div className="relative h-62.5 w-full lg:h-full">
            <img
              className="absolute inset-0 size-full cursor-pointer object-cover
                transition-opacity hover:opacity-95"
              src="https://dummyimage.com/1920x1080/d156c3/000000.jpg"
              alt="Main location view"
            />
          </div>
          <div className="hidden h-full w-full flex-col gap-2 lg:flex">
            <div className="relative flex-1">
              <img
                className="absolute inset-0 size-full cursor-pointer
                  object-cover transition-opacity hover:opacity-95"
                src="https://dummyimage.com/800x600/a34298/fff.jpg"
                alt="Gallery top right"
              />
            </div>
            <div className="relative flex-1">
              <img
                className="absolute inset-0 size-full cursor-pointer
                  object-cover transition-opacity hover:opacity-95"
                src="https://dummyimage.com/800x600/8a3781/fff.jpg"
                alt="Gallery bottom right"
              />
            </div>
          </div>
        </div>

        <div
          className="relative flex flex-col pt-0 pb-24 lg:flex-row lg:gap-3
            lg:pt-14 xl:gap-18"
        >
          <div
            className="relative z-10 -mt-6 flex w-full flex-col gap-10
              rounded-t-3xl px-6 pt-10 pb-12 lg:mt-0 lg:w-[65%] lg:gap-12
              lg:rounded-none lg:px-0 lg:pt-0"
          >
            {isWindowSmall && (
              <div className="flex flex-col gap-3.5">
                <h1 className="text-3xl font-bold">
                  {merchantInfo.merchant_name}
                </h1>
                <div className="flex flex-wrap items-center gap-3">
                  <div className="flex items-center gap-2 text-yellow-500">
                    <Icon icon={StarIcon} styles="size-5 fill-yellow-500" />
                    <div className="flex items-center gap-1">
                      <span>4.6</span>
                      <button className="text-primary" onClick={() => {}}>
                        (9)
                      </button>
                    </div>
                  </div>
                  <span className="text-gray-400">•</span>
                  <div className="flex items-center gap-2">
                    <Icon icon={Clock01Icon} styles="size-5" />
                    {businessHoursStatus.is_business_open ? (
                      <span className="flex items-center gap-1">
                        <span
                          className="font-medium text-green-600
                            dark:text-green-500"
                        >
                          Open
                        </span>
                        - closes at {businessHoursStatus.close_time}
                      </span>
                    ) : (
                      <span className="flex items-center gap-1 text-nowrap">
                        <span
                          className="font-medium text-orange-700
                            dark:text-orange-500"
                        >
                          Closed
                        </span>
                        - opens {nextOpenStr}
                      </span>
                    )}
                  </div>
                </div>
                <div
                  className="flex items-center gap-3 text-gray-600
                    dark:text-gray-300"
                >
                  <Icon icon={Location01Icon} styles="size-5" />
                  <span>{merchantInfo.formatted_location}</span>
                </div>
              </div>
            )}

            <div className="flex w-full flex-col gap-2 lg:gap-4">
              <h2 className="text-xl font-semibold lg:text-2xl">
                Get to Know Us
              </h2>
              <p className="lg:text-lg">{merchantInfo.introduction}</p>
              {merchantInfo.announcement && (
                <div
                  className="border-primary bg-primary/5 mt-2 rounded-r-lg
                    border-l-3 py-3 pl-4"
                >
                  <p className="text-primary-dark text-base lg:text-lg">
                    {merchantInfo.announcement}
                  </p>
                </div>
              )}
            </div>
            <ReservationSection name="Select a Service" show={true}>
              <ServiceSection
                onSelect={() => {}}
                router={Route}
                merchantInfo={merchantInfo}
                isWindowSmall={isWindowSmall}
                merchantName={merchantName}
              />
            </ReservationSection>
            <ReservationSection name="About us" show={merchantInfo.about_us}>
              <p className="lg:text-lg">{merchantInfo.about_us}</p>
            </ReservationSection>

            <ReservationSection name="Team" show={true}>
              <TeamSection
                isWindowSmall={isWindowSmall}
                merchantName={merchantName}
              />
            </ReservationSection>

            <ReservationSection
              name="Location"
              show={merchantInfo.formatted_location}
            >
              <div className="flex flex-col gap-4 lg:gap-6">
                <div className="flex flex-wrap items-start gap-1">
                  <p className="lg:text-lg">
                    {merchantInfo.formatted_location}
                  </p>

                  <a
                    href={`https://www.google.com/maps/dir/?api=1&destination=${merchantInfo.geo_point.latitude},${merchantInfo.geo_point.longitude}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary hover:underline"
                  >
                    (Plan a ride)
                  </a>
                </div>
                <div
                  className="border-border_color overflow-hidden rounded-2xl
                    border shadow-sm"
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

                <p className="text-center lg:text-lg">
                  {merchantInfo.parking_info}
                </p>
              </div>
            </ReservationSection>

            {isWindowSmall && (
              <div
                className="border-border_color bg-layer_bg sticky bottom-0 z-10
                  flex w-full items-center justify-between border-y py-4"
              >
                <span>{merchantName}</span>
                <Link
                  from={Route.fullPath}
                  to="book"
                  search={{ locationId: merchantInfo.location_id }}
                >
                  <Button buttonText="Reserve Now" styles="py-2 px-4" />
                </Link>
              </div>
            )}

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
                show={merchantInfo.contact_email}
              >
                <div className="flex flex-col gap-6 pt-2 lg:gap-8">
                  <div
                    className="flex flex-col gap-4 md:flex-row md:flex-wrap
                      lg:gap-6"
                  >
                    <div className="flex items-center gap-4 lg:text-lg">
                      <Icon
                        icon={Mail01Icon}
                        styles="size-5 lg:size-6 text-gray-500"
                      />
                      {merchantInfo.contact_email}
                    </div>
                    <div className="flex items-center gap-4 lg:text-lg">
                      <Icon
                        icon={Call02Icon}
                        styles="size-5 lg:size-6 text-gray-500"
                      />
                      {"+36 20 538  3565"}
                    </div>
                  </div>

                  <div
                    className="border-border_color pt-2 lg:max-w-md lg:border-t
                      lg:pt-6"
                  >
                    <p
                      className="mb-4 text-sm font-semibold tracking-wider
                        text-gray-500 lg:mb-5 lg:uppercase"
                    >
                      Follow us
                    </p>
                    <div
                      className="*:hover:text-primary flex items-center
                        justify-evenly gap-6 *:size-7 *:cursor-pointer
                        *:text-gray-600 *:transition-colors sm:justify-start
                        lg:gap-8 lg:*:size-7 dark:*:text-gray-400"
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

          {!isWindowSmall && (
            <div className="lg:w-[35%]">
              <div
                className="border-border_color bg-layer_bg sticky top-14 flex
                  flex-col gap-5 rounded-xl border p-8 shadow-lg"
              >
                <div className="flex items-start justify-between gap-2 pb-4">
                  <div className="flex flex-col gap-2">
                    <h2 className="text-4xl font-bold">
                      {merchantInfo.merchant_name}
                    </h2>
                    <div className="flex items-center gap-2 text-xl font-medium">
                      <Icon
                        icon={StarIcon}
                        styles="size-5 fill-yellow-500 text-yellow-500"
                      />
                      <span>4.6</span>
                      <span className="text-primary ml-1 font-normal">
                        (9 reviews)
                      </span>
                    </div>
                  </div>
                  <button
                    className="bg-layer_bg hover:bg-hvr_gray/20
                      border-border_color h-fit rounded-full border p-3
                      shadow-md transition-colors"
                    onClick={() => {
                      setIsFavourite(!isFavourite);
                    }}
                  >
                    <Icon
                      icon={FavouriteIcon}
                      styles={` size-6
                      ${isFavourite ? "text-red-600 fill-red-600" : "text-text_color"}`}
                    />
                  </button>
                </div>
                <Link
                  from={Route.fullPath}
                  to="book"
                  search={{ locationId: merchantInfo.location_id }}
                >
                  <Button
                    buttonText="Reserve Now"
                    styles="w-full py-2.5 text-lg"
                  />
                </Link>

                <div
                  className="border-border_color mt-2 flex flex-col gap-5
                    border-t pt-6 text-gray-600 dark:text-gray-300"
                >
                  <div className="flex items-center gap-2">
                    <Icon icon={Clock01Icon} styles="size-5" />
                    {businessHoursStatus.is_business_open ? (
                      <span className="flex items-center gap-1">
                        <span
                          className="font-medium text-green-600
                            dark:text-green-500"
                        >
                          Open
                        </span>
                        - closes at {businessHoursStatus.close_time}
                      </span>
                    ) : (
                      <span className="flex items-center gap-1 text-nowrap">
                        <span
                          className="font-medium text-orange-700
                            dark:text-orange-500"
                        >
                          Closed
                        </span>
                        - opens {nextOpenStr}
                      </span>
                    )}
                  </div>
                  <div className="flex items-start gap-1.5">
                    <Icon icon={Location01Icon} styles="size-5 mt-0.5" />
                    <span>{merchantInfo.formatted_location}</span>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
