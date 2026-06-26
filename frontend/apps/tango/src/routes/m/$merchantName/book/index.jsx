import { ArrowLeft01Icon, ArrowRight02Icon } from "@hugeicons/core-free-icons";
import { Button, Icon, ServerError } from "@reservations/components";
import {
  formatDuration,
  formatTimeRange,
  getDisplayPrice,
  invalidateLocalStorageAuth,
  useToast,
  useWindowSize,
} from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import AppointmentTimeSelectionStep from "./-components/AppointmentTimeSelectionStep";
import BookingSummary from "./-components/BookingSummary";
import EmployeeSelectionStep from "./-components/EmployeeSelectionStep";
import ServiceSelectionStep from "./-components/ServiceSelectionStep";

function validateSearch(search) {
  const isValidType = search.type === "appointment" || search.type === "class";

  const keepServiceData = isValidType && search.serviceId;

  return {
    type: keepServiceData ? search.type : undefined,
    serviceId: keepServiceData ? search.serviceId : undefined,
    employeeId: search.employeeId,
    locationId: search.locationId,
  };
}

async function fetchSummaryInfo(
  merchantName,
  locationId,
  serviceId,
  employeeId
) {
  const params = new URLSearchParams();
  if (serviceId) params.append("serviceId", serviceId);
  if (employeeId != "no-pref" && employeeId)
    params.append("employeeId", employeeId);

  const queryString = params.toString();
  const url = `/api/v1/public/merchants/${merchantName}/locations/${locationId}/summary${queryString ? `?${queryString}` : ""}`;

  const response = await fetch(url, {
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

function bookingSummaryQueryOptions(
  merchantName,
  locationId,
  serviceId,
  employeeId
) {
  return queryOptions({
    queryKey: [
      "booking-summary",
      merchantName,
      locationId,
      serviceId,
      employeeId,
    ],
    queryFn: () =>
      fetchSummaryInfo(merchantName, locationId, serviceId, employeeId),
  });
}

export const Route = createFileRoute("/m/$merchantName/book/")({
  component: BookingFLow,
  validateSearch: validateSearch,
});

function BookingFLow() {
  const { merchantName } = Route.useParams();
  const search = Route.useSearch();
  const navigate = Route.useNavigate({ from: Route.fullPath });
  const router = useRouter();
  const { showToast } = useToast();
  const [selectedSummary, setSelectedSummary] = useState({
    service: null,
    employee: null,
    time: null,
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const windowSize = useWindowSize();
  const isWindowSmall = windowSize === "sm" || windowSize === "md";

  const [isScrolled, setIsScrolled] = useState(false);

  // detect when the user scrolled 50px down
  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 50);
    };
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  const currentStep = !search.serviceId
    ? "service"
    : !search.employeeId
      ? "employee"
      : "time";

  const stepTitles = {
    service: "Select a Service",
    employee: "Select a Professional",
    time: search.type === "class" ? "Select a Class" : "Select Date & Time",
  };

  const canContinue =
    (currentStep === "service" && selectedSummary.service) ||
    (currentStep === "employee" && selectedSummary.employee) ||
    (currentStep === "time" &&
      selectedSummary.time?.time &&
      selectedSummary.time?.date);

  const {
    data: fetchedSummary,
    isLoading,
    isError,
    error,
  } = useQuery(
    bookingSummaryQueryOptions(
      merchantName,
      search.locationId,
      search.serviceId,
      search.employeeId
    )
  );

  if (isError) {
    return <ServerError error={error.message} />;
  }

  function updateBookingDetails(key, data) {
    setSelectedSummary((prev) => ({ ...prev, [key]: data }));
  }

  function handleBack() {
    // rely on only what is inside the URL
    setSelectedSummary({
      service: null,
      employee: null,
      time: null,
    });
    router.history.back();
  }

  function handleContinue() {
    if (currentStep === "service" && selectedSummary.service) {
      navigate({
        search: (prev) => ({
          ...prev,
          serviceId: selectedSummary.service.id,
          type: selectedSummary.service.booking_type,
        }),
      });
    } else if (currentStep === "employee" && selectedSummary.employee) {
      navigate({
        search: (prev) => ({
          ...prev,
          employeeId: selectedSummary.employee.id,
        }),
      });
    } else if (currentStep === "time") {
      onSubmitHandler();
    }
  }

  async function onSubmitHandler() {
    if (!selectedSummary.time?.time || !selectedSummary.time?.date) {
      showToast({
        message: "Please select a date and time",
        variant: "error",
      });
      return;
    }

    const date = new Date(selectedSummary.time.date);

    const [hours, minutes] = selectedSummary.time.time.split(":").map(Number);
    date.setHours(hours, minutes, 0, 0);
    const timeStamp = date.toISOString();

    setIsSubmitting(true);

    try {
      const response = await fetch("/api/v1/public/bookings", {
        method: "POST",
        headers: {
          "Content-type": "application/json; charset=UTF-8",
        },
        body: JSON.stringify({
          merchant_name: merchantName,
          service_id: search.serviceId,
          location_id: search.locationId,
          timeStamp: timeStamp,
          customer_note: selectedSummary.time.customer_note,
        }),
      });

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);

        if (response.status === 401) {
          navigate({
            from: Route.fullPath,
            to: "/login",
            search: {
              redirect: router.history.location.href,
            },
          });
        }

        const result = await response.json();
        showToast({
          message: result.error.message,
          variant: "error",
        });
      } else {
        navigate({
          from: Route.fullPath,
          to: "completed",
        });
      }
    } catch (err) {
      showToast({
        message: err.message,
        variant: "error",
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div
      className="bg-bg_color relative flex min-h-screen w-full flex-col gap-6"
    >
      <div
        className={`bg-bg_color sticky top-0 z-50 flex items-center
          justify-between py-3 ${
            isScrolled ? "border-border_color border-b shadow-sm" : "border-0"
          }`}
      >
        <div
          className="mx-auto flex w-full max-w-7xl items-center gap-4 px-4
            md:px-8"
        >
          <button
            className="hover:bg-hvr_gray/20 md:bg-layer_bg border-border_color
              flex h-fit w-fit rounded-full p-1 md:border md:p-2 md:shadow-sm"
            onClick={handleBack}
          >
            <Icon icon={ArrowLeft01Icon} styles="text-text_color size-7" />
          </button>
          <div
            className={`text-2xl font-semibold ${
              isScrolled ? "block" : "hidden"
            }`}
          >
            {stepTitles[currentStep]}
          </div>
        </div>
      </div>
      <div className="relative flex justify-center">
        <div
          className="flex w-full px-6 md:gap-10 md:px-8 lg:max-w-6xl lg:gap-12"
        >
          <div className="flex-1 pb-24">
            {currentStep === "service" && (
              <ServiceSelectionStep
                merchantName={merchantName}
                locationId={search.locationId}
                employeeId={search.employeeId}
                onServiceSelect={(data) => {
                  updateBookingDetails("service", data);
                }}
                employee={{
                  first_name: fetchedSummary?.employee_first_name,
                  last_name: fetchedSummary?.employee_last_name,
                }}
              />
            )}

            {currentStep === "employee" && (
              <EmployeeSelectionStep
                merchantName={merchantName}
                locationId={search.locationId}
                serviceId={search.serviceId}
                onSelect={(data) => {
                  updateBookingDetails("employee", data);
                }}
                onAutoSkip={(data) => {
                  updateBookingDetails("employee", data);
                  navigate({
                    search: (prev) => ({ ...prev, employeeId: data.id }),
                    replace: true,
                  });
                }}
              />
            )}

            {currentStep === "time" && search.type === "appointment" && (
              <AppointmentTimeSelectionStep
                merchantName={merchantName}
                locationId={search.locationId}
                serviceId={search.serviceId}
                employeeId={search.employeeId}
                onSelect={(data) => {
                  updateBookingDetails("time", data);
                }}
                employee={{
                  first_name: fetchedSummary?.employee_first_name,
                  last_name: fetchedSummary?.employee_last_name,
                }}
              />
            )}

            {currentStep === "time" && search.type === "class" && (
              <div className="">
                <h1 className="text-3xl font-bold">Select a Class</h1>
              </div>
            )}
          </div>

          {!isWindowSmall && (
            <div className="w-110">
              <BookingSummary
                fetchedSummary={fetchedSummary}
                selectedSummary={selectedSummary}
                isLoading={isLoading}
                onContinue={handleContinue}
                canContinue={canContinue}
                isSubmitting={isSubmitting}
                currentStep={currentStep}
              />
            </div>
          )}
        </div>
      </div>
      {isWindowSmall && (
        <div
          className={`bg-layer_bg border-border_color fixed bottom-0 z-20 flex
          w-full items-center justify-between border-t-2 px-6 py-3
          transition-transform duration-300 ease-in-out
          ${canContinue ? "translate-y-0" : "translate-y-full"}`}
        >
          <div className="flex flex-col">
            <span className="text-lg font-medium">
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
            {selectedSummary.time?.time ? (
              <span className="text-sm text-gray-500 dark:text-gray-400">
                {formatTimeRange(
                  selectedSummary.time?.time,
                  fetchedSummary.total_duration
                )}
              </span>
            ) : (
              <span className="text-sm text-gray-500 dark:text-gray-400">
                {fetchedSummary?.service_name
                  ? formatDuration(fetchedSummary?.total_duration)
                  : formatDuration(selectedSummary.service?.total_duration)}
              </span>
            )}
          </div>
          <Button
            styles="w-fit px-4 py-2"
            buttonText={currentStep === "time" ? "Book" : "Continue"}
            variant="primary"
            onClick={handleContinue}
            disabled={!canContinue}
            isLoading={isSubmitting}
          >
            <Icon icon={ArrowRight02Icon} styles="size-5" />
          </Button>
        </div>
      )}
    </div>
  );
}
