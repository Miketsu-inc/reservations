import { Cancel01Icon } from "@hugeicons/core-free-icons";
import { Icon, ServerError } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { invalidateLocalStorageAuth, useWindowSize } from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import BlockedTimePanel from "./BlockedTimePanel";
import EditBookingPanel from "./EditBookingPanel";
import NewBookingPanel from "./NewBookingPanel";

async function fetchSidePanelResource(merchantId, resource) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/calendar/${resource}`,
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

function customersForCalendarQueryOptions(merchantId) {
  return queryOptions({
    queryKey: [merchantId, "customers-calendar"],
    queryFn: () => fetchSidePanelResource(merchantId, "customers"),
  });
}

function servicesForCalendarQueryOptions(merchantId) {
  return queryOptions({
    queryKey: [merchantId, "services-calendar"],
    queryFn: () => fetchSidePanelResource(merchantId, "services"),
  });
}

export default function CalendarSidePanel({
  isOpen,
  onClose,
  onTransitionEnd,
  type,
  data,
  initialStart,
  panelKey,
  onSave,
  onSoftUpdate,
  preferences,
}) {
  const { isWindowSmall } = useWindowSize();
  const { merchantId } = useAuth();
  const shouldLoadBookingResources =
    isOpen && (type === "new-booking" || type === "edit-booking");

  const {
    data: customers = [],
    isError: customersIsError,
    error: customersError,
  } = useQuery({
    ...customersForCalendarQueryOptions(merchantId),
    enabled: shouldLoadBookingResources,
  });

  const {
    data: services = [],
    isError: servicesIsError,
    error: servicesError,
  } = useQuery({
    ...servicesForCalendarQueryOptions(merchantId),
    enabled: shouldLoadBookingResources,
  });

  if (customersIsError || servicesIsError) {
    return (
      <ServerError error={customersError?.message || servicesError?.message} />
    );
  }

  return (
    <>
      <div
        className={`bg-bg-backdrop fixed inset-0 z-40 transition-opacity
          duration-200
          ${isOpen ? "opacity-100" : "pointer-events-none opacity-0"}`}
        onClick={onClose}
        aria-hidden="true"
      />

      <aside
        className={`bg-bg border-border shadow-modal fixed top-0 right-0 z-50
          h-full transition-all duration-200 ease-in-out
          ${isOpen ? "translate-x-0" : "translate-x-full"}
          ${isWindowSmall ? "w-full" : "w-fit border-l"}`}
        onTransitionEnd={(event) => {
          if (event.target === event.currentTarget && !isOpen) {
            onTransitionEnd?.();
          }
        }}
      >
        {isOpen && !isWindowSmall && (
          <button
            onClick={onClose}
            className="border-border bg-bg text-fg-muted hover:text-fg
              shadow-overlay absolute top-4 -left-12 cursor-pointer rounded-full
              border p-1 transition-transform hover:scale-110"
            aria-label="Close panel"
          >
            <Icon icon={Cancel01Icon} styles="size-7 fill-fg" />
          </button>
        )}
        <div key={panelKey} className="bg-bg h-full min-w-full">
          {type === "new-booking" && (
            <NewBookingPanel
              onSave={onSave}
              onClose={onClose}
              categories={services}
              customers={customers}
              initialStart={initialStart}
            />
          )}
          {type === "blocked-time" && (
            <BlockedTimePanel
              blockedTime={data}
              initialStart={initialStart}
              preferences={preferences}
              onClose={onClose}
              onSubmitted={onSave}
              onDeleted={onSave}
            />
          )}
          {type === "edit-booking" && (
            <EditBookingPanel
              originalBookingData={data}
              onClose={onClose}
              onSave={onSave}
              onSoftUpdate={onSoftUpdate}
              customers={customers}
              categories={services}
              preferences={preferences}
            />
          )}
        </div>
      </aside>
    </>
  );
}
