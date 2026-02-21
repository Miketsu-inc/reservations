import {
  BackArrowIcon,
  ClockIcon,
  CustomersIcon,
  EnvelopeIcon,
  PersonIcon,
  PhoneIcon,
  PlusIcon,
  RefreshIcon,
  ThreeDotsIcon,
  TrashBinIcon,
} from "@reservations/assets";
import {
  Avatar,
  Card,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
} from "@reservations/components";
import {
  formatDuration,
  getDaySuffix,
  getDisplayPrice,
  timeStringFromDate,
} from "@reservations/lib";

export function ParticipantsCard({ participants, onClick, maxParticipants }) {
  const hasParticipants = participants.length > 0;
  const displayedParticipants = participants.slice(0, 3);
  const remainingParticipants = participants.length - 3;

  return (
    <button
      onClick={onClick}
      className="border-input_border_color hover:bg-hvr_gray/20 flex w-full
        items-center justify-between rounded-md border px-6 py-4 transition-all"
    >
      {hasParticipants ? (
        <div className="flex items-center gap-6">
          <div className="flex -space-x-6">
            {displayedParticipants.map((participant) => (
              <Avatar
                key={participant.id}
                img={participant.avatar_url}
                initials={participant.first_name[0] + participant.last_name[0]}
                styles="size-12! rounded-full! border-2 border-layer_bg"
              />
            ))}
            {remainingParticipants > 0 && (
              <div
                className="border-layer_bg flex size-12 items-center
                  justify-center rounded-full border-2 bg-gray-300 text-sm
                  font-semibold text-white dark:bg-gray-600"
              >
                +{remainingParticipants}
              </div>
            )}
          </div>

          <div className="flex flex-col items-start gap-0.5">
            <span className="text-text_color font-medium">
              {participants.length} / {maxParticipants} added
            </span>
          </div>
        </div>
      ) : (
        <span className="text-text_color text-left">
          Add customers to this group event
        </span>
      )}

      {hasParticipants ? (
        <BackArrowIcon
          styles="size-6 rotate-180 stroke-current text-gray-500
            dark:text-gray-400"
        />
      ) : (
        <div
          className="bg-primary/20 flex size-12 shrink-0 items-center
            justify-center rounded-full"
        >
          <PlusIcon styles="size-6 text-primary" />
        </div>
      )}
    </button>
  );
}

export function SelectedCustomerCard({ customer, onRemove, onView }) {
  return (
    <Card
      styles="flex justify-start border-input_border_color relative
        shadow-none!"
    >
      <div className="flex gap-4">
        <Avatar
          styles="size-14! text-[16px]! shrink-0 rounded-full!"
          img={customer?.avatar_url}
          initials={
            customer?.first_name && customer?.last_name
              ? `${customer.first_name[0]}${customer.last_name[0]}`
              : "?"
          }
        />
        <div className="flex flex-col gap-2">
          <span className="text-lg font-medium">{`${customer?.first_name} ${customer?.last_name}`}</span>
          <div className="flex w-full flex-col items-start gap-2">
            {customer.email && (
              <div
                className="text-text_color/70 flex items-center gap-2 text-sm"
              >
                <EnvelopeIcon styles="size-4 text-text_color/70" />

                <span className="truncate">{customer.email}</span>
              </div>
            )}
            {customer.phone_number && (
              <div
                className="text-text_color/70 flex items-center gap-2 text-sm"
              >
                <PhoneIcon
                  styles="size-3.5 fill-text_color/70 stroke-text_color/10"
                />

                <span>{customer.phone_number}</span>
              </div>
            )}
          </div>
        </div>
      </div>
      <Popover>
        <PopoverTrigger asChild>
          <button
            className="hover:bg-hvr_gray hover:*:stroke-text_color absolute
              top-4 right-3 h-fit cursor-pointer rounded-lg p-1"
          >
            <ThreeDotsIcon
              styles="size-6 stroke-4 stroke-gray-600 dark:stroke-gray-500
                rotate-90"
            />
          </button>
        </PopoverTrigger>
        <PopoverContent side="left" styles="w-auto">
          <div
            className="itmes-start flex flex-col *:flex *:w-full *:flex-row
              *:items-center *:rounded-lg *:p-2"
          >
            <PopoverClose asChild>
              <button
                className="hover:bg-hvr_gray cursor-pointer gap-3"
                onClick={onView}
              >
                <PersonIcon styles="size-5 fill-text_color" />
                <p>View Customer</p>
              </button>
            </PopoverClose>
            {customer.phone_number && (
              <PopoverClose asChild>
                <a
                  className="hover:bg-hvr_gray cursor-pointer gap-3"
                  href={`tel:${customer.phone_number}`}
                >
                  <PhoneIcon styles="size-4 ml-1 fill-text_color" />
                  <p>Call customer</p>
                </a>
              </PopoverClose>
            )}
            {customer.email && (
              <PopoverClose asChild>
                <a
                  className="hover:bg-hvr_gray cursor-pointer gap-3"
                  href={`mailto:${customer.email}`}
                >
                  <EnvelopeIcon styles="size-4 ml-1 text-text_color" />
                  <p>Email customer</p>
                </a>
              </PopoverClose>
            )}
            <PopoverClose asChild>
              <button
                onClick={onRemove}
                className="hover:bg-hvr_gray cursor-pointer gap-3"
              >
                <TrashBinIcon styles="size-5 mb-0.5" />
                <p className="text-red-600 dark:text-red-500">
                  Remove customer
                </p>
              </button>
            </PopoverClose>
          </div>
        </PopoverContent>
      </Popover>
    </Card>
  );
}

export function ServiceCard({ service, onClick, disabled, styles }) {
  const isGroup = service.booking_type !== "appointment";

  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={`hover:bg-hvr_gray/20 border-input_border_color relative flex
        w-full cursor-pointer items-center justify-between overflow-hidden
        rounded-md border border-l-0 px-7 py-4 ${styles}`}
    >
      <div
        className="absolute top-0 bottom-0 left-0 w-1.5 rounded-2xl opacity-50"
        style={{ backgroundColor: service.color }}
      />

      <div className="flex flex-col items-start gap-2">
        <span className="text-text_color text-left text-lg font-semibold">
          {service.name}
        </span>

        <div
          className="flex items-center gap-5 text-sm text-gray-500
            dark:text-gray-400"
        >
          <div className="flex items-center gap-1.5 font-medium">
            <ClockIcon styles="size-4 stroke-gray-500 dark:stroke-gray-400" />
            <span>{formatDuration(service.duration)}</span>
          </div>

          {isGroup && (
            <div
              className="flex items-center gap-1.5 font-medium text-gray-500
                dark:text-gray-400"
            >
              <CustomersIcon styles="size-5" />
              <span>Max {service.max_participants}</span>
            </div>
          )}
        </div>
      </div>

      <div className="flex items-center">
        <span className="text-text_color text-lg font-medium">
          {getDisplayPrice(service.price, service.price_type)}
        </span>
      </div>
    </button>
  );
}

export function AddCustomerCard({ onClick }) {
  return (
    <Card
      styles="w-full hover:bg-gray-200/30 dark:hover:bg-gray-700/10
        border-input_border_color! shadow-none!"
    >
      <button
        className="flex w-full items-center justify-between px-2"
        onClick={onClick}
      >
        <div className="flex flex-col items-start gap-1">
          <span className="text-text_color font-medium">Add Client</span>
          <span className="text-sm text-gray-500 dark:text-gray-400">
            Leave empty for walk-ins
          </span>
        </div>
        <div
          className="bg-primary/20 text-primary flex size-12 items-center
            justify-center rounded-full"
        >
          <PlusIcon styles="size-6" />
        </div>
      </button>
    </Card>
  );
}

export function RecurSummaryCard({ recurData, booking, onClick }) {
  function getFrequencyText() {
    switch (recurData.frequency) {
      case "daily":
        return "daily";
      case "weekly":
        return `every ${booking.start.toLocaleDateString("en-US", { weekday: "long" })}`;
      case "monthly":
        return `monthly on the ${getDaySuffix(booking.start.getDate())}`;
      case "custom":
        if (recurData.days.length > 0) {
          return `every ${recurData.interval} ${recurData.intervalUnit}`;
        }
        return "Custom";
    }
  }

  return (
    <button
      onClick={onClick}
      className="border-input_border_color flex w-full cursor-pointer
        items-center justify-between rounded-md border p-3.5 shadow-none!
        transition-colors hover:bg-gray-200/30 dark:hover:bg-gray-700/10"
    >
      <div className="flex flex-col items-start gap-1">
        <div className="flex items-center gap-2">
          <RefreshIcon styles="size-5 text-text_color/80 mt-0.5" />
          <span className="text-text_color">
            {recurData.isRecurring
              ? `Repeats ${getFrequencyText()}`
              : "Does not repeat"}
          </span>
        </div>

        {recurData.isRecurring && (
          <>
            <div className="text-left text-sm text-gray-500 dark:text-gray-400">
              {timeStringFromDate(booking.start)} -{" "}
              {timeStringFromDate(booking.end)}
              {" • "}
              Until{" "}
              {recurData.endDate.toLocaleDateString([], {
                month: "short",
                day: "numeric",
              })}
            </div>

            {recurData.frequency === "custom" && recurData.days.length > 0 && (
              <div className="mt-1 flex flex-wrap gap-1">
                {recurData.days.map((day) => (
                  <span
                    key={day}
                    className="bg-primary/10 text-primary rounded px-2 py-0.5
                      text-xs font-medium"
                  >
                    {day}
                  </span>
                ))}
              </div>
            )}
          </>
        )}
      </div>

      <BackArrowIcon
        styles="size-6 rotate-180 stroke-current text-gray-500
          dark:text-gray-400"
      />
    </button>
  );
}
