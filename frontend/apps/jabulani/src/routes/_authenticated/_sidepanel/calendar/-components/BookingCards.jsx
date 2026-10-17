import {
  ArrowLeft01Icon,
  ArrowReloadHorizontalIcon,
  Call02Icon,
  Clock01Icon,
  Delete02Icon,
  Mail01Icon,
  Message01Icon,
  MoreVerticalIcon,
  PlusSignIcon,
  User03Icon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";
import {
  Avatar,
  Card,
  Icon,
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
      className="border-border-strong hover:bg-hover flex w-full items-center
        justify-between rounded-md border px-6 py-4 transition-all"
    >
      {hasParticipants ? (
        <div className="flex items-center gap-6">
          <div className="flex -space-x-6">
            {displayedParticipants.map((participant) => (
              <Avatar
                key={participant.customer_id}
                img={participant.avatar_url}
                initials={participant.first_name[0] + participant.last_name[0]}
                styles="size-12! border-2 border-bg-raised"
              />
            ))}
            {remainingParticipants > 0 && (
              <div
                className="border-bg-raised bg-bg-inactive flex size-12
                  items-center justify-center rounded-full border-2 text-sm
                  font-semibold text-white"
              >
                +{remainingParticipants}
              </div>
            )}
          </div>

          <div className="flex flex-col items-start gap-0.5">
            <span className="text-fg font-medium">
              {participants.length} / {maxParticipants} added
            </span>
          </div>
        </div>
      ) : (
        <span className="text-fg text-left">
          Add customers to this group event
        </span>
      )}

      {hasParticipants ? (
        <Icon
          icon={ArrowLeft01Icon}
          styles="size-6 rotate-180 stroke-current text-fg-muted"
        />
      ) : (
        <div
          className="bg-primary-subtle flex size-12 shrink-0 items-center
            justify-center rounded-full"
        >
          <Icon icon={PlusSignIcon} styles="size-6 text-fg-primary" />
        </div>
      )}
    </button>
  );
}

export function SelectedCustomerCard({ customer, onRemove, onView, disabled }) {
  return (
    <Card styles="flex justify-start border-border-strong relative shadow-none!">
      <div className="flex gap-4">
        <Avatar
          styles="size-14! text-[16px]! shrink-0"
          img={customer?.avatar_url}
          initials={
            customer?.first_name && customer?.last_name
              ? `${customer.first_name[0]}${customer.last_name[0]}`
              : "?"
          }
        />
        <div className="flex flex-col justify-center gap-2">
          <div className="flex items-center gap-4">
            <span className="text-lg font-medium">{`${customer?.first_name} ${customer?.last_name}`}</span>
          </div>
          <div className="flex w-full flex-col items-start gap-2">
            {customer.email && (
              <div className="text-fg-muted flex items-center gap-2 text-sm">
                <Icon icon={Mail01Icon} styles="size-4 text-fg-muted" />

                <span className="truncate">{customer.email}</span>
              </div>
            )}
            {customer.phone_number && (
              <div className="text-fg-muted flex items-center gap-2 text-sm">
                <Icon icon={Call02Icon} styles="size-3.5 text-fg-muted" />

                <span>{customer.phone_number}</span>
              </div>
            )}
          </div>
        </div>
      </div>
      {customer.customer_note && (
        <button
          className="ring-bg-raised bg-warning absolute top-4 left-14
            inline-flex items-center gap-2 rounded-full p-1 text-xs ring-2"
        >
          <Icon icon={Message01Icon} styles="size-3 fill-warning" />
        </button>
      )}
      <Popover>
        <PopoverTrigger asChild>
          <button
            className="hover:bg-hover hover:*:stroke-fg absolute top-4 right-3
              h-fit cursor-pointer rounded-lg p-1"
          >
            <Icon icon={MoreVerticalIcon} styles="size-6 text-fg-muted" />
          </button>
        </PopoverTrigger>
        <PopoverContent side="left" styles="w-auto">
          <div
            className="itmes-start flex flex-col *:flex *:w-full *:flex-row
              *:items-center *:rounded-lg *:p-2"
          >
            <PopoverClose asChild>
              <button
                className="hover:bg-hover cursor-pointer gap-3"
                onClick={onView}
              >
                <Icon icon={User03Icon} styles="size-5 ml-1 text-fg" />
                View Customer
              </button>
            </PopoverClose>
            {customer.customer_note && (
              <PopoverClose asChild>
                <button
                  className="hover:bg-hover text-fg-warning cursor-pointer
                    gap-3"
                  onClick={onView}
                >
                  <Icon
                    icon={Message01Icon}
                    styles="size-4 ml-1 mt-0.5 fill-warning"
                  />
                  View Note
                </button>
              </PopoverClose>
            )}
            {customer.phone_number && (
              <PopoverClose nativeButton={false} asChild>
                <a
                  className="hover:bg-hover cursor-pointer gap-3"
                  href={`tel:${customer.phone_number}`}
                >
                  <Icon icon={Call02Icon} styles="size-4 ml-1 text-fg" />
                  Call customer
                </a>
              </PopoverClose>
            )}
            {customer.email && (
              <PopoverClose nativeButton={false} asChild>
                <a
                  className="hover:bg-hover cursor-pointer gap-3"
                  href={`mailto:${customer.email}`}
                >
                  <Icon icon={Mail01Icon} styles="size-4 ml-1 text-fg" />
                  Email customer
                </a>
              </PopoverClose>
            )}
            {!disabled && (
              <PopoverClose asChild>
                <button
                  onClick={onRemove}
                  className="hover:bg-hover cursor-pointer gap-3"
                >
                  <Icon
                    icon={Delete02Icon}
                    styles="size-5 mb-0.5 text-fg-danger"
                  />
                  <p className="text-fg-danger">Remove customer</p>
                </button>
              </PopoverClose>
            )}
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
      className={`hover:bg-hover border-border-strong relative flex w-full
        cursor-pointer items-center justify-between overflow-hidden rounded-md
        border border-l-0 px-7 py-4 disabled:pointer-events-none
        disabled:hover:bg-transparent ${styles}`}
    >
      <div
        className="absolute top-0 bottom-0 left-0 w-1.5 rounded-2xl opacity-50"
        style={{ backgroundColor: service.color }}
      />

      <div className="flex flex-col items-start gap-2">
        <span className="text-fg text-left text-lg font-semibold">
          {service.name}
        </span>

        <div className="text-fg-muted flex items-center gap-5 text-sm">
          <div className="flex items-center gap-1.5 font-medium">
            <Icon icon={Clock01Icon} styles="size-4 text-fg-muted" />
            <span>{formatDuration(service.duration)}</span>
          </div>

          {isGroup && (
            <div className="text-fg-muted flex items-center gap-1.5 font-medium">
              <Icon icon={UserGroupIcon} styles="size-5" />
              <span>Max {service.max_participants}</span>
            </div>
          )}
        </div>
      </div>

      <div className="flex items-center">
        <span className="text-fg text-lg font-medium">
          {getDisplayPrice(service.price, service.price_type)}
        </span>
      </div>
    </button>
  );
}

export function AddCustomerCard({ onClick }) {
  return (
    <Card styles="w-full hover:bg-hover border-border-strong! shadow-none!">
      <button
        className="flex w-full items-center justify-between px-2"
        onClick={onClick}
      >
        <div className="flex flex-col items-start gap-1">
          <span className="text-fg font-medium">Add Client</span>
          <span className="text-fg-muted text-sm">
            Leave empty for walk-ins
          </span>
        </div>
        <div
          className="bg-primary-subtle text-fg-primary flex size-12 items-center
            justify-center rounded-full"
        >
          <Icon icon={PlusSignIcon} styles="size-6" />
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
      className="border-border-strong hover:bg-hover flex w-full cursor-pointer
        items-center justify-between rounded-md border p-3.5 shadow-none!
        transition-colors"
    >
      <div className="flex flex-col items-start gap-1">
        <div className="flex items-center gap-2">
          <Icon
            icon={ArrowReloadHorizontalIcon}
            styles="size-5 text-fg mt-0.5"
          />
          <span className="text-fg">
            {recurData.isRecurring
              ? `Repeats ${getFrequencyText()}`
              : "Does not repeat"}
          </span>
        </div>

        {recurData.isRecurring && (
          <>
            <div className="text-fg-muted text-left text-sm">
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
                    className="bg-primary-subtle text-fg-primary rounded px-2
                      py-0.5 text-xs font-medium"
                  >
                    {day}
                  </span>
                ))}
              </div>
            )}
          </>
        )}
      </div>

      <Icon
        icon={ArrowLeft01Icon}
        styles="size-6 rotate-180 stroke-current text-fg-muted"
      />
    </button>
  );
}
