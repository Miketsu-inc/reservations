import {
  CalendarOffIcon,
  Clock01Icon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";
import {
  Avatar,
  Button,
  CloseButton,
  Icon,
  ResponsiveDialog,
  ResponsiveDialogContent,
  ServerError,
} from "@reservations/components";
import {
  activeTeamQueryOptions,
  formatDuration,
  getDisplayPrice,
  invalidateLocalStorageAuth,
  timeStringFromDate,
  useWindowSize,
} from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";

async function fetchNextAvailable(merchantName, serviceId, locationId) {
  const response = await fetch(
    `/api/v1/public/merchants/${merchantName}/locations/${locationId}/services/${serviceId}/availability/next`,
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

export default function ServiceDetails({
  merchantName,
  locationId,
  category,
  service,
  isOpen,
  onClose,
  router,
}) {
  const { isWindowSmall } = useWindowSize();

  const {
    data: nextAvailable,
    isLoading,
    isError,
    error,
  } = useQuery({
    ...nextAvailableQueryOptions(merchantName, service?.id, locationId),
    enabled: isOpen,
  });

  const { data: teamMembers } = useQuery({
    ...activeTeamQueryOptions(merchantName),
    enabled: isOpen && !!merchantName,
  });

  if (isError) {
    return <ServerError error={error.message} />;
  }

  const hasAvailableSlot = Boolean(nextAvailable?.from_date);

  return (
    <ResponsiveDialog
      open={isOpen}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <ResponsiveDialogContent
        styles={isWindowSmall ? "relative" : "px-6 pb-4 pt-1"}
        popUpStyles="h-[calc(80vh+3rem)]! overflow-y-hidden!"
      >
        <DetailsContent
          nextAvailable={nextAvailable}
          service={service}
          onClose={onClose}
          hasAvailable={hasAvailableSlot}
          category={category}
          isLoading={isLoading}
          router={router}
          locationId={locationId}
          teamMembers={teamMembers}
        />
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}

function monthNameFromDate(date) {
  return date.toLocaleDateString([], { month: "short" });
}

function DetailsContent({
  service,
  category,
  nextAvailable,
  hasAvailable,
  router,
  isLoading,
  onClose,
  locationId,
  teamMembers,
}) {
  const { isWindowSmall } = useWindowSize();
  const isGroupService = service?.booking_type !== "appointment";

  const assignedEmployee = teamMembers?.find(
    (emp) => emp.id === nextAvailable?.employee
  );

  const fromDate = hasAvailable ? new Date(nextAvailable.from_date) : null;
  const toDate = hasAvailable ? new Date(nextAvailable.to_date) : null;

  return (
    <div
      className={`mt-3 flex h-full flex-col justify-start p-2
        ${isWindowSmall ? "w-full" : "w-150"} `}
    >
      <div className="flex justify-end">
        {!isWindowSmall && <CloseButton onClick={onClose} />}
      </div>
      <div className="flex flex-col gap-8">
        <div className="flex flex-col gap-3">
          <div className="flex flex-col gap-4">
            <p className="text-2xl font-medium">{service?.name}</p>
            <div className="flex items-center gap-7">
              {category && (
                <div
                  className="border-input_border_color flex w-fit items-center
                    gap-2 rounded-lg border bg-gray-100 px-3 py-1 text-sm
                    dark:bg-gray-200/10"
                >
                  {isGroupService && (
                    <div className="flex items-center gap-2">
                      <span>Group</span>
                      <span>•</span>
                    </div>
                  )}
                  {category}
                </div>
              )}
              <div className="text-text_color/80 flex items-center gap-1">
                <Icon icon={Clock01Icon} styles="size-5" />
                <span>
                  {service?.min_duration &&
                  service?.max_duration &&
                  service.min_duration !== service.max_duration
                    ? `${formatDuration(service.min_duration)} - ${formatDuration(service.max_duration)}`
                    : formatDuration(service?.total_duration)}
                </span>
              </div>
            </div>
          </div>
        </div>
        {service?.description && <p className="">{service?.description}</p>}
        {isLoading ? (
          <div
            className="border-border_color h-24 animate-pulse rounded-lg border
              bg-gray-200/20 dark:bg-gray-200/5"
          ></div>
        ) : hasAvailable ? (
          <div
            className="border-border_color bg-layer_bg rounded-lg border
              shadow-sm"
          >
            <div
              className="border-border_color flex flex-row items-center
                justify-between border-b px-4 py-3.5"
            >
              <div className="flex flex-row items-center gap-4">
                <div
                  className="flex min-w-8 flex-col items-center justify-center"
                >
                  <p className="text-lg leading-none font-semibold">
                    {fromDate.getDate()}
                  </p>
                  <p className="text-text_color/60 text-xs">
                    {monthNameFromDate(fromDate)}
                  </p>
                </div>
                <div className="border-border_color h-8 border-r" />
                <div className="flex flex-col items-start justify-center gap-1">
                  <p className="font-medium text-green-500">Next Available</p>
                  <div className="flex flex-row items-center gap-1.5">
                    <Icon
                      icon={Clock01Icon}
                      styles="size-4.5 text-text_color/60"
                    />
                    <p className="text-text_color/60 text-sm">
                      {fromDate.toLocaleDateString([], { weekday: "short" })},{" "}
                      {timeStringFromDate(fromDate)}
                      {toDate ? ` - ${timeStringFromDate(toDate)}` : ""}
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <div
              className="flex flex-row items-center justify-between px-4 py-2.5"
            >
              <div className="flex flex-row items-center gap-2">
                <Avatar
                  styles="size-7! text-xs! rounded-full!"
                  initials={`${assignedEmployee.first_name[0]}${assignedEmployee.last_name[0]}`}
                />
                <p className="text-text_color text-sm">{`${assignedEmployee.first_name} ${assignedEmployee.last_name}`}</p>
              </div>

              {isGroupService && (
                <div
                  className="border-border_color bg-bg_color text-text_color/70
                    flex w-fit flex-row items-center gap-1.5 rounded-lg border
                    px-2 py-1 text-xs"
                >
                  <Icon icon={UserGroupIcon} styles="size-3.5" />
                  <span>
                    {nextAvailable?.current_participants} /{" "}
                    {service?.max_participants}
                  </span>
                </div>
              )}
            </div>
          </div>
        ) : (
          <div
            className="border-border_color bg-layer_bg flex items-center gap-4
              rounded-lg border p-4 shadow-sm"
          >
            <div className="rounded-lg py-1">
              <Icon icon={CalendarOffIcon} styles="size-6 text-gray-400" />
            </div>
            <div className="flex items-center gap-2 text-sm">
              <span className="text-text_color/70">
                {isGroupService
                  ? "No open sessions in the near future"
                  : "No availability in the near future"}
              </span>
            </div>
          </div>
        )}
      </div>

      <div
        className={`${isWindowSmall ? "border-border_color absolute right-0 bottom-0 left-0 border-t" : "pt-10"}
          flex w-full items-center justify-between px-3 py-3`}
      >
        <span className="text-lg font-medium">
          {getDisplayPrice(service?.price, service?.price_type)}
        </span>

        <Link
          from={router.fullPath}
          to="book"
          search={{
            locationId: locationId,
            serviceId: service?.id,
            type: service?.booking_type,
          }}
          onClick={onClose}
        >
          <Button
            variant="primary"
            styles="w-fit py-2 px-8"
            name="Reserve"
            buttonText="Reserve"
          />
        </Link>
      </div>
    </div>
  );
}
