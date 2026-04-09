import {
  BackArrowIcon,
  BanIcon,
  CalendarIcon,
  ClockIcon,
  PlusIcon,
  RefreshIcon,
  TickIcon,
  WalkingIcon,
} from "@reservations/assets";
import {
  Button,
  CloseButton,
  DatePicker,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
  Textarea,
} from "@reservations/components";
import {
  combineDateTimeLocal,
  timeStringFromDate,
  useToast,
} from "@reservations/lib";
import { useState } from "react";
import {
  AddCustomerCard,
  ParticipantsCard,
  SelectedCustomerCard,
  ServiceCard,
} from "./BookingCards";
import CustomerProfile from "./CustomerProfile";
import CustomerSelector from "./CustomerSelector";
import NestedSidePanel from "./NestedSidePanel";
import ParticipantManager from "./ParticipantManager";
import ServiceSelector from "./ServiceSelector";

function monthDateFormat(date) {
  return date.toLocaleDateString([], {
    month: "short",
    day: "numeric",
    weekday: "long",
  });
}

export default function EditBookingPanel({
  originalBookingData,
  customers,
  categories,
  isWindowSmall,
  onClose,
  onSave,
  onSoftUpdate,
  onOpenCancelModal,
  onOpenRecurModal,
  preferences,
}) {
  const { showToast } = useToast();
  const isPastBooking = new Date(originalBookingData.end) <= new Date();
  const isBookingCompleted =
    originalBookingData.extendedProps.booking_status === "completed";
  const isRecurring = originalBookingData.extendedProps.is_recurring;

  const [isCustomerSectionExpanded, setIsCustomerSectionExpanded] =
    useState(false);
  const [nestedPageState, setNestedPageState] = useState({
    isOpen: false,
    active: "service-selector",
  });

  const isMobilePanelActive =
    nestedPageState.active === "customer-selector" ||
    nestedPageState.active === "customer-profile" ||
    nestedPageState.active === "participant-manager";

  const isNestedPanelOpen =
    nestedPageState.isOpen && (isWindowSmall || !isMobilePanelActive);

  const mappedParticipants =
    originalBookingData.extendedProps?.participants.map((participant) => {
      const customerData = customers.find(
        (c) => c.customer_id === participant.customer_id
      );

      return {
        first_name: customerData.first_name || participant.first_name,
        last_name: customerData.last_name || participant.last_name,
        birthday: customerData.birthday || "",
        is_dummy: customerData.is_dummy || false,
        last_visited: customerData.last_visited || null,
        phone_number: customerData.phone_number || "",
        email: customerData.email || "",
        customer_id: participant.customer_id,
        participant_id: participant.id,
        status: participant.status,
        customer_note:
          participant.customer_note ||
          "this is a great booking and i cant wait to fet to do it tofetjer with you bevause your are one of  the best",
      };
    });

  const [bookingData, setBookingData] = useState({
    date: originalBookingData.start,
    time: timeStringFromDate(originalBookingData.start).split(" ")[0],
    serviceId: originalBookingData.extendedProps.service_id,
    bookingStatus: originalBookingData.extendedProps.booking_status,
    participants: mappedParticipants,
    merchantNote: originalBookingData.extendedProps.merchant_note || "",
  });

  const hasSelection = bookingData.participants.length > 0;

  const selectedService = categories
    .flatMap((category) => category.services)
    .find((service) => service.id === bookingData.serviceId);

  const isGroupBooking = selectedService?.booking_type !== "appointment";

  function updateBookingData(newData) {
    if (isBookingCompleted) return;
    setBookingData((prev) => ({ ...prev, ...newData }));
  }

  function handleSelectCustomers(newCustomers) {
    updateBookingData({ participants: newCustomers });
    if (newCustomers.length > 0) {
      setIsCustomerSectionExpanded(false);
    } else if (!isGroupBooking) {
      setIsCustomerSectionExpanded(false);
    }
  }

  function handleRemoveCustomer() {
    updateBookingData({ participants: [] });
    setIsCustomerSectionExpanded(false);
  }

  function handleRemoveParticipant(customerToRemove) {
    const newCustomers = bookingData.participants.filter(
      (c) => c.customer_id !== customerToRemove.customer_id
    );
    updateBookingData({ participants: newCustomers });
  }

  function handleServiceChange(newService) {
    const isNewServiceGroup = newService?.booking_type === "class";
    let updatedCustomers = bookingData.participants;

    if (!isNewServiceGroup && bookingData.participants.length > 1) {
      updatedCustomers = [];
    }

    updateBookingData({
      serviceId: newService.id,
      participants: updatedCustomers,
    });
    setNestedPageState((prev) => ({ ...prev, isOpen: false }));
  }

  async function handleSave(option = "this") {
    if (isBookingCompleted) {
      showToast({
        message: "You cant update a completed booking",
        variant: "error",
      });
      return;
    }

    const timestamp = combineDateTimeLocal(
      bookingData.date,
      bookingData.time
    ).toISOString();

    const customerInput = bookingData.participants.map((p) => ({
      id: p.customer_id,
      first_name: p.first_name,
      last_name: p.last_name,
      email: p.email,
      phone_number: p.phone_number,
    }));

    try {
      const response = await fetch(
        `/api/v1/bookings/merchant/${originalBookingData.extendedProps.id}`,
        {
          method: "PATCH",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify({
            customers: customerInput,
            service_id: bookingData.serviceId,
            timestamp: timestamp,
            merchant_note: bookingData.merchantNote,
            booking_status: bookingData.bookingStatus,
            update_all_future: option !== "this",
          }),
        }
      );

      if (!response.ok) {
        const result = await response.json();
        showToast({ message: result.error.message, variant: "error" });
      } else {
        showToast({
          message: "Successfully updated the booking",
          variant: "success",
        });
        onSave();
      }
    } catch (err) {
      showToast({ message: err.message, variant: "error" });
    }
  }

  async function handleStatusChange(participantId, newStatus, oldStatus) {
    setBookingData((prev) => {
      const updatedParticipants = prev.participants.map((p) =>
        p.participant_id === participantId ? { ...p, status: newStatus } : p
      );
      return { ...prev, participants: updatedParticipants };
    });

    try {
      const response = await fetch(
        `/api/v1/bookings/merchant/${originalBookingData.extendedProps.id}/participant/${participantId}`,
        {
          method: "PATCH",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify({
            status: newStatus,
          }),
        }
      );

      if (!response.ok) {
        const result = await response.json();
        showToast({ message: result.error.message, variant: "error" });
      } else {
        showToast({
          message: "Successfully created the booking",
          variant: "success",
        });
        onSoftUpdate();
      }
    } catch (err) {
      setBookingData((prev) => {
        const revertedParticipants = prev.participants.map((p) =>
          p.participant_id === participantId ? { ...p, status: oldStatus } : p
        );
        return { ...prev, participants: revertedParticipants };
      });

      showToast({ message: err.message, variant: "error" });
    }
  }

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
              participants={bookingData.participants}
              maxParticipants={selectedService.max_participants}
              isWindowSmall={isWindowSmall}
              disabled={isPastBooking}
              onAdd={handleSelectCustomers}
              onRemove={handleRemoveParticipant}
              onStatusChange={handleStatusChange}
            />
          ) : hasSelection ? (
            <CustomerProfile
              customer={bookingData.participants[0]}
              onRemove={handleRemoveCustomer}
              disabled={isPastBooking}
            />
          ) : isPastBooking ? (
            <div
              className="flex flex-col items-center justify-center gap-3 px-3
                py-10 opacity-70"
            >
              <div
                className="flex size-16 items-center justify-center rounded-full
                  bg-gray-200 dark:bg-gray-400/20"
              >
                <WalkingIcon styles="fill-gray-300 size-6" />
              </div>
              <div className="text-center">
                <p className="font-semibold">Walk-in Customer</p>
                <button
                  className="border-primary hover:bg-primary/10 mt-2 rounded-md
                    p-2 text-xs"
                >
                  Assing Customer
                </button>
              </div>
            </div>
          ) : !isPastBooking && isCustomerSectionExpanded ? (
            <CustomerSelector
              onSave={handleSelectCustomers}
              customers={customers}
              isGroupMode={false}
              walkIn={() => setIsCustomerSectionExpanded(false)}
              selected={bookingData.participants}
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
          {isWindowSmall && (
            <>
              {nestedPageState.active === "customer-selector" && (
                <CustomerSelector
                  key={nestedPageState.isOpen}
                  onSave={(selectedCustomers) => {
                    handleSelectCustomers(selectedCustomers);
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
                  selected={bookingData.participants}
                />
              )}
              {nestedPageState.active === "customer-profile" && (
                <CustomerProfile
                  customer={bookingData.participants[0]}
                  onRemove={handleRemoveCustomer}
                  disabled={isPastBooking}
                />
              )}
              {nestedPageState.active === "participant-manager" && (
                <ParticipantManager
                  customers={customers}
                  participants={bookingData.participants}
                  maxParticipants={selectedService.max_participants}
                  isWindowSmall={isWindowSmall}
                  disabled={isPastBooking}
                  onAdd={handleSelectCustomers}
                  onRemove={handleRemoveParticipant}
                  onStatusChange={handleStatusChange}
                />
              )}
            </>
          )}
        </NestedSidePanel>
        <div
          className="no-scrollbar relative flex h-full w-full flex-col
            overflow-y-auto"
        >
          <BookingHeader
            originalBookingData={originalBookingData}
            bookingData={bookingData}
            isBookingCompleted={isBookingCompleted}
            isPastBooking={isPastBooking}
            isGroupBooking={isGroupBooking}
            isWindowSmall={isWindowSmall}
            selectedService={selectedService}
            updateBookingData={updateBookingData}
            onCancel={onOpenCancelModal}
            onClose={onClose}
          />
          <div className="flex w-full flex-col gap-8 px-6 pt-9 pb-28">
            <label className="flex flex-col gap-1">
              <span className="">Service</span>
              <ServiceCard
                onClick={() => {
                  if (!isPastBooking) {
                    setNestedPageState({
                      isOpen: true,
                      active: "service-selector",
                    });
                  }
                }}
                service={selectedService}
                isGroup={isGroupBooking}
                disabled={isPastBooking}
              />
            </label>
            {isWindowSmall && (
              <div className="flex flex-col gap-1">
                <span className="">
                  {isGroupBooking ? "Participants" : "Customer"}
                </span>
                {isGroupBooking ? (
                  isPastBooking && !hasSelection ? (
                    <div
                      className="border-input_border_color/70 flex items-center
                        gap-3 rounded-md border p-4"
                    >
                      <span className="text-gray-800 dark:text-gray-300">
                        No participants attended
                      </span>
                    </div>
                  ) : (
                    <ParticipantsCard
                      participants={bookingData.participants}
                      onClick={() =>
                        setNestedPageState({
                          isOpen: true,
                          active: "participant-manager",
                        })
                      }
                      maxParticipants={selectedService.max_participants}
                    />
                  )
                ) : (
                  <div className="flex flex-col gap-2">
                    {hasSelection ? (
                      <SelectedCustomerCard
                        customer={bookingData.participants[0]}
                        onRemove={handleRemoveCustomer}
                        isGroupBooking={isGroupBooking}
                        onView={() =>
                          setNestedPageState({
                            isOpen: true,
                            active: "customer-profile",
                          })
                        }
                        disabled={isPastBooking}
                      />
                    ) : isPastBooking ? (
                      <div
                        className="border-input_border_color/70 flex
                          items-center gap-6 rounded-md border p-4 px-6"
                      >
                        <div
                          className="flex size-12 items-center justify-center
                            rounded-full bg-gray-200 dark:bg-gray-400/20"
                        >
                          <WalkingIcon styles="fill-gray-300 size-6" />
                        </div>
                        <span className="text-gray-800 dark:text-gray-300">
                          Walk-in Customer
                        </span>
                      </div>
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
            {isPastBooking ? (
              <div
                className="border-input_border_color flex items-center
                  justify-between rounded-lg border px-4 py-3"
              >
                <div className="flex items-center gap-2">
                  <CalendarIcon styles="size-5" />
                  <span>{monthDateFormat(originalBookingData.start)}</span>
                </div>
                <div className="flex items-center gap-2">
                  <ClockIcon styles="size-4 fill-text_color" />
                  <span>{`${timeStringFromDate(originalBookingData.start, preferences?.time_format)} - ${timeStringFromDate(originalBookingData.end, preferences?.time_format)}`}</span>{" "}
                </div>
              </div>
            ) : (
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
            )}

            {isRecurring && (
              <div
                className="border-input_border_color flex items-center gap-2
                  rounded-lg border p-3"
              >
                <RefreshIcon styles="size-5" />
                <span>Part of repeating series</span>
              </div>
            )}
            {!isPastBooking ? (
              <Textarea
                styles="p-2 text-sm h-24"
                name="merchantNote"
                labelText="Merchant Note"
                required={false}
                placeholder="Add notes about this appointment..."
                value={bookingData.merchantNote}
                inputData={(data) =>
                  updateBookingData({ merchantNote: data.value })
                }
              />
            ) : bookingData.merchantNote ? (
              <div className="flex flex-col gap-3">
                <span className="text-sm font-medium">Your Note</span>
                <div className="text-sm text-gray-600 dark:text-gray-300">
                  {bookingData.merchantNote}
                </div>
              </div>
            ) : (
              <></>
            )}
          </div>
        </div>
        {!isBookingCompleted && (
          <div
            className="border-border_color bg-layer_bg items center fixed
              bottom-0 flex w-full border-t px-6 py-4 md:w-110"
          >
            <Button
              styles="py-2 px-4 w-full"
              variant="primary"
              name="saveButton"
              buttonText="Save"
              onClick={() => {
                if (isRecurring) {
                  onOpenRecurModal(handleSave);
                } else {
                  handleSave();
                }
              }}
            />
          </div>
        )}
      </div>
    </div>
  );
}

function BookingHeader({
  originalBookingData,
  bookingData,
  isBookingCompleted,
  isPastBooking,
  isGroupBooking,
  isWindowSmall,
  selectedService,
  updateBookingData,
  onCancel,
  onClose,
}) {
  return (
    <>
      {isWindowSmall && (
        <div
          className="flex w-full items-center justify-end px-4 pt-4"
          style={{ backgroundColor: selectedService.color }}
        >
          <CloseButton onClick={onClose} styles="size-8 fill-white" />
        </div>
      )}
      <div
        className="sticky top-0 z-10 -mt-1 flex w-full items-center
          justify-between rounded-b-xl px-5 pt-4 pb-5"
        style={{ backgroundColor: selectedService.color }}
      >
        <span className="text-[20px] font-medium text-white opacity-100!">
          {monthDateFormat(bookingData.date)}
        </span>

        {isBookingCompleted ? (
          <div
            className="flex items-center gap-2 rounded-md border border-white/50
              px-3 py-2 text-sm font-medium text-white shadow-lg"
          >
            <div className="rounded-full border-white">
              <TickIcon styles="size-5" />
            </div>
            <span>Completed</span>
          </div>
        ) : (
          <Popover>
            <PopoverTrigger asChild>
              <button
                className="inline-flex items-center gap-3 rounded-md border
                  border-white/50 px-3 py-2 text-sm font-medium text-white
                  shadow-lg"
              >
                <span>{bookingData.bookingStatus}</span>
                <BackArrowIcon
                  styles="-rotate-90 size-4 mt-0.5 stroke-white shadow-lg
                    transition-all duration-200"
                />
              </button>
            </PopoverTrigger>
            <PopoverContent styles="w-auto" align="start">
              <div
                className="flex w-32 flex-col *:flex *:w-full *:flex-row
                  *:items-center *:rounded-lg *:p-2"
              >
                {isPastBooking && bookingData.bookingStatus !== "completed" && (
                  <PopoverClose asChild>
                    <button
                      className="hover:bg-hvr_gray gap-2"
                      onClick={() =>
                        updateBookingData({ bookingStatus: "completed" })
                      }
                    >
                      <TickIcon styles="size-6 fill-text_color" />
                      Completed
                    </button>
                  </PopoverClose>
                )}

                {isPastBooking &&
                  !isGroupBooking &&
                  bookingData.bookingStatus !== "no-show" && (
                    <PopoverClose asChild>
                      <button
                        className="hover:bg-hvr_gray gap-3 text-red-600
                          dark:text-red-800"
                        onClick={() =>
                          updateBookingData({ bookingStatus: "no-show" })
                        }
                      >
                        <BanIcon styles="size-5" />
                        No-show
                      </button>
                    </PopoverClose>
                  )}
                {bookingData.bookingStatus !== "booked" && (
                  <PopoverClose asChild>
                    <button
                      className="hover:bg-hvr_gray cursor-pointer gap-3"
                      onClick={() =>
                        updateBookingData({ bookingStatus: "booked" })
                      }
                    >
                      <CalendarIcon styles="size-5 text-text_color" />
                      Booked
                    </button>
                  </PopoverClose>
                )}
                {!isPastBooking && (
                  <PopoverClose asChild>
                    <button className="hover:bg-hvr_gray">Reschedule</button>
                  </PopoverClose>
                )}
                {!isPastBooking && (
                  <PopoverClose asChild>
                    <button
                      className="hover:bg-hvr_gray text-red-600
                        dark:text-red-800"
                      onClick={() =>
                        onCancel(originalBookingData.extendedProps)
                      }
                    >
                      Cancel
                    </button>
                  </PopoverClose>
                )}
              </div>
            </PopoverContent>
          </Popover>
        )}
      </div>
    </>
  );
}
