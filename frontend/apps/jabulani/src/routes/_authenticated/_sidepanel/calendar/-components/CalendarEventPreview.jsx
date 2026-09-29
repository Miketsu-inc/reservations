import {
  ArrowReloadHorizontalIcon,
  Calendar02Icon,
  Clock01Icon,
  User03Icon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";
import {
  Avatar,
  Icon,
  PreviewCard,
  PreviewCardContent,
  PreviewCardTrigger,
} from "@reservations/components";
import {
  formatDuration,
  getDisplayPrice,
  timeStringFromDate,
} from "@reservations/lib";

function capitalize(value) {
  if (!value) return "";

  return value.charAt(0).toUpperCase() + value.slice(1).replaceAll("-", " ");
}

function formatDate(date) {
  return date.toLocaleDateString([], {
    weekday: "long",
    month: "long",
    day: "numeric",
  });
}

function getParticipantName(participant) {
  return [participant.first_name, participant.last_name]
    .filter(Boolean)
    .join(" ");
}

function getParticipantInitials(participant) {
  const firstInitial = participant.first_name?.[0] ?? "";
  const lastInitial = participant.last_name?.[0] ?? "";

  return `${firstInitial}${lastInitial}` || "?";
}

function EventLabel({ eventInfo }) {
  const isListView = eventInfo.view.type.startsWith("list");
  const isMonthView = eventInfo.view.type === "dayGridMonth";

  if (isListView) {
    return (
      <div className="flex size-full min-w-0 items-center gap-4">
        {eventInfo.timeText && (
          <span
            className="w-28 shrink-0 text-sm text-gray-500 dark:text-gray-400"
          >
            {eventInfo.timeText}
          </span>
        )}
        <span className="min-w-0 truncate font-medium">
          {eventInfo.event.title}
        </span>
      </div>
    );
  }

  return (
    <div
      className={`flex size-full min-w-0 overflow-hidden
        ${isMonthView ? "flex-row items-center gap-1" : "flex-col"}`}
    >
      {eventInfo.timeText && (
        <span className="shrink-0 text-xs leading-tight font-medium">
          {eventInfo.timeText}
        </span>
      )}
      <span className="min-w-0 truncate text-xs leading-tight font-semibold">
        {eventInfo.event.title}
      </span>
    </div>
  );
}

function ParticipantAvatars({ participants }) {
  if (participants.length === 0) {
    return (
      <div
        className="bg-hvr_gray flex size-10 shrink-0 items-center justify-center
          rounded-full"
      >
        <Icon icon={User03Icon} styles="size-5 text-text_color/60" />
      </div>
    );
  }

  const visibleParticipants = participants.slice(0, 3);
  const remainingParticipants =
    participants.length - visibleParticipants.length;

  return (
    <div className="flex shrink-0 -space-x-2">
      {visibleParticipants.map((participant) => (
        <Avatar
          key={participant.id}
          img={participant.avatar_url}
          initials={getParticipantInitials(participant)}
          alt={getParticipantName(participant)}
          styles="size-10! rounded-full! border-2 border-layer_bg text-xs!"
        />
      ))}
      {remainingParticipants > 0 && (
        <div
          className="border-layer_bg bg-hvr_gray flex size-10 items-center
            justify-center rounded-full border-2 text-xs font-semibold"
        >
          +{remainingParticipants}
        </div>
      )}
    </div>
  );
}

function DetailRow({ icon, children }) {
  return (
    <div className="text-text_color/70 flex items-center gap-2.5 text-sm">
      <Icon icon={icon} styles="size-4 shrink-0" />
      <span>{children}</span>
    </div>
  );
}

function BookingPreview({ event, timeFormat }) {
  const { extendedProps } = event;
  const participants = extendedProps.participants ?? [];
  const isGroupBooking = extendedProps.booking_type !== "appointment";
  const customerName =
    participants.length > 0
      ? getParticipantName(participants[0]) || "Customer"
      : "Walk-in";
  const additionalParticipants = participants.length - 1;
  const participantSummary = isGroupBooking
    ? `${participants.length} of ${extendedProps.max_participants} participants`
    : capitalize(extendedProps.booking_status);

  return (
    <div className="w-80 max-w-[calc(100vw-2rem)] overflow-hidden">
      <div className="flex items-center gap-3 p-4">
        <ParticipantAvatars participants={participants} />
        <div className="min-w-0 flex-1">
          <p className="truncate font-semibold">
            {customerName}
            {isGroupBooking && additionalParticipants > 0
              ? ` +${additionalParticipants}`
              : ""}
          </p>
          <p className="text-text_color/60 mt-0.5 text-xs">
            {participantSummary}
          </p>
        </div>
        <span
          className="bg-hvr_gray text-text_color/70 rounded-full px-2 py-1
            text-[11px] font-medium"
        >
          {isGroupBooking ? "Group" : "Appointment"}
        </span>
      </div>

      <div className="border-border_color border-t px-4 py-3">
        <div className="flex items-center justify-between gap-4">
          <div className="flex min-w-0 items-center gap-2">
            <span
              className="size-2.5 shrink-0 rounded-full"
              style={{ backgroundColor: extendedProps.service_color }}
            />
            <span className="truncate text-sm font-semibold">
              {extendedProps.service_name}
            </span>
          </div>
          <span className="shrink-0 text-sm font-semibold">
            {getDisplayPrice(extendedProps.price, extendedProps.price_type)}
          </span>
        </div>
      </div>

      <div className="border-border_color flex flex-col gap-2.5 border-t p-4">
        <DetailRow icon={Calendar02Icon}>{formatDate(event.start)}</DetailRow>
        <DetailRow icon={Clock01Icon}>
          {timeStringFromDate(event.start, timeFormat)} –{" "}
          {timeStringFromDate(event.end, timeFormat)}
          <span className="text-text_color/45">
            {" "}
            · {formatDuration(extendedProps.duration)}
          </span>
        </DetailRow>
        {extendedProps.is_recurring && (
          <DetailRow icon={ArrowReloadHorizontalIcon}>
            Part of a repeating series
          </DetailRow>
        )}
      </div>
    </div>
  );
}

function BlockedTimePreview({ event, timeFormat }) {
  const { extendedProps } = event;
  const employeeCount = extendedProps.employee_ids?.length ?? 0;

  return (
    <div className="w-72 max-w-[calc(100vw-2rem)] overflow-hidden">
      <div className="flex items-center gap-3 p-4">
        <div
          className="bg-hvr_gray flex size-10 shrink-0 items-center
            justify-center rounded-full text-lg"
        >
          {extendedProps.icon || "—"}
        </div>
        <div className="min-w-0">
          <p className="truncate font-semibold">{extendedProps.name}</p>
          <p className="text-text_color/60 mt-0.5 text-xs">Blocked time</p>
        </div>
      </div>

      <div className="border-border_color flex flex-col gap-2.5 border-t p-4">
        <DetailRow icon={Calendar02Icon}>{formatDate(event.start)}</DetailRow>
        <DetailRow icon={Clock01Icon}>
          {extendedProps.allDay
            ? "All day"
            : `${timeStringFromDate(event.start, timeFormat)} – ${timeStringFromDate(event.end, timeFormat)}`}
        </DetailRow>
        {employeeCount > 0 && (
          <DetailRow icon={UserGroupIcon}>
            {employeeCount} team {employeeCount === 1 ? "member" : "members"}
          </DetailRow>
        )}
      </div>
    </div>
  );
}

export default function CalendarEventPreview({ eventInfo, timeFormat }) {
  const isBlockedTime = eventInfo.event.extendedProps.type === "blocked";

  return (
    <PreviewCard>
      <PreviewCardTrigger asChild delay={350} closeDelay={120}>
        <div className="size-full min-w-0 cursor-pointer">
          <EventLabel eventInfo={eventInfo} />
        </div>
      </PreviewCardTrigger>
      <PreviewCardContent
        align="start"
        side="right"
        styles="overflow-hidden"
        collisionPadding={12}
      >
        {isBlockedTime ? (
          <BlockedTimePreview event={eventInfo.event} timeFormat={timeFormat} />
        ) : (
          <BookingPreview event={eventInfo.event} timeFormat={timeFormat} />
        )}
      </PreviewCardContent>
    </PreviewCard>
  );
}
