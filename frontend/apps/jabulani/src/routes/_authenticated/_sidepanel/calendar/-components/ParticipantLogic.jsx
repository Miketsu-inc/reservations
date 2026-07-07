import { PlusSignIcon, WalkingIcon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";
import {
  AddCustomerCard,
  ParticipantsCard,
  SelectedCustomerCard,
} from "./BookingCards";
import CustomerProfile from "./CustomerProfile";
import CustomerSelector from "./CustomerSelector";
import ParticipantManager from "./ParticipantManager";

export function ParticipantSideBar({
  isGroupBooking,
  hasSelection,
  isPastBooking = false,
  isExpanded,
  setIsExpanded,
  customers,
  selectedCustomers,
  maxParticipants,
  isWindowSmall,
  onAddCustomer,
  onRemoveCustomer,
  onRemoveParticipant,
  onStatusChange,
}) {
  function renderContent() {
    if (isGroupBooking) {
      return (
        <ParticipantManager
          customers={customers}
          participants={selectedCustomers}
          maxParticipants={maxParticipants}
          isWindowSmall={isWindowSmall}
          disabled={isPastBooking}
          onAdd={onAddCustomer}
          onRemove={onRemoveParticipant}
          onStatusChange={onStatusChange}
        />
      );
    }

    if (isPastBooking && !hasSelection) {
      return (
        <div
          className="flex flex-col items-center justify-center gap-3 px-3 py-10
            opacity-70"
        >
          <div
            className="flex size-16 items-center justify-center rounded-full
              bg-gray-200 dark:bg-gray-400/20"
          >
            <Icon icon={WalkingIcon} styles="fill-gray-300 size-6" />
          </div>
          <div className="text-center">
            <p className="font-semibold">Walk-in Customer</p>
            <button
              className="border-primary hover:bg-primary/10 mt-2 rounded-md p-2
                text-xs"
            >
              Assign Customer
            </button>
          </div>
        </div>
      );
    }

    if (!isPastBooking && isExpanded) {
      return (
        <CustomerSelector
          onSave={onAddCustomer}
          customers={customers}
          isGroupMode={false}
          walkIn={() => setIsExpanded(false)}
          selected={selectedCustomers}
          isWindowSmall={isWindowSmall}
        />
      );
    }

    if (hasSelection) {
      return (
        <CustomerProfile
          customer={selectedCustomers[0]}
          onRemove={onRemoveCustomer}
          disabled={isPastBooking}
        />
      );
    }

    return (
      <button
        className="flex h-full w-full cursor-pointer flex-col items-start px-3
          py-10 hover:bg-gray-200/40 dark:hover:bg-gray-700/10"
        onClick={() => setIsExpanded(true)}
      >
        <div className="flex flex-col items-center justify-center gap-3">
          <div
            className="bg-primary/20 text-primary flex size-14 items-center
              justify-center rounded-full"
          >
            <Icon icon={PlusSignIcon} styles="size-6" />
          </div>
          <div>
            <p className="font-semibold">Add customer</p>
            <span className="text-gray-400 dark:text-gray-500">
              Or leave empty for walk-ins
            </span>
          </div>
        </div>
      </button>
    );
  }

  if (!isWindowSmall) {
    return (
      <div
        className={`border-border_color overflow-hidden border-r transition-all
          duration-300 ease-in-out
          ${isExpanded || hasSelection || isGroupBooking ? "w-80" : "w-40"}`}
      >
        {renderContent()}
      </div>
    );
  }
}

export function MobileParticipantSection({
  isWindowSmall,
  isGroupBooking,
  hasSelection,
  isPastBooking = false,
  selectedCustomers,
  maxParticipants,
  onRemoveCustomer,
  onOpenCustomerSelector,
  onOpenProfile,
  onOpenParticipantManager,
}) {
  if (isWindowSmall) {
    return (
      <div className="flex flex-col gap-1">
        <span>{isGroupBooking ? "Participants" : "Participant"}</span>
        {isGroupBooking ? (
          isPastBooking && !hasSelection ? (
            <div
              className="border-input_border_color/70 flex items-center gap-3
                rounded-md border p-4"
            >
              <span className="text-gray-800 dark:text-gray-300">
                No participants attended
              </span>
            </div>
          ) : (
            <ParticipantsCard
              participants={selectedCustomers}
              onClick={onOpenParticipantManager}
              maxParticipants={maxParticipants}
            />
          )
        ) : (
          <div className="flex flex-col gap-2">
            {hasSelection ? (
              <SelectedCustomerCard
                customer={selectedCustomers[0]}
                onRemove={onRemoveCustomer}
                isGroupBooking={isGroupBooking}
                onView={onOpenProfile}
                disabled={isPastBooking}
              />
            ) : isPastBooking ? (
              <div
                className="border-input_border_color/70 flex items-center gap-6
                  rounded-md border p-4 px-6"
              >
                <div
                  className="flex size-12 items-center justify-center
                    rounded-full bg-gray-200 dark:bg-gray-400/20"
                >
                  <Icon icon={WalkingIcon} styles="fill-gray-300 size-6" />
                </div>
                <span className="text-gray-800 dark:text-gray-300">
                  Walk-in Customer
                </span>
              </div>
            ) : (
              <AddCustomerCard onClick={onOpenCustomerSelector} />
            )}
          </div>
        )}
      </div>
    );
  }
}
