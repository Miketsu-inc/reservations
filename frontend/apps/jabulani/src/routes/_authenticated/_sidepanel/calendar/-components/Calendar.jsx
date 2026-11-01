import { Calendar as FullCalendar } from "@fullcalendar/react";
import interactionPlugin from "@fullcalendar/react/interaction";
import listPlugin from "@fullcalendar/react/list";
import { ServerError } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  businessHoursQueryOptions,
  calendarTeamMembersQueryOptions,
  dateStringToLocalDate,
  formatToDateString,
  preferencesQueryOptions,
  useWindowSize,
} from "@reservations/lib";
import {
  keepPreviousData,
  useQuery,
  useSuspenseQueries,
} from "@tanstack/react-query";
import { useParams } from "@tanstack/react-router";
import {
  useCallback,
  useEffect,
  useEffectEvent,
  useMemo,
  useRef,
  useState,
} from "react";
import AgendaEventContent from "./AgendaEventContent";
import CalendarEventContent from "./CalendarEventContent";
import CalendarEventPreview from "./CalendarEventPreview";
import {
  applyDroppedEmployee,
  formatBlockedTimes,
  formatBookings,
  formatResources,
  sortMembersByName,
  transformBusinessHours,
} from "./calendarEvents";
import { DayHeader, ListDayHeader, ResourceHeader } from "./CalendarHeaders";
import CalendarMemberFilter, { UNASSIGNED_ID } from "./CalendarMemberFilter";
import {
  calendarBookingQueryOptions,
  calendarBookingsQueryOptions,
} from "./calendarQueries";
import { observeCurrentTime, scrollToDay } from "./calendarScroll";
import CalendarSidePanel from "./CalendarSidePanel";
import calendarTheme from "./calendarTheme";
import CalendarToolbar, { formatCalendarTitle } from "./CalendarToolbar";
import {
  canViewTeamCalendar,
  getCalendarRange,
  getCalendarRanges,
  getCalendarView,
  hasTeamColumns,
  resolveCalendarRange,
  timeGridViews,
} from "./calendarViews";
import CreateMenu, { CalendarCreateMenu } from "./CreateMenu";
import resourceTimeGridPlugin from "./resourceTimeGridPlugin";

import "@fullcalendar/react/skeleton.css";
import "@fullcalendar/react/themes/breezy/theme.css";

// Staff don't load the team, and a shared array keeps the useMemo's stable.
const NO_TEAM_MEMBERS = [];

const defaultSidePanelState = {
  isOpen: false,
  type: null,
  data: null,
  initialStart: null,
  initialEmployeeId: null,
  initialAllDay: false,
  bookingId: null,
  panelKey: null,
};

function createBookingPanelState(bookingId, event = null) {
  return {
    isOpen: true,
    type: "edit-booking",
    data: event,
    initialStart: null,
    initialEmployeeId: null,
    initialAllDay: false,
    bookingId,
    panelKey: `edit-booking:${bookingId}`,
  };
}

function getCalendarCreateStart(date) {
  const now = new Date();
  const calendarDate = new Date(date);

  calendarDate.setHours(now.getHours(), now.getMinutes(), 0, 0);
  return calendarDate < now ? now : calendarDate;
}

function getInitialCalendarDate(search) {
  return dateStringToLocalDate(search.start) ?? new Date();
}

export default function Calendar({ router, route, search }) {
  const { bookingId } = useParams({ strict: false });
  const [sidePanelState, setSidePanelState] = useState(() =>
    bookingId ? createBookingPanelState(bookingId) : defaultSidePanelState
  );
  const [visibleDates, setVisibleDates] = useState(null);
  const [createSelection, setCreateSelection] = useState(null);

  // Keep the old booking only when the route disappears, so its panel can
  // animate out before the transition callback clears it.
  if (bookingId && sidePanelState.bookingId !== bookingId) {
    setSidePanelState(createBookingPanelState(bookingId));
  }

  const { merchantId, employeeId, role } = useAuth();
  const canViewTeam = canViewTeamCalendar(role);
  const { queryClient } = route.useRouteContext({ from: route.id });
  const {
    data: events = { bookings: [], blocked_times: [] },
    isError,
    error,
  } = useQuery({
    ...calendarBookingsQueryOptions(merchantId, search.start, search.end),
    placeholderData: keepPreviousData,
  });
  const { data: selectedBooking } = useQuery({
    ...calendarBookingQueryOptions(merchantId, sidePanelState.bookingId),
    enabled: Boolean(sidePanelState.bookingId),
  });
  // The team loads with the rest, so the team columns don't pop in.
  const [{ data: preferences }, { data: businessHours }, teamQuery] =
    useSuspenseQueries({
      queries: [
        preferencesQueryOptions(merchantId, employeeId),
        businessHoursQueryOptions(merchantId),
        ...(canViewTeam
          ? [
              {
                ...calendarTeamMembersQueryOptions(merchantId),
                select: sortMembersByName,
              },
            ]
          : []),
      ],
    });
  const teamMembers = teamQuery?.data ?? NO_TEAM_MEMBERS;
  // null shows everyone, so new team members and unassigned bookings show up.
  const [memberFilter, setMemberFilter] = useState(null);
  // Kept in state rather than read from the URL, so the resources update in
  // the same render as the view instead of after the navigation.
  const [selectedRange, setSelectedRange] = useState(search.view);

  const { isWindowSmall } = useWindowSize();
  const calendarRanges = getCalendarRanges(role, isWindowSmall);
  const calendarRef = useRef();
  const calendarScrollRef = useRef();
  const dragRevertRef = useRef(null);
  const isDraggingRef = useRef(false);

  const memberIds = useMemo(
    () =>
      new Set(
        canViewTeam ? teamMembers.map((member) => member.id) : [employeeId]
      ),
    [canViewTeam, employeeId, teamMembers]
  );

  const allEvents = useMemo(
    () => [
      ...formatBookings(events.bookings, memberIds),
      ...formatBlockedTimes(events.blocked_times, memberIds),
    ],
    [memberIds, events.bookings, events.blocked_times]
  );

  const hasUnassigned =
    canViewTeam && allEvents.some((event) => event.extendedProps.is_unassigned);

  const filterOptionIds = useMemo(
    () => (hasUnassigned ? [...memberIds, UNASSIGNED_ID] : [...memberIds]),
    [hasUnassigned, memberIds]
  );

  // Drops members that have since left the team, and unassigned once the
  // shown dates have none.
  const filteredIds = useMemo(() => {
    if (!canViewTeam) return [employeeId];
    if (!memberFilter) return null;

    const ids = filterOptionIds.filter((id) => memberFilter.includes(id));
    return ids.length > 0 ? ids : null;
  }, [canViewTeam, employeeId, memberFilter, filterOptionIds]);

  const membersById = useMemo(
    () =>
      canViewTeam
        ? new Map(teamMembers.map((member) => [member.id, member]))
        : null,
    [canViewTeam, teamMembers]
  );

  const selectedIds = filteredIds ?? filterOptionIds;

  const calendarEvents = useMemo(() => {
    if (!filteredIds) return allEvents;

    const filter = new Set(filteredIds);
    return allEvents.filter((event) =>
      event.resourceIds.some((id) => filter.has(id))
    );
  }, [filteredIds, allEvents]);

  const calendarRange = resolveCalendarRange(
    calendarRanges,
    selectedRange,
    preferences.calendar_view
  );
  const resolvedView = getCalendarView(calendarRange);
  const showTeamColumns =
    hasTeamColumns(calendarRange) && selectedIds.length > 1;

  // Without resources the time grid merges everyone into one column per day.
  const resources = useMemo(() => {
    if (!showTeamColumns) return [];

    const selected = new Set(selectedIds);
    return formatResources(
      teamMembers.filter((member) => selected.has(member.id)),
      selected.has(UNASSIGNED_ID)
    );
  }, [showTeamColumns, teamMembers, selectedIds]);

  const selectedBookingEvent = useMemo(() => {
    if (!selectedBooking) return null;

    const event = formatBookings([selectedBooking], memberIds)[0];
    return {
      ...event,
      start: new Date(event.start),
      end: new Date(event.end),
    };
  }, [selectedBooking, memberIds]);

  const panelData = sidePanelState.bookingId
    ? (sidePanelState.data ?? selectedBookingEvent)
    : sidePanelState.data;

  const isPanelOpen = sidePanelState.bookingId
    ? Boolean(bookingId && panelData)
    : sidePanelState.isOpen;

  const panelType =
    sidePanelState.type === "edit-booking" && !panelData
      ? null
      : sidePanelState.type;

  const invalidateBookingsQuery = useCallback(async () => {
    await queryClient.invalidateQueries(
      calendarBookingsQueryOptions(merchantId, search.start, search.end)
    );
  }, [merchantId, queryClient, search]);

  const invalidateSelectedBookingQuery = useCallback(async () => {
    if (!sidePanelState.bookingId) return;

    await queryClient.invalidateQueries(
      calendarBookingQueryOptions(merchantId, sidePanelState.bookingId)
    );
  }, [sidePanelState.bookingId, merchantId, queryClient]);

  const closeCreateMenu = useCallback(() => {
    setCreateSelection(null);
  }, []);

  function openSidePanel(
    type,
    data,
    id,
    {
      revert = null,
      initialStart = null,
      initialEmployeeId = null,
      initialAllDay = false,
    } = {}
  ) {
    dragRevertRef.current = revert;
    setCreateSelection(null);

    setSidePanelState({
      isOpen: true,
      type,
      data,
      initialStart,
      initialEmployeeId,
      initialAllDay,
      bookingId: null,
      panelKey: `${type}:${id}`,
    });
  }

  function openBookingSidePanel(event, revert = null) {
    dragRevertRef.current = revert;
    setCreateSelection(null);

    setSidePanelState(createBookingPanelState(event.id, event));

    router.navigate({
      to: "/calendar/bookings/$bookingId",
      params: { bookingId: event.id },
      search,
    });
  }

  function openSidePanelAtTime(
    type,
    {
      start = getCalendarCreateStart(calendarRef.current.getApi().getDate()),
      employeeId = null,
      allDay = false,
    } = {}
  ) {
    openSidePanel(type, null, start.toISOString(), {
      initialStart: start,
      initialEmployeeId: employeeId,
      initialAllDay: allDay,
    });
  }

  function closeSidePanel() {
    dragRevertRef.current?.();
    dragRevertRef.current = null;

    if (sidePanelState.bookingId) {
      router.navigate({ to: "/calendar", search });
      return;
    }

    setSidePanelState((prev) => ({
      ...prev,
      isOpen: false,
    }));
  }

  function saveSidePanel() {
    dragRevertRef.current = null;
    invalidateBookingsQuery();

    if (sidePanelState.bookingId) {
      invalidateSelectedBookingQuery();
      router.navigate({ to: "/calendar", search });
      return;
    }

    setSidePanelState((prev) => ({
      ...prev,
      isOpen: false,
    }));
  }

  function finishSidePanelClose() {
    if (sidePanelState.bookingId) {
      dragRevertRef.current?.();
      dragRevertRef.current = null;
    }

    setSidePanelState(defaultSidePanelState);
  }

  function datesChangedHandler({ view }) {
    calendarScrollRef.current?.center();

    const lastDay = new Date(view.currentEnd);
    lastDay.setDate(lastDay.getDate() - 1);
    setVisibleDates({
      title: formatCalendarTitle(
        view.currentStart,
        lastDay,
        view.type === "listWeek"
      ),
      range: { from: view.currentStart, to: lastDay },
      buttons: view.calendar.getButtonState(),
    });

    const range = getCalendarRange(view.type);
    const start = formatToDateString(view.activeStart);
    const end = formatToDateString(view.activeEnd);
    if (range === search.view && start === search.start && end === search.end) {
      return;
    }

    router.navigate({
      to: "/calendar",
      search: (prev) => ({ ...prev, view: range, start, end }),
      replace: true,
    });
  }

  function todayButtonHandler() {
    const api = calendarRef.current.getApi();
    // Keeps the current range when it already shows today, instead of moving
    // the 2 days view to start today.
    if (api.getButtonState().today.isDisabled) return;

    api.today();
  }

  function changeViewHandler(view) {
    const api = calendarRef.current.getApi();

    if (view === api.view.type) return;

    // Otherwise switching from a week to a shorter range lands on its first day.
    const now = new Date();
    const showsToday = api.view.activeStart <= now && now < api.view.activeEnd;
    api.changeView(view, showsToday ? now : api.view.currentStart);
  }

  const syncCalendarView = useEffectEvent(() => {
    changeViewHandler(resolvedView);
  });

  useEffect(() => {
    syncCalendarView();
  }, [resolvedView]);

  function changeRangeHandler(range) {
    setSelectedRange(range);
    changeViewHandler(getCalendarView(range));
  }

  function changeMembersHandler(ids) {
    setMemberFilter(ids.length === filterOptionIds.length ? null : ids);
  }

  function changeDateHandler(date) {
    calendarRef.current.getApi().gotoDate(formatToDateString(date));
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col px-4 pt-4">
      <CalendarSidePanel
        isOpen={isPanelOpen}
        type={panelType}
        data={panelData}
        initialStart={sidePanelState.initialStart}
        initialEmployeeId={sidePanelState.initialEmployeeId}
        initialAllDay={sidePanelState.initialAllDay}
        panelKey={sidePanelState.panelKey}
        onClose={closeSidePanel}
        onTransitionEnd={finishSidePanelClose}
        onSave={saveSidePanel}
        onSoftUpdate={() => {
          invalidateBookingsQuery();
          invalidateSelectedBookingQuery();
        }}
        preferences={preferences}
      />
      {createSelection && (
        <CalendarCreateMenu
          selection={createSelection}
          onClose={closeCreateMenu}
          onCreateBlockedTime={() =>
            openSidePanelAtTime("blocked-time", createSelection)
          }
          onCreateBooking={() =>
            openSidePanelAtTime("new-booking", {
              ...createSelection,
              // A booking needs a time, so the all-day row starts it now.
              start: createSelection.allDay
                ? getCalendarCreateStart(createSelection.start)
                : createSelection.start,
            })
          }
        />
      )}
      <CalendarToolbar
        visibleDates={visibleDates}
        firstDayOfWeek={preferences.first_day_of_week}
        onSelectDate={changeDateHandler}
        onPrev={() => calendarRef.current.getApi().prev()}
        onToday={todayButtonHandler}
        onNext={() => calendarRef.current.getApi().next()}
        ranges={calendarRanges}
        range={calendarRange}
        onRangeChange={changeRangeHandler}
      >
        {canViewTeam && (
          <CalendarMemberFilter
            members={teamMembers}
            selectedIds={selectedIds}
            showUnassigned={hasUnassigned}
            onChange={changeMembersHandler}
            currentEmployeeId={employeeId}
          />
        )}
        <CreateMenu
          isFloating={isWindowSmall}
          onCreateBlockedTime={() => openSidePanelAtTime("blocked-time")}
          onCreateBooking={() => openSidePanelAtTime("new-booking")}
        />
      </CalendarToolbar>
      <div className="min-h-0 w-full flex-1 overflow-hidden">
        <FullCalendar
          ref={calendarRef}
          plugins={[
            calendarTheme,
            interactionPlugin,
            resourceTimeGridPlugin,
            listPlugin,
          ]}
          resources={resources}
          resourceOrder="isUnassigned,title"
          resourceDayHeaderContent={(info) => (
            <ResourceHeader resource={info.resource} />
          )}
          locale={navigator.language}
          editable={true}
          eventDurationEditable={true}
          selectable={true}
          initialView={resolvedView}
          initialDate={getInitialCalendarDate(search)}
          height="100%"
          className={
            resolvedView === "listWeek" ? "bg-transparent!" : undefined
          }
          headerToolbar={false}
          // The hints name the view's unit, which the 2 days view doesn't have.
          prevHint={(unit) => `Previous ${unit || "2 days"}`}
          nextHint={(unit) => `Next ${unit || "2 days"}`}
          viewDidMount={(info) => {
            calendarScrollRef.current = observeCurrentTime(info);
          }}
          viewWillUnmount={() => {
            calendarScrollRef.current?.disconnect();
            calendarScrollRef.current = null;
          }}
          events={calendarEvents}
          blockEventClass="calendar-event bg-(--fc-event-color)!"
          blockEventInnerClass="text-(--fc-event-contrast-color)!"
          eventContent={(eventInfo) => (
            <CalendarEventPreview
              event={eventInfo.event}
              timeFormat={preferences.time_format}
              canOpen={() =>
                !eventInfo.isMirror &&
                !eventInfo.isDragging &&
                !isDraggingRef.current
              }
            >
              <CalendarEventContent eventInfo={eventInfo} />
            </CalendarEventPreview>
          )}
          datesSet={datesChangedHandler}
          eventClick={(e) => {
            const type = e.event.extendedProps.type;

            if (type === "blocked") {
              openSidePanel("blocked-time", e.event, e.event.id);
              return;
            }

            openBookingSidePanel(e.event);
          }}
          dateClick={(info) => {
            // The all-day row stays clickable until the day is over.
            const end = new Date(info.date);
            if (info.allDay) end.setDate(end.getDate() + 1);

            if (
              !timeGridViews.has(info.view.type) ||
              end.getTime() < Date.now()
            ) {
              return;
            }

            setCreateSelection({
              start: info.date,
              allDay: info.allDay,
              employeeId:
                info.resource && info.resource.id !== UNASSIGNED_ID
                  ? Number(info.resource.id)
                  : null,
              x: info.jsEvent.clientX,
              y: info.jsEvent.clientY,
            });
          }}
          firstDay={preferences.first_day_of_week === "Monday" ? "1" : "0"}
          slotHeaderFormat={{
            hour: "numeric",
            minute: "numeric",
            hour12: preferences.time_format === "12-hour",
          }}
          slotDuration={preferences.time_frequency}
          nowIndicator={true}
          allDaySlot={true}
          displayEventEnd={false}
          snapDuration={{ minutes: 5 }}
          eventDragStart={() => {
            isDraggingRef.current = true;
          }}
          eventDragStop={() => {
            isDraggingRef.current = false;
          }}
          eventDrop={(e) => {
            applyDroppedEmployee(e);
            const type = e.event.extendedProps.type;

            if (type === "blocked") {
              openSidePanel("blocked-time", e.event, e.event.id, {
                revert: e.revert,
              });
              return;
            }

            openBookingSidePanel(e.event, e.revert);
          }}
          eventAllow={(dropInfo, draggedEvent) =>
            dropInfo.start.getTime() >= Date.now() &&
            // Bookings always need a team member, so nothing can be dropped on unassigned.
            (dropInfo.resource?.id !== UNASSIGNED_ID ||
              draggedEvent.extendedProps.is_unassigned)
          }
          dayHeaderFormat={{
            weekday: "short",
            day: "numeric",
            omitCommas: true,
          }}
          dayHeaderContent={(info) => <DayHeader info={info} />}
          eventTimeFormat={{
            hour: "2-digit",
            minute: "2-digit",
            hour12: preferences.time_format === "12-hour",
          }}
          views={{
            resourceTimeGrid2Day: {
              type: "resourceTimeGrid",
              duration: { days: 2 },
            },
            listWeek: {
              // The agenda already shows everything the preview would.
              eventContent: (eventInfo) => (
                <AgendaEventContent
                  eventInfo={eventInfo}
                  membersById={membersById}
                  timeFormat={preferences.time_format}
                />
              ),
              borderless: true,
              noEventsText: "Nothing scheduled this week",
              noEventsClass: "flex flex-1 items-center justify-center",
              noEventsInnerClass: "text-fg-muted text-sm",
              listDayFormat: {
                weekday: "long",
                month: "short",
                day: "numeric",
              },
              listDayHeaderContent: (info) => <ListDayHeader info={info} />,
              listDayHeaderDidMount: ({ el, isToday }) => {
                if (isToday) scrollToDay(el);
              },
              // Leaves room for the floating create button.
              listDaysClass: "flex flex-col gap-3 pb-20",
              listDayClass: "flex flex-col",
              listDayHeaderClass: "bg-bg z-10 pt-1 pb-2",
              // FullCalendar always renders a second date line, so hide it.
              listDayHeaderInnerClass: (info) =>
                info.level ? "hidden" : "flex items-baseline gap-2",
              listDayBodyClass: "flex flex-col gap-2",
              listItemEventClass: `relative flex cursor-pointer overflow-hidden
                rounded-xl border border-border bg-bg-raised py-3 pr-4 pl-5
                shadow-card hover:bg-hover`,
              listItemEventBeforeClass:
                "absolute inset-y-0 left-0 w-1.5 bg-(--fc-event-color)",
              listItemEventInnerClass: "min-w-0 flex-1",
            },
            resourceTimeGrid: {
              dayMinWidth: 120,
            },
          }}
          businessHours={transformBusinessHours(businessHours)}
        />
      </div>
    </div>
  );
}
