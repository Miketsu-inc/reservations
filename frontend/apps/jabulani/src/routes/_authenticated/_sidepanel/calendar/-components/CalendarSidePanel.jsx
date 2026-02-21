import { XIcon } from "@reservations/assets";
import { ServerError } from "@reservations/components";
import { invalidateLocalStorageAuth, useWindowSize } from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import BlockedTimePanel from "./BlockedTimePanel";
import NewBookingPanel from "./NewBookingPanel";

// async function fetchEmployees() {
//   const response = await fetch(`/api/v1/merchants/calendar/employees`, {
//     method: "GET",
//     headers: {
//       Accept: "application/json",
//       "constent-type": "application/json",
//     },
//   });

//   const result = await response.json();
//   if (!response.ok) {
//     throw result.error;
//   } else {
//     return result.data;
//   }
// }

// function employeeQueryOptions() {
//   return queryOptions({
//     queryKey: ["calendar-employees"],
//     queryFn: fetchEmployees,
//   });
// }

async function fetchCustomersForCalendar() {
  const response = await fetch("/api/v1/merchants/calendar/customers", {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

function customersForCalendarQueryOptions() {
  return queryOptions({
    queryKey: ["customers-calendar"],
    queryFn: fetchCustomersForCalendar,
  });
}

async function fetchServicesForCalendar() {
  const response = await fetch("/api/v1/merchants/calendar/services", {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

function servicesForCalendarQueryOptions() {
  return queryOptions({
    queryKey: ["services-calendar"],
    queryFn: fetchServicesForCalendar,
  });
}

export default function CalendarSidePanel({
  isOpen,
  onClose,
  type,
  data,
  onSave,
  preferences,
}) {
  const windowSize = useWindowSize();
  const isWindowSmall = windowSize === "sm" || windowSize === "md";
  // const { data: employees = [] } = useQuery(employeeQueryOptions());

  const {
    data: customers = [],
    isError: customersIsError,
    error: customersError,
  } = useQuery(customersForCalendarQueryOptions());

  const {
    data: services = [],
    isError: servicesIsError,
    error: servicesError,
  } = useQuery(servicesForCalendarQueryOptions());

  if (customersIsError || servicesIsError) {
    return (
      <ServerError error={customersError.message || servicesError.message} />
    );
  }

  return (
    <>
      <div
        className={`fixed inset-0 z-40
          ${isOpen ? "opacity-100" : "pointer-events-none opacity-0"}`}
        onClick={onClose}
        aria-hidden="true"
      />

      <aside
        className={`bg-layer_bg border-border_color fixed top-0 right-0 z-50
          h-full border-l-2 shadow-2xl transition-all duration-300 ease-in-out
          ${isOpen ? "translate-x-0" : "translate-x-full"}
          ${isWindowSmall ? "w-full" : "w-fit"}`}
      >
        {isOpen && !isWindowSmall && (
          <button
            onClick={onClose}
            className="border-border_color bg-layer_bg absolute top-4 -left-12
              rounded-full border p-1 shadow-xl transition-transform
              hover:scale-110"
            aria-label="Close panel"
          >
            <XIcon styles="size-7 fill-text_color" />
          </button>
        )}
        <div
          key={isOpen ? "open" : "closed"}
          className="bg-layer_bg h-full min-w-full"
        >
          {type === "new-booking" && (
            <NewBookingPanel
              onSave={onSave}
              onClose={onClose}
              isWindowSmall={isWindowSmall}
              categories={services}
              customers={customers}
            />
          )}
          {type === "blocked-time" && (
            <BlockedTimePanel
              blockedTime={data}
              preferences={preferences}
              onClose={onClose}
              onSubmitted={onSave}
              onDeleted={onSave}
              isWindowSmall={isWindowSmall}
            />
          )}
          {type === "edit-booking" && {}}
        </div>
      </aside>
    </>
  );
}
