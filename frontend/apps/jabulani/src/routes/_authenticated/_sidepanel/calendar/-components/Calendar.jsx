import { Calendar as FullCalendar } from "@fullcalendar/react";
import dayGridPlugin from "@fullcalendar/react/daygrid";
import interactionPlugin from "@fullcalendar/react/interaction";
import listPlugin from "@fullcalendar/react/list";
import themePlugin from "@fullcalendar/react/themes/breezy";
import timeGridPlugin from "@fullcalendar/react/timegrid";
import { ArrowLeft01Icon } from "@hugeicons/core-free-icons";
import {
  Button,
  DatePicker,
  Icon,
  Select,
  ServerError,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  businessHoursQueryOptions,
  DEFAULT_SERVICE_COLOR,
  formatToDateString,
  getMonthFromCalendarStart,
  preferencesQueryOptions,
  useWindowSize,
} from "@reservations/lib";
import {
  keepPreviousData,
  useQuery,
  useSuspenseQueries,
} from "@tanstack/react-query";
import { useParams } from "@tanstack/react-router";
import { useCallback, useMemo, useRef, useState } from "react";
import {
  calendarBookingQueryOptions,
  calendarBookingsQueryOptions,
} from "./calendarQueries";
import CalendarSidePanel from "./CalendarSidePanel";
import CreateMenu from "./CreateMenu";

import "@fullcalendar/react/skeleton.css";
import "@fullcalendar/react/themes/breezy/palettes/indigo.css";
import "@fullcalendar/react/themes/breezy/theme.css";

const calendarViewOptions = [
  { value: "dayGridMonth", label: "Month" },
  { value: "timeGridWeek", label: "Week" },
  { value: "timeGridDay", label: "Day" },
  { value: "listWeek", label: "List" },
];

function getContrastColor(color) {
  const hex = color.replace("#", "");

  const red = parseInt(hex.substring(0, 2), 16);
  const green = parseInt(hex.substring(2, 4), 16);
  const blue = parseInt(hex.substring(4, 6), 16);

  const brightness = red * 0.289 + green * 0.587 + blue * 0.114;
  return brightness > 186 ? "#000000" : "#ffffff";
}

function transformBusinessHours(businessHours) {
  if (!businessHours) return {};

  return Object.entries(businessHours).map(([day, times]) => ({
    daysOfWeek: [parseInt(day)],
    startTime: times.start_time.slice(0, 5),
    endTime: times.end_time.slice(0, 5),
  }));
}

function formatBookings(data) {
  if (data === undefined) return;

  return data.map((booking) => {
    let title = "Walk-in";
    const isGroup = booking.booking_type !== "appointment";

    if (isGroup) {
      title = `${booking.service_name} - ${booking.participants.length}/${booking.max_participants}`;
    } else if (booking.participants.length > 0) {
      const participant = booking.participants[0];
      title = `${participant.first_name} ${participant.last_name}`;
    }

    const serviceColor = booking.service_color ?? DEFAULT_SERVICE_COLOR;

    return {
      id: booking.id,
      title: title,
      start: booking.from_date,
      end: booking.to_date,
      color: serviceColor,
      textColor: getContrastColor(serviceColor),
      durationEditable: false,
      startEditable: new Date(booking.to_date) > new Date() ? true : false,
      extendedProps: {
        // this is a number unlike the normal 'id' which get's converted to a string
        id: booking.id,
        event_type: "booking",
        booking_type: booking.booking_type,
        booking_status: booking.booking_status,
        is_recurring: booking.is_recurring,
        max_participants: booking.max_participants,
        participants: booking.participants,
        merchant_note: booking.merchant_note,
        service_name: booking.service_name,
        service_id: booking.service_id,
        employee_id: booking.employee_id,
        duration: booking.duration,
        price: booking.price,
        price_type: booking.price_type,
      },
    };
  });
}

function formatBlockedTimes(data) {
  if (data === undefined) return;

  return data.map((blockedTime) => {
    const start = blockedTime.is_all_day
      ? blockedTime.blocked_day
      : blockedTime.from_date;
    const end = blockedTime.is_all_day ? undefined : blockedTime.to_date;
    const editableUntil = blockedTime.is_all_day
      ? new Date(`${blockedTime.blocked_day}T23:59:59`)
      : new Date(blockedTime.to_date);

    return {
      id: blockedTime.id,
      title: `${blockedTime.name} ${blockedTime?.icon || ""}`,
      start: start,
      end: end,
      color: "rgba(0, 0, 0, 0.6)",
      textColor: getContrastColor("#333333"),
      durationEditable: !blockedTime.is_all_day,
      allDay: blockedTime.is_all_day,
      startEditable: editableUntil > new Date(),
      extendedProps: {
        id: blockedTime.id,
        type: "blocked",
        name: blockedTime.name,
        blocked_type_id: blockedTime.blocked_type_id,
        employee_ids: blockedTime.employee_ids,
        allDay: blockedTime.is_all_day,
      },
    };
  });
}

const defaultSidePanelState = {
  isOpen: false,
  type: null,
  data: null,
  bookingId: null,
  panelKey: null,
};

function createBookingPanelState(bookingId, event = null) {
  return {
    isOpen: true,
    type: "edit-booking",
    data: event,
    bookingId,
    panelKey: `edit-booking:${bookingId}`,
  };
}

export default function Calendar({ router, route, search }) {
  const { bookingId } = useParams({ strict: false });
  const [sidePanelState, setSidePanelState] = useState(() =>
    bookingId ? createBookingPanelState(bookingId) : defaultSidePanelState
  );
  const [calendarTitle, setCalendarTitle] = useState("");

  // Keep the old booking only when the route disappears, so its panel can
  // animate out before the transition callback clears it.
  if (bookingId && sidePanelState.bookingId !== bookingId) {
    setSidePanelState(createBookingPanelState(bookingId));
  }

  const { merchantId, employeeId } = useAuth();
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
  const [{ data: preferences }, { data: businessHours }] = useSuspenseQueries({
    queries: [
      preferencesQueryOptions(merchantId, employeeId),
      businessHoursQueryOptions(merchantId),
    ],
  });
  const [calendarView, setCalendarView] = useState(search.view);

  const { isWindowSmall } = useWindowSize();
  const calendarRef = useRef();
  const dragRevertRef = useRef(null);

  const calendarEvents = useMemo(() => {
    return [
      ...formatBookings(events.bookings),
      ...formatBlockedTimes(events.blocked_times),
    ];
  }, [events.bookings, events.blocked_times]);

  const selectedBookingEvent = useMemo(() => {
    if (!selectedBooking) return null;

    const event = formatBookings([selectedBooking])[0];
    return {
      ...event,
      start: new Date(event.start),
      end: new Date(event.end),
    };
  }, [selectedBooking]);

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

  function openSidePanel(type, data, id, revert = null) {
    dragRevertRef.current = revert;

    setSidePanelState({
      isOpen: true,
      type,
      data,
      bookingId: null,
      panelKey: `${type}:${id}`,
    });
  }

  function openBookingSidePanel(event, revert = null) {
    dragRevertRef.current = revert;

    setSidePanelState(createBookingPanelState(event.id, event));

    router.navigate({
      to: "/calendar/bookings/$bookingId",
      params: { bookingId: event.id },
      search,
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

  const datesChanged = useCallback(
    (api) => {
      const start = formatToDateString(api.view.activeStart);
      const end = formatToDateString(api.view.activeEnd);

      router.navigate({
        to: "/calendar",
        search: () => ({ view: api.view.type, start: start, end: end }),
        replace: true,
      });
    },
    [router]
  );

  function navButtonHandler(dir) {
    const api = calendarRef.current.getApi();

    if (dir === "prev") {
      api.prev();
    } else if (dir === "next") {
      api.next();
    }

    datesChanged(api);
  }

  function todayButtonHandler() {
    const api = calendarRef.current.getApi();

    const today = new Date().getTime();
    if (
      api.view.currentStart.getTime() <= today &&
      api.view.currentEnd.getTime() >= today
    ) {
      return;
    }

    api.today();

    datesChanged(api);
  }

  function changeViewHandler(view) {
    const api = calendarRef.current.getApi();

    if (view === api.view.type) return;
    api.changeView(view);

    datesChanged(api);
    setCalendarView(view);
  }

  function changeDateHandler(date) {
    const api = calendarRef.current.getApi();
    const dateStr = formatToDateString(date);

    api.gotoDate(dateStr);
    datesChanged(api);
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  return (
    <div className="flex h-[85svh] flex-col px-4 pt-4 md:h-fit md:max-h-[90svh]">
      <CalendarSidePanel
        isOpen={isPanelOpen}
        type={panelType}
        data={panelData}
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
      <div className="relative flex flex-col pb-4 md:flex-row md:gap-2">
        <div
          className="flex w-full flex-col justify-between md:flex-row
            md:items-center"
        >
          <p className="text-2xl whitespace-nowrap md:text-3xl">
            {calendarTitle}
          </p>
          <div className="flex flex-row items-center justify-between gap-2">
            <div className="flex flex-row items-center gap-2">
              <CreateMenu
                isFloating={isWindowSmall}
                onCreateBlockedTime={() =>
                  openSidePanel("blocked-time", null, "new")
                }
                onCreateBooking={() =>
                  openSidePanel("new-booking", null, "new")
                }
              />
              <DatePicker
                styles="w-fit"
                hideText={true}
                firstDayOfWeek={preferences.first_day_of_week}
                clearAfterClose={true}
                onSelect={changeDateHandler}
              />
              <button
                className="hover:bg-hvr_gray cursor-pointer rounded-lg"
                type="button"
                onClick={() => navButtonHandler("prev")}
              >
                <Icon icon={ArrowLeft01Icon} styles="size-8" />
              </button>
              <Button
                variant="primary"
                styles="p-2 text-sm"
                buttonText="today"
                onClick={todayButtonHandler}
              />
              <button
                className="hover:bg-hvr_gray cursor-pointer rounded-lg"
                type="button"
                onClick={() => navButtonHandler("next")}
              >
                <Icon icon={ArrowLeft01Icon} styles="size-8 rotate-180" />
              </button>
            </div>
            <Select
              options={calendarViewOptions}
              value={calendarView}
              onSelect={(option) => changeViewHandler(option.value)}
              styles="w-28!"
            />
          </div>
        </div>
      </div>
      <div className="max-h-full w-full overflow-auto rounded-lg">
        <FullCalendar
          ref={calendarRef}
          plugins={[
            themePlugin,
            dayGridPlugin,
            interactionPlugin,
            timeGridPlugin,
            listPlugin,
          ]}
          locale="hu"
          editable={true}
          eventDurationEditable={true}
          selectable={true}
          initialView={search.view ? search.view : "timeGridWeek"}
          // dayGridMonth dates do not start or end with the current month's dates
          initialDate={
            search.view === "dayGridMonth"
              ? getMonthFromCalendarStart(search.start)
              : search.start
                ? search.start
                : undefined
          }
          height="auto"
          headerToolbar={false}
          events={calendarEvents}
          datesSet={({ view }) => setCalendarTitle(view.title)}
          eventClick={(e) => {
            const type = e.event.extendedProps.type;

            if (type === "blocked") {
              openSidePanel("blocked-time", e.event, e.event.id);
              return;
            }

            openBookingSidePanel(e.event);
          }}
          firstDay={preferences.first_day_of_week === "Monday" ? "1" : "0"}
          lazyFetching={true}
          slotHeaderFormat={{
            hour: "numeric",
            minute: "numeric",
            hour12: preferences.time_format === "12-hour",
          }}
          slotDuration={preferences.time_frequency}
          slotMinTime={preferences.start_hour}
          slotMaxTime={preferences.end_hour}
          nowIndicator={true}
          titleFormat={{
            year: "numeric",
            month: "long",
            day: "numeric",
          }}
          fixedWeekCount={false}
          allDaySlot={true}
          displayEventEnd={false}
          snapDuration={{ minutes: 5 }}
          eventDrop={(e) => {
            const type = e.event.extendedProps.type;

            if (type === "blocked") {
              openSidePanel("blocked-time", e.event, e.event.id, e.revert);
              return;
            }

            openBookingSidePanel(e.event, e.revert);
          }}
          eventAllow={(dropInfo) => {
            if (dropInfo.start.getTime() < Date.now()) {
              return false;
            } else {
              return true;
            }
          }}
          dayHeaderFormat={{
            weekday: "short",
            day: isWindowSmall ? undefined : "numeric",
            omitCommas: true,
          }}
          eventTimeFormat={{
            hour: "2-digit",
            minute: "2-digit",
            hour12: preferences.time_format === "12-hour",
          }}
          views={{
            dayGridMonth: {
              titleFormat: {
                year: "numeric",
                month: "long",
              },
              displayEventEnd: false,
            },
            timeGridDay: {
              dayHeaderFormat: {
                weekday: "long",
              },
              displayEventTime: false,
            },
            listWeek: {
              displayEventEnd: true,
            },
          }}
          businessHours={transformBusinessHours(businessHours)}
        />
      </div>
    </div>
  );
}
