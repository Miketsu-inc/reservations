import { CustomersIcon } from "@reservations/assets";
import {
  Avatar,
  Button,
  DatePicker,
  Modal,
  MultiSelect,
  Select,
  ServerError,
  Textarea,
} from "@reservations/components";
import {
  addTimeToDate,
  combineDateTimeLocal,
  invalidateLocalStorageAuth,
  timeStringFromDate,
  useToast,
} from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { useCallback, useState } from "react";
import RecurSection from "./RecurSection";

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

export default function NewBookingModal({ isOpen, onClose, onNewBooking }) {
  const { showToast } = useToast();
  const currentDate = new Date();
  const [recurData, setRecurData] = useState({
    isRecurring: false,
    frequency: "weekly",
    endDate: new Date(currentDate.setMonth(currentDate.getMonth() + 1)),
    interval: 1,
    intervalUnit: "weeks",
    days: [],
  });
  const [bookingData, setBookingData] = useState({
    date: new Date(),
    time: timeStringFromDate(new Date()).split(" ")[0],
    serviceId: null,
    customerIds: [],
    // employeeId: 0,
    merchantNote: "",
  });
  const [isDatepickerOpen, setIsDatepickerOpen] = useState(false);
  const [isSelectOpen, setIsSelectOpen] = useState(false);

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

  const selectedService = services?.find(
    (service) => service.id === bookingData.serviceId
  );
  const isGroupBooking = selectedService?.booking_type === "class";

  function updateRecurData(data) {
    setRecurData((prev) => ({ ...prev, ...data }));
  }

  function updateBookingData(data) {
    setBookingData((prev) => ({ ...prev, ...data }));
  }

  if (services.length === 1 && !bookingData.serviceId) {
    updateBookingData({ serviceId: services[0].id });
  }

  const selectedServiceDuration = useCallback(() => {
    return services.filter((service) => service.id === bookingData.serviceId)[0]
      ?.total_duration;
  }, [services, bookingData.serviceId]);

  if (customersIsError || servicesIsError) {
    return (
      <ServerError error={customersError.message || servicesError.message} />
    );
  }

  const customerOptions = customers?.map((customer) => {
    const initials =
      customer.last_name.substring(0, 1) + customer.first_name.substring(0, 1);

    return {
      value: customer.id,
      label: customer.last_name + " " + customer.first_name,
      initials: initials,
      icon: (
        <Avatar
          initials={initials}
          styles="size-6! text-[10px]! shrink-0 rounded-full!"
        />
      ),
    };
  });

  const serviceOptions = services?.map((service) => ({
    value: service.id,
    label: service.name,
  }));

  async function submitHandler() {
    const timestamp = combineDateTimeLocal(
      bookingData.date,
      bookingData.time
    ).toISOString();

    let frequency = recurData.frequency;
    let interval = recurData.interval;
    let days = recurData.days;

    if (recurData.frequency === "custom") {
      if (recurData.intervalUnit === "days") {
        frequency = "daily";
      } else if (recurData.intervalUnit === "weeks") {
        frequency = "weekly";
      }
    } else {
      interval = 1;
      days = [];
    }

    try {
      const response = await fetch(`/api/v1/merchants/bookings`, {
        method: "POST",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          customers: bookingData.customerIds.map((id) => ({
            customer_id: id,
          })),
          service_id: bookingData.serviceId,
          timestamp: timestamp,
          merchant_note: bookingData.merchantNote,
          is_recurring: recurData.isRecurring,
          recurrence_rule: {
            frequency: frequency,
            interval: Number(interval),
            weekdays: days,
            until: recurData.endDate.toISOString(),
          },
        }),
      });

      if (!response.ok) {
        const result = await response.json();
        showToast({ message: result.error.message, variant: "error" });
      } else {
        showToast({
          message: "Successfully created the booking",
          variant: "success",
        });
        onNewBooking();
        onClose();
      }
    } catch (err) {
      showToast({ message: err.message, variant: "error" });
    }
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      disableFocusTrap={true}
      suspendCloseOnClickOutside={isDatepickerOpen || isSelectOpen}
    >
      <div className="flex h-fit flex-col gap-4 p-2">
        <p className="text-2xl">New booking</p>
        <Select
          labelText="Service"
          options={serviceOptions}
          value={bookingData.serviceId}
          onSelect={(option) =>
            updateBookingData({ serviceId: option.value, customerIds: [] })
          }
          onOpenChange={(open) => setIsSelectOpen(open)}
        />
        <RecurSection
          booking={{
            start: bookingData.date,
            end: addTimeToDate(bookingData.date, 0, selectedServiceDuration()),
          }}
          recurData={recurData}
          updateRecurData={updateRecurData}
          disabled={false}
          onSelectOpenChange={(open) => setIsSelectOpen(open)}
          onDatePickerOpenChange={(open) => setIsDatepickerOpen(open)}
        />
        {isGroupBooking ? (
          <MultiSelect
            labelText="Customers"
            options={customerOptions}
            values={bookingData.customerIds}
            required={false}
            onOpenChange={(open) => setIsSelectOpen(open)}
            onSelect={(option) => updateBookingData({ customerIds: option })}
            placeholder="Select customers"
            icon={<CustomersIcon styles="size-5 text-text_color" />}
            displayText="customer"
          />
        ) : (
          <Select
            labelText="Customer"
            placeholder="Select a customer"
            options={customerOptions}
            value={bookingData.customerIds[0]}
            required={false}
            onSelect={(option) =>
              updateBookingData({ customerIds: [option.value] })
            }
            onOpenChange={(open) => setIsSelectOpen(open)}
          />
        )}
        {/* <Select
          labelText="Team member"
          options={[]}
          value={bookingData.employeeId}
          required={false}
          onSelect={(option) => updateBookingData({ employeeId: option.value })}
        /> */}
        <div className="flex flex-row items-end gap-4">
          <DatePicker
            labelText="Date"
            value={bookingData.date}
            disabledBefore={new Date()}
            onSelect={(date) => updateBookingData({ date: date })}
            onOpenChange={(open) => setIsDatepickerOpen(open)}
          />
          <input
            className="h-10 w-28 dark:scheme-dark"
            type="time"
            value={bookingData.time}
            onChange={(e) => {
              updateBookingData({ time: e.target.value });
            }}
          />
        </div>
        <Textarea
          styles="p-2 text-sm"
          name="merchantNote"
          labelText="Note"
          placeholder="Write yourself a note here..."
          required={false}
          value={bookingData.merchantNote}
          inputData={(data) => updateBookingData({ merchantNote: data.value })}
        />
        <div className="flex w-full items-center justify-end gap-2">
          <Button
            styles="py-2 px-4"
            variant="tertiary"
            name="cancelButton"
            buttonText="Cancel"
            onClick={onClose}
          />
          <Button
            styles="py-2 px-4"
            variant="primary"
            name="createButton"
            buttonText="Create"
            onClick={submitHandler}
          />
        </div>
      </div>
    </Modal>
  );
}
