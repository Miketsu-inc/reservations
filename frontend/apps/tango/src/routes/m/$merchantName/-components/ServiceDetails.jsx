import {
  Calendar02Icon,
  CalendarOffIcon,
  Clock01Icon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";
import {
  Avatar,
  Button,
  CloseButton,
  Drawer,
  DrawerContent,
  Icon,
  Modal,
  ServerError,
} from "@reservations/components";
import {
  formatDuration,
  getDisplayPrice,
  invalidateLocalStorageAuth,
  timeStringFromDate,
} from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";

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
  onReserve,
  isWindowSmall,
}) {
  const {
    data: nextAvailable,
    isLoading,
    isError,
    error,
  } = useQuery({
    ...nextAvailableQueryOptions(merchantName, service?.id, locationId),
    enabled: isOpen,
  });

  if (isError) {
    return <ServerError error={error.message} />;
  }

  const hasAvailableSlot = nextAvailable?.from_date;

  return isWindowSmall ? (
    <Drawer
      open={isOpen}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <DrawerContent
        styles="h-full relative"
        popUpStyles="h-[calc(80vh+3rem)]! overflow-y-hidden!"
      >
        <DetailsContent
          nextAvailable={nextAvailable}
          service={service}
          onReserve={onReserve}
          onClose={onClose}
          hasAvailable={hasAvailableSlot}
          isWindowSmall={isWindowSmall}
          category={category}
          isLoading={isLoading}
        />
      </DrawerContent>
    </Drawer>
  ) : (
    <Modal isOpen={isOpen} onClose={onClose} styles="px-6 pb-4 pt-1">
      <DetailsContent
        nextAvailable={nextAvailable}
        service={service}
        onReserve={onReserve}
        onClose={onClose}
        isWindowSmall={isWindowSmall}
        hasAvailable={hasAvailableSlot}
        category={category}
        isLoading={isLoading}
      />
    </Modal>
  );
}

function formatDate(dateString) {
  if (!dateString) return "";
  const date = new Date(dateString);

  const day = date.getDate();
  const month = date.toLocaleDateString("default", { month: "short" });
  const weekday = date.toLocaleDateString("default", { weekday: "short" });

  return `${weekday}, ${month} ${day}`;
}

function DetailsContent({
  service,
  category,
  nextAvailable,
  hasAvailable,
  onReserve,
  isWindowSmall,
  isLoading,
  onClose,
}) {
  const isGroupService = service?.booking_type !== "appointment";

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
                  className="border-secondary flex w-fit items-center gap-2
                    rounded-lg border bg-gray-100 px-3 py-1 text-sm
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
              <div
                className="flex items-center gap-1 font-medium text-gray-500
                  dark:text-gray-400"
              >
                <Icon icon={Clock01Icon} styles="size-5" />
                <span>{formatDuration(service?.total_duration)}</span>
              </div>
            </div>
          </div>
        </div>
        {service?.description && <p className="">{service?.description}</p>}
        {isLoading ? (
          <div
            className={`border-border_color animate-pulse rounded-md border
              bg-gray-200/50 ${isGroupService ? "h-34" : "h-20"}
              dark:bg-gray-200/5`}
          ></div>
        ) : (
          <div
            className="border-border_color gap-4 rounded-md border bg-gray-100
              p-3.5 dark:bg-gray-200/5"
          >
            {hasAvailable ? (
              isGroupService ? (
                <div className="flex flex-col gap-4">
                  <div className="flex items-center gap-4">
                    <div
                      className="flex items-center justify-center rounded-lg
                        bg-green-600/15 p-2"
                    >
                      <Icon
                        icon={Calendar02Icon}
                        styles="size-7 text-green-600"
                      />
                    </div>
                    <div className="flex flex-col gap-0.5">
                      <span className="text-gray-500 dark:text-gray-400">
                        Nearest session with spots
                      </span>
                      <div className="flex items-center gap-2 text-base">
                        <span>{formatDate(nextAvailable.from_date)}</span>

                        <span>
                          {timeStringFromDate(
                            new Date(nextAvailable.from_date)
                          )}
                          {` - ${timeStringFromDate(new Date(nextAvailable.to_date))}`}
                        </span>
                      </div>
                    </div>
                  </div>

                  <div
                    className="border-text_color/10 flex items-center
                      justify-between border-t px-2 pt-2"
                  >
                    {nextAvailable.employee && (
                      <div className="flex items-center gap-2">
                        <Avatar
                          styles="size-8! text-xs rounded-full!"
                          initials="MM"
                        />
                        <span className="text-text_color text-sm">
                          {/* {nextAvailable.employee} */}
                          Mikes Marcell
                        </span>
                      </div>
                    )}
                    <div className="flex items-center gap-1.5 py-1.5 text-sm">
                      <Icon icon={UserGroupIcon} styles="size-5" />
                      <span>
                        {nextAvailable.current_participants} /{" "}
                        {service?.max_participants}
                      </span>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="flex items-center gap-4">
                  <div className="rounded-lg bg-green-600/15 p-2">
                    <Icon
                      icon={Calendar02Icon}
                      styles="text-green-600 size-8"
                    />
                  </div>
                  <div className="flex flex-col gap-1">
                    <span className="text-gray-500 dark:text-gray-400">
                      Next Available
                    </span>
                    <div className="flex gap-2">
                      <span className="">
                        {formatDate(nextAvailable.from_date)}
                      </span>
                      <span>•</span>
                      <span className="">
                        {timeStringFromDate(new Date(nextAvailable.from_date))}
                      </span>
                    </div>
                  </div>
                </div>
              )
            ) : (
              <div className="flex items-center gap-4">
                <div className="rounded-lg py-2">
                  <Icon icon={CalendarOffIcon} styles="text-gray-400 size-7" />
                </div>
                <div className="flex items-center gap-2 text-sm">
                  <span className="">
                    {isGroupService
                      ? "No open sessions in the near future"
                      : "No availability in the near future"}
                  </span>
                </div>
              </div>
            )}
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
        <Button
          variant="primary"
          styles="w-fit py-2 px-8"
          buttonText="Reserve"
          onClick={onReserve}
        />
      </div>
    </div>
  );
}
