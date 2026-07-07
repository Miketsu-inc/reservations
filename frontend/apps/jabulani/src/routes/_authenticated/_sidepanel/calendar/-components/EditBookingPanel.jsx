import {
  ArrowLeft01Icon,
  ArrowReloadHorizontalIcon,
  Calendar02Icon,
  Clock01Icon,
  Tick02Icon,
  UnavailableIcon,
} from "@hugeicons/core-free-icons";
import {
  Avatar,
  Button,
  CloseButton,
  DatePicker,
  Icon,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
  Select,
  Textarea,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  combineDateTimeLocal,
  DEFAULT_SERVICE_COLOR,
  timeStringFromDate,
  useToast,
} from "@reservations/lib";
import { useState } from "react";
import { ServiceCard } from "./BookingCards";
import CancelBookingModal from "./CancelBookingModal";
import CustomerProfile from "./CustomerProfile";
import CustomerSelector from "./CustomerSelector";
import NestedSidePanel from "./NestedSidePanel";
import {
  MobileParticipantSection,
  ParticipantSideBar,
} from "./ParticipantLogic";
import ParticipantManager from "./ParticipantManager";
import UpdateRecurringModal from "./UpdateRecurringModal";

function monthDateFormat(date) {
  return date.toLocaleDateString([], {
    month: "short",
    day: "numeric",
    weekday: "long",
  });
}

function resolveServiceData(service, booking) {
  return {
    booking_type: service.booking_type ?? booking.booking_type,
    color: service.color ?? DEFAULT_SERVICE_COLOR,
    name: service.name ?? booking.service_name,
    duration: service.duration ?? booking.duration,
    max_participants: service.max_participants ?? booking.max_participants,
    price: service.price ?? booking.price,
    price_type: service.price_type ?? booking.price_type,
  };
}

export default function EditBookingPanel({
  originalBookingData,
  customers,
  categories,
  isWindowSmall,
  onClose,
  onSave,
  onSoftUpdate,
  preferences,
  team,
}) {
  const { showToast } = useToast();
  const { merchantId } = useAuth();
  const isPastBooking = new Date(originalBookingData.end) <= new Date();
  const isBookingCompleted =
    originalBookingData.extendedProps.booking_status === "completed";
  const isRecurring = originalBookingData.extendedProps.is_recurring;

  const [isCustomerSectionExpanded, setIsCustomerSectionExpanded] =
    useState(false);
  const [nestedPageState, setNestedPageState] = useState({
    isOpen: false,
    active: "customer-selector",
  });
  const [isRecurModalOpen, setIsRecurModalOpen] = useState(false);
  const [isCancelModalOpen, setIsCancelModalOpen] = useState(false);

  const isNestedPanelOpen = nestedPageState.isOpen && isWindowSmall;

  const mappedParticipants =
    originalBookingData.extendedProps?.participants.map((participant) => {
      const customerData =
        customers.find((c) => c.customer_id === participant.customer_id) || {};

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
        customer_note: participant.customer_note || "",
      };
    });

  const [bookingData, setBookingData] = useState({
    date: originalBookingData.start,
    time: timeStringFromDate(originalBookingData.start).split(" ")[0],
    serviceId: originalBookingData.extendedProps.service_id,
    employeeId: originalBookingData.extendedProps.employee_id,
    bookingStatus: originalBookingData.extendedProps.booking_status,
    participants: mappedParticipants,
    merchantNote: originalBookingData.extendedProps.merchant_note || "",
  });

  const hasSelection = bookingData.participants.length > 0;

  const selectedService = categories
    .flatMap((category) => category.services)
    .find((service) => service.id === bookingData.serviceId);

  const isGroupBooking =
    (selectedService?.booking_type ??
      originalBookingData.extendedProps.booking_type) !== "appointment";

  const teamOptions = team?.map((member) => ({
    value: member.id,
    label: member.first_name + " " + member.last_name,
    icon: (
      <Avatar
        styles="size-6! rounded-full! text-[10px]!"
        initials={`${member.first_name[0]}${member.last_name[0]}`}
      />
    ),
  }));

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

  async function handleSave(option = "this") {
    if (isBookingCompleted) {
      showToast({
        message: "You cant update a completed booking",
        variant: "error",
      });
      return;
    }

    if (
      !isGroupBooking &&
      isRecurring &&
      bookingData.participants.length === 0
    ) {
      showToast({
        message: "You must have a participant for a repeating 1-on-1 booking",
        variant: "error",
      });
      return;
    }

    const timestamp = combineDateTimeLocal(
      bookingData.date,
      bookingData.time
    ).toISOString();

    const customerInput = bookingData.participants.map((p) => {
      const payload = {
        first_name: p.first_name,
        last_name: p.last_name,
        email: p.email,
        phone_number: p.phone_number,
      };

      if (!p.isNewCustomer) {
        payload.id = p.customer_id;
      }

      return payload;
    });

    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/bookings/${originalBookingData.extendedProps.id}`,
        {
          method: "PATCH",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify({
            customers: customerInput,
            timestamp: timestamp,
            merchant_note: bookingData.merchantNote,
            employee_id: bookingData.employeeId,
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

  function updateParticipantStatus(id, status) {
    setBookingData((prev) => ({
      ...prev,
      participants: prev.participants.map((p) =>
        p.participant_id === id ? { ...p, status } : p
      ),
    }));
  }

  async function handleStatusChange(participantId, newStatus, oldStatus) {
    updateParticipantStatus(participantId, newStatus);

    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/bookings/${originalBookingData.extendedProps.id}/participant/${participantId}`,
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
      updateParticipantStatus(participantId, oldStatus);
      showToast({ message: err.message, variant: "error" });
    }
  }

  return (
    <div className="flex h-full transition-all duration-300 ease-in-out">
      <UpdateRecurringModal
        isOpen={isRecurModalOpen}
        onClose={() => setIsRecurModalOpen(false)}
        onSave={(option) => handleSave(option)}
      />
      <CancelBookingModal
        bookingId={originalBookingData.extendedProps.id}
        isOpen={isCancelModalOpen}
        onClose={() => setIsCancelModalOpen(false)}
        onDeleted={() => {
          setIsCancelModalOpen(false);
          onSave();
        }}
        isRecurring={isRecurring}
      />

      <ParticipantSideBar
        isGroupBooking={isGroupBooking}
        hasSelection={hasSelection}
        isPastBooking={isPastBooking}
        isExpanded={isCustomerSectionExpanded}
        setIsExpanded={setIsCustomerSectionExpanded}
        customers={customers}
        selectedCustomers={bookingData.participants}
        maxParticipants={
          selectedService?.max_participants ??
          originalBookingData.extendedProps.max_participants
        }
        isWindowSmall={isWindowSmall}
        onAddCustomer={handleSelectCustomers}
        onRemoveCustomer={handleRemoveCustomer}
        onRemoveParticipant={handleRemoveParticipant}
        onStatusChange={handleStatusChange}
      />

      <div className={`${isWindowSmall ? "w-full" : "w-110"} relative`}>
        <NestedSidePanel
          isOpen={isNestedPanelOpen}
          onBack={() =>
            setNestedPageState((prev) => ({ ...prev, isOpen: false }))
          }
          styles="size-8"
        >
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
                  isWindowSmall={isWindowSmall}
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
                  maxParticipants={
                    selectedService?.max_participants ??
                    originalBookingData.extendedProps.max_participants
                  }
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
            bookingData={bookingData}
            isBookingCompleted={isBookingCompleted}
            isPastBooking={isPastBooking}
            isGroupBooking={isGroupBooking}
            isWindowSmall={isWindowSmall}
            serviceColor={selectedService?.color ?? DEFAULT_SERVICE_COLOR}
            updateBookingData={updateBookingData}
            onCancel={() => setIsCancelModalOpen(true)}
            onClose={onClose}
          />
          <div className="flex w-full flex-col gap-8 px-6 pt-9 pb-28">
            <label className="flex flex-col gap-1">
              <span>Service</span>
              <ServiceCard
                service={resolveServiceData(
                  selectedService,
                  originalBookingData.extendedProps
                )}
                isGroup={isGroupBooking}
                disabled={true}
                onClick={() =>
                  setNestedPageState({ isOpen: true, active: "service-editor" })
                }
              />
            </label>
            <MobileParticipantSection
              isWindowSmall={isWindowSmall}
              isGroupBooking={isGroupBooking}
              hasSelection={hasSelection}
              isPastBooking={isPastBooking}
              selectedCustomers={bookingData.participants}
              maxParticipants={selectedService?.max_participants}
              onRemoveCustomer={handleRemoveCustomer}
              onOpenCustomerSelector={() =>
                setNestedPageState({
                  isOpen: true,
                  active: "customer-selector",
                })
              }
              onOpenProfile={() =>
                setNestedPageState({ isOpen: true, active: "customer-profile" })
              }
              onOpenParticipantManager={() =>
                setNestedPageState({
                  isOpen: true,
                  active: "participant-manager",
                })
              }
            />
            {isPastBooking ? (
              <div
                className="border-input_border_color flex items-center
                  justify-between rounded-lg border px-4 py-3"
              >
                <div className="flex items-center gap-2">
                  <Icon icon={Calendar02Icon} styles="size-5" />
                  <span>{monthDateFormat(originalBookingData.start)}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Icon icon={Clock01Icon} styles="size-5" />
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

            {teamOptions.length > 1 && (
              <Select
                options={teamOptions}
                value={bookingData.employeeId}
                labelText="Employee"
                onSelect={(option) =>
                  updateBookingData({ employeeId: option.value })
                }
                disabled={isPastBooking || isBookingCompleted}
              />
            )}
            {isRecurring && (
              <div
                className="border-input_border_color flex items-center gap-2
                  rounded-lg border p-3"
              >
                <Icon icon={ArrowReloadHorizontalIcon} styles="size-5" />
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
              bottom-0 flex w-full border-t px-6 py-4 lg:w-110"
          >
            <Button
              styles="py-2 px-4 w-full"
              variant="primary"
              name="saveButton"
              buttonText="Save"
              onClick={() => {
                if (isRecurring) {
                  setIsRecurModalOpen(true);
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
  bookingData,
  isBookingCompleted,
  isPastBooking,
  isGroupBooking,
  isWindowSmall,
  serviceColor,
  updateBookingData,
  onCancel,
  onClose,
}) {
  return (
    <>
      {isWindowSmall && (
        <div
          className="flex w-full items-center justify-end px-4 pt-4"
          style={{ backgroundColor: serviceColor }}
        >
          <CloseButton onClick={onClose} styles="size-8 fill-white" />
        </div>
      )}
      <div
        className="sticky top-0 z-10 -mt-1 flex w-full items-center
          justify-between rounded-b-xl px-5 pt-4 pb-5"
        style={{ backgroundColor: serviceColor }}
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
              <Icon icon={Tick02Icon} styles="size-5" />
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
                <Icon
                  icon={ArrowLeft01Icon}
                  styles="-rotate-90 size-4 mt-0.5 text-white shadow-lg
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
                      <Icon icon={Tick02Icon} styles="size-6 fill-text_color" />
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
                        <Icon icon={UnavailableIcon} styles="size-5" />
                        No-show
                      </button>
                    </PopoverClose>
                  )}
                {bookingData.bookingStatus !== "confirmed" &&
                  !isPastBooking && (
                    <PopoverClose asChild>
                      <button
                        className="hover:bg-hvr_gray cursor-pointer gap-3"
                        onClick={() =>
                          updateBookingData({ bookingStatus: "confirmed" })
                        }
                      >
                        Confirmed
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
                      onClick={onCancel}
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
