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
      className="bg-bg-input border-border-strong hover:border-border-hover flex
        w-full items-center justify-between rounded-lg border px-3 py-2.5
        text-sm"
    >
      {hasParticipants ? (
        <div className="flex items-center gap-3">
          <div className="flex -space-x-4">
            {displayedParticipants.map((participant) => (
              <Avatar
                key={participant.customer_id}
                img={participant.avatar_url}
                initials={participant.first_name[0] + participant.last_name[0]}
                styles="size-8! text-[11px]! border-2 border-bg"
              />
            ))}
            {remainingParticipants > 0 && (
              <div
                className="border-bg bg-bg-inactive text-fg-muted flex size-8
                  items-center justify-center rounded-full border-2 text-[11px]
                  font-semibold"
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
        <span className="text-fg-muted text-left">Add participants</span>
      )}

      {hasParticipants ? (
        <Icon
          icon={ArrowLeft01Icon}
          styles="size-4.5 rotate-180 text-fg-muted"
        />
      ) : (
        <div
          className="bg-primary-subtle flex size-8 shrink-0 items-center
            justify-center rounded-full"
        >
          <Icon icon={PlusSignIcon} styles="size-4.5 text-fg-primary" />
        </div>
      )}
    </button>
  );
}

export function SelectedCustomerCard({ customer, onRemove, onView, disabled }) {
  return (
    <Card styles="flex justify-start relative p-3!">
      <div className="flex gap-3">
        <Avatar
          styles="size-10! text-xs! shrink-0"
          img={customer?.avatar_url}
          initials={
            customer?.first_name && customer?.last_name
              ? `${customer.first_name[0]}${customer.last_name[0]}`
              : "?"
          }
        />
        <div className="flex min-w-0 flex-col justify-center gap-1">
          <div className="flex items-center gap-4">
            <span className="text-sm font-medium">{`${customer?.first_name} ${customer?.last_name}`}</span>
          </div>
          <div className="flex w-full flex-col items-start gap-1">
            {customer.email && (
              <div className="text-fg-muted flex items-center gap-2 text-xs">
                <Icon icon={Mail01Icon} styles="size-3.5 text-fg-muted" />

                <span className="truncate">{customer.email}</span>
              </div>
            )}
            {customer.phone_number && (
              <div className="text-fg-muted flex items-center gap-2 text-xs">
                <Icon icon={Call02Icon} styles="size-3.5 text-fg-muted" />

                <span>{customer.phone_number}</span>
              </div>
            )}
          </div>
        </div>
      </div>
      {customer.customer_note && (
        <button
          className="ring-bg-raised bg-warning absolute top-9.5 left-9.5
            inline-flex items-center gap-2 rounded-full p-1 text-xs ring-2"
        >
          <Icon icon={Message01Icon} styles="size-3" />
        </button>
      )}
      <Popover>
        <PopoverTrigger asChild>
          <button
            className="text-fg-muted hover:bg-hover hover:text-fg absolute
              top-2.5 right-2 h-fit cursor-pointer rounded-lg p-1.5"
          >
            <Icon icon={MoreVerticalIcon} styles="size-4.5" />
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
                <Icon icon={User03Icon} styles="size-4.5 text-fg" />
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
                  <Icon icon={Message01Icon} styles="size-4.5" />
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
                  <Icon icon={Call02Icon} styles="size-4.5 text-fg" />
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
                  <Icon icon={Mail01Icon} styles="size-4.5 text-fg" />
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
                  <Icon icon={Delete02Icon} styles="size-4.5 text-fg-danger" />
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
      className={`bg-bg-raised border-border shadow-card
        hover:border-border-strong relative flex w-full cursor-pointer
        items-center justify-between gap-4 overflow-hidden rounded-xl border
        py-3 pr-4 pl-5 disabled:pointer-events-none ${styles}`}
    >
      <div
        className="absolute top-0 bottom-0 left-0 w-1.5"
        style={{ backgroundColor: service.color }}
      />

      <div className="flex min-w-0 flex-col items-start gap-1">
        <span className="text-fg truncate text-left font-semibold">
          {service.name}
        </span>

        <div className="text-fg-muted flex items-center gap-4 text-sm">
          <div className="flex items-center gap-1.5">
            <Icon icon={Clock01Icon} styles="size-3.5" />
            <span>{formatDuration(service.duration)}</span>
          </div>

          {isGroup && (
            <div className="flex items-center gap-1.5">
              <Icon icon={UserGroupIcon} styles="size-3.5" />
              <span>Max {service.max_participants}</span>
            </div>
          )}
        </div>
      </div>

      <div className="flex items-center">
        <span className="text-fg shrink-0 text-sm font-medium">
          {getDisplayPrice(service.price, service.price_type)}
        </span>
      </div>
    </button>
  );
}

export function AddCustomerCard({ onClick }) {
  return (
    <button
      className="bg-bg-input border-border-strong hover:border-border-hover flex
        w-full items-center justify-between rounded-lg border px-3 py-2.5
        text-left text-sm"
      onClick={onClick}
    >
      <div className="flex flex-col items-start gap-0.5">
        <span className="text-fg font-medium">Add customer</span>
        <span className="text-fg-muted text-xs">Leave empty for walk-ins</span>
      </div>
      <div
        className="bg-primary-subtle text-fg-primary flex size-10 shrink-0
          items-center justify-center rounded-full"
      >
        <Icon icon={PlusSignIcon} styles="size-5" />
      </div>
    </button>
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
      className="bg-bg-input border-border-strong hover:border-border-hover flex
        w-full cursor-pointer items-center justify-between rounded-lg border
        px-3 py-2 text-left text-sm"
    >
      <div className="flex flex-col items-start gap-1">
        <div className="flex items-center gap-2">
          <Icon
            icon={ArrowReloadHorizontalIcon}
            styles="size-4.5 text-fg-muted"
          />
          <span className="text-fg">
            {recurData.isRecurring
              ? `Repeats ${getFrequencyText()}`
              : "Does not repeat"}
          </span>
        </div>

        {recurData.isRecurring && (
          <>
            <div className="text-fg-muted text-left text-xs">
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

      <Icon icon={ArrowLeft01Icon} styles="size-5 rotate-180 text-fg-muted" />
    </button>
  );
}
