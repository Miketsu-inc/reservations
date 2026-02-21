import { PlusIcon } from "@reservations/assets";
import {
  Button,
  CloseButton,
  DatePicker,
  Textarea,
} from "@reservations/components";
import {
  addTimeToDate,
  combineDateTimeLocal,
  timeStringFromDate,
  useToast,
} from "@reservations/lib";
import { useCallback, useState } from "react";
import {
  AddCustomerCard,
  ParticipantsCard,
  RecurSummaryCard,
  SelectedCustomerCard,
  ServiceCard,
} from "./BookingCards";
import CustomerProfile from "./CustomerProfile";
import CustomerSelector from "./CustomerSelector";
import NestedSidePanel from "./NestedSidePanel";
import ParticipantManager from "./ParticipantManager";
import RecurSection from "./RecurSection";
import ServiceSelector from "./ServiceSelector";

export default function NewBookingPanel({
  onSave,
  isWindowSmall,
  onClose,
  categories,
  customers,
}) {
  const [isCustomerSectionExpanded, setIsCustomerSectionExpanded] =
    useState(false);
  const [nestedPageState, setNestedPageState] = useState({
    isOpen: false,
    active: "service-selector",
  });
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
    customers: [],
    // employeeId: 0,
    merchantNote: "",
  });

  const selectedService = categories
    .flatMap((category) => category.services)
    .find((service) => service.id === bookingData.serviceId);
  const isGroupBooking = selectedService?.booking_type === "class";
  const hasSelection = bookingData?.customers.length > 0;

  const isMobilePanelActive =
    nestedPageState.active === "customer-selector" ||
    nestedPageState.active === "customer-profile";

  const isNestedPanelOpen =
    nestedPageState.isOpen && (isWindowSmall || !isMobilePanelActive);

  function updateBookingData(data) {
    setBookingData((prev) => ({ ...prev, ...data }));
  }

  function handleCollapseCustomerSection() {
    setIsCustomerSectionExpanded(false);
    updateBookingData({ customers: [] });
  }

  function handleSelectCustomers(customers) {
    updateBookingData({ customers: customers });
    if (customers.length > 0) {
      setIsCustomerSectionExpanded(false);
    } else if (!isGroupBooking) {
      handleCollapseCustomerSection();
    }
  }

  function handleRemoveCustomer() {
    updateBookingData({ customers: [] });
    setIsCustomerSectionExpanded(false);
  }

  function handleRemoveParticipant(customerToRemove) {
    const newCustomers = bookingData.customers.filter(
      (c) => c.id !== customerToRemove.id
    );
    updateBookingData({ customers: newCustomers });
  }

  function handleServiceChange(newService) {
    const isNewServiceGroup = newService.booking_type === "class";
    let updatedCustomers = bookingData.customers;

    if (!isNewServiceGroup && bookingData.customers.length > 1) {
      updatedCustomers = [];
    }

    updateBookingData({
      serviceId: newService.id,
      customers: updatedCustomers,
    });

    setNestedPageState((prev) => ({ ...prev, isOpen: false }));
  }

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
          customers: bookingData.customers,
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
        onSave();
      }
    } catch (err) {
      showToast({ message: err.message, variant: "error" });
    }
  }

  const selectedServiceDuration = useCallback(() => {
    return categories
      ?.flatMap((category) => category.services)
      ?.find((service) => service.id === bookingData.serviceId)?.duration;
  }, [categories, bookingData.serviceId]);

  const combinedStartDateTime = combineDateTimeLocal(
    bookingData.date,
    bookingData.time
  );

  return (
    <div className="flex h-full transition-all duration-300 ease-in-out">
      {!isWindowSmall && (
        <div
          className={`border-border_color overflow-hidden border-r
          transition-all duration-300 ease-in-out
          ${isCustomerSectionExpanded || hasSelection || isGroupBooking ? "w-80" : "w-40"}`}
        >
          {isGroupBooking ? (
            <ParticipantManager
              customers={customers}
              participants={bookingData.customers}
              onAdd={handleSelectCustomers}
              onRemove={handleRemoveParticipant}
              maxParticipants={selectedService?.max_participants}
              isWindowSmalll={isWindowSmall}
            />
          ) : isCustomerSectionExpanded ? (
            <CustomerSelector
              onSave={handleSelectCustomers}
              customers={customers}
              isGroupMode={false}
              walkIn={() => setIsCustomerSectionExpanded(false)}
              selected={bookingData.customers}
            />
          ) : hasSelection ? (
            <CustomerProfile
              customer={bookingData.customers[0]}
              onRemove={handleRemoveCustomer}
            />
          ) : (
            <button
              className="flex h-full w-full cursor-pointer flex-col items-start
                px-3 py-10 hover:bg-gray-200/40 dark:hover:bg-gray-700/10"
              onClick={() => {
                setIsCustomerSectionExpanded(true);
              }}
            >
              <div className="flex flex-col items-center justify-center gap-3">
                <div
                  className="bg-primary/20 text-primary flex size-14
                    items-center justify-center rounded-full"
                >
                  <PlusIcon styles="size-6" />
                </div>
                <div>
                  <p className="font-semibold">Add customer</p>
                  <span className="text-gray-400 dark:text-gray-500">
                    Or leave empty for walk-ins
                  </span>
                </div>
              </div>
            </button>
          )}
        </div>
      )}
      <div className={`${isWindowSmall ? "w-full" : "w-110"} relative`}>
        {!selectedService ? (
          <ServiceSelector
            categories={categories}
            onClose={onClose}
            isWindowSmall={isWindowSmall}
            onSelect={handleServiceChange}
            isNested={false}
          />
        ) : (
          <>
            <NestedSidePanel
              isOpen={isNestedPanelOpen}
              onBack={() =>
                setNestedPageState((prev) => ({ ...prev, isOpen: false }))
              }
              styles="size-8"
            >
              {nestedPageState.active === "service-selector" && (
                <ServiceSelector
                  categories={categories}
                  onClose={onClose}
                  isWindowSmall={isWindowSmall}
                  isNested={true}
                  onSelect={handleServiceChange}
                />
              )}
              {nestedPageState.active === "recurring" && (
                <RecurSection
                  key={nestedPageState.isOpen}
                  booking={{
                    start: combinedStartDateTime,
                    end: addTimeToDate(
                      combinedStartDateTime,
                      0,
                      selectedServiceDuration()
                    ),
                  }}
                  recurringData={recurData}
                  onSave={(recurData) => {
                    setRecurData(recurData);
                    setNestedPageState((prev) => ({ ...prev, isOpen: false }));
                  }}
                />
              )}
              {isWindowSmall && (
                <>
                  {nestedPageState.active === "customer-selector" && (
                    <CustomerSelector
                      key={isNestedPanelOpen}
                      onSave={(customer) => {
                        handleSelectCustomers(customer);
                        setNestedPageState({
                          isOpen: false,
                          active: "customer-selector",
                        });
                      }}
                      customers={customers}
                      isGroupMode={isGroupBooking}
                      walkIn={() => {
                        setNestedPageState({
                          isOpen: false,
                          active: "customer-selector",
                        });
                      }}
                      selected={bookingData.customers}
                    />
                  )}
                  ;
                  {nestedPageState.active === "customer-profile" && (
                    <CustomerProfile
                      customer={bookingData.customers[0]}
                      onRemove={handleRemoveCustomer}
                    />
                  )}
                </>
              )}
            </NestedSidePanel>
            <div
              className={`no-scrollbar relative h-full w-full overflow-y-auto
                pb-20 ${isWindowSmall ? "pt-0" : "pt-10"} md:w-110`}
            >
              {isWindowSmall && (
                <div className="flex w-full items-center justify-end px-4 pt-5">
                  <CloseButton
                    onClick={() => {
                      onClose();
                    }}
                    styles="size-8"
                  />
                </div>
              )}
              <div className="flex w-full flex-col gap-10 px-6 pb-6">
                <p className="text-2xl font-semibold">New Booking</p>
                <div className="flex w-full flex-col gap-8 md:gap-6">
                  <label className="flex flex-col gap-1">
                    <span className="">Service</span>
                    <ServiceCard
                      onClick={() =>
                        setNestedPageState({
                          isOpen: true,
                          active: "service-selector",
                        })
                      }
                      service={selectedService}
                      isGroup={isGroupBooking}
                    />
                  </label>

                  {isWindowSmall && (
                    <div className="flex flex-col gap-1">
                      <span>
                        {isGroupBooking ? "Participants" : "Customer"}
                      </span>
                      {isGroupBooking ? (
                        <ParticipantsCard
                          participants={bookingData.customers}
                          onClick={() =>
                            setNestedPageState({
                              isOpen: true,
                              active: "customer-selector",
                            })
                          }
                          maxParticipants={selectedService.max_participants}
                        />
                      ) : (
                        <div className="flex flex-col gap-2">
                          {hasSelection ? (
                            <SelectedCustomerCard
                              customer={bookingData.customers[0]}
                              onRemove={handleRemoveCustomer}
                              isGroupBooking={isGroupBooking}
                              onView={() =>
                                setNestedPageState({
                                  isOpen: true,
                                  active: "customer-profile",
                                })
                              }
                            />
                          ) : (
                            <AddCustomerCard
                              onClick={() =>
                                setNestedPageState({
                                  isOpen: true,
                                  active: "customer-selector",
                                })
                              }
                            />
                          )}
                        </div>
                      )}
                    </div>
                  )}
                  <div className="flex flex-row items-end gap-2">
                    <DatePicker
                      styles=""
                      labelText="Date"
                      value={bookingData.date}
                      disabledBefore={new Date()}
                      onSelect={(date) => updateBookingData({ date: date })}
                    />
                    <input
                      className="bg-layer_bg border-input_border_color
                        focus:border-primary focus:ring-primary/30
                        disabled:text-text_color/70
                        disabled:border-input_border_color/60 w-32 resize-none
                        rounded-lg border px-2 py-1.5 placeholder-stone-500
                        outline-hidden transition-[border-color,box-shadow]
                        duration-150 ease-in-out focus:ring-4 dark:scheme-dark"
                      type="time"
                      value={bookingData.time}
                      onChange={(e) => {
                        updateBookingData({ time: e.target.value });
                      }}
                    />
                  </div>
                  <RecurSummaryCard
                    recurData={recurData}
                    booking={{
                      start: combinedStartDateTime,
                      end: addTimeToDate(
                        combinedStartDateTime,
                        0,
                        selectedServiceDuration()
                      ),
                    }}
                    onClick={() =>
                      setNestedPageState({ isOpen: true, active: "recurring" })
                    }
                  />
                  <Textarea
                    styles="p-2 text-sm h-24"
                    name="merchantNote"
                    labelText="Note"
                    placeholder="Write yourself a note here..."
                    required={false}
                    value={bookingData.merchantNote}
                    inputData={(data) =>
                      updateBookingData({ merchantNote: data.value })
                    }
                  />
                </div>
              </div>
              <div
                className="border-border_color bg-layer_bg items center fixed
                  bottom-0 flex w-full border-t px-6 py-4 md:w-110"
              >
                <Button
                  styles="py-2 px-4 w-full"
                  variant="primary"
                  name="createButton"
                  buttonText="Create"
                  onClick={submitHandler}
                />
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
