import Button from "@components/Button";
import Select from "@components/Select";
import ServerError from "@components/ServerError";
import dayGridPlugin from "@fullcalendar/daygrid";
import interactionPlugin from "@fullcalendar/interaction";
import listPlugin from "@fullcalendar/list";
import FullCalendar from "@fullcalendar/react";
import timeGridPlugin from "@fullcalendar/timegrid";
import BackArrowIcon from "@icons/BackArrowIcon";
import { formatToDateString, getMonthFromCalendarStart } from "@lib/datetime";
import { useWindowSize } from "@lib/hooks";
import { useCallback, useEffect, useRef, useState } from "react";
import CalendarModal from "./CalendarModal";

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

function formatData(data) {
  if (data === undefined) return;

  return data.map((event) => ({
    id: event.id,
    title: event.service_name,
    start: event.from_date,
    end: event.to_date,
    color: event.service_color,
    textColor: getContrastColor(event.service_color),
    durationEditable: false,
    startEditable: new Date(event.to_date) > new Date() ? true : false,
    extendedProps: {
      appointment_id: event.id,
      first_name: event.first_name,
      last_name: event.last_name,
      phone_number: event.phone_number,
      user_comment: event.user_comment,
      merchant_comment: event.merchant_comment,
      price: event.price,
    },
  }));
}

const defaultEventInfo = {
  id: 0,
  title: "",
  start: new Date(),
  end: new Date(),
  startEditable: true,
  extendedProps: {
    appointment_id: 0,
    first_name: "",
    last_name: "",
    phone_number: "",
    user_comment: "",
    merchant_comment: "",
    price: 0,
  },
};

export default function Calendar({
  router,
  view,
  start,
  eventData,
  preferences,
}) {
  const [calendarTitle, setCalendarTitle] = useState("");
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [eventInfo, setEventInfo] = useState(defaultEventInfo);
  const [serverError, setServerError] = useState();
  const calendarRef = useRef();
  const [calendarView, setCalendarView] = useState(view);
  const windowSize = useWindowSize();

  const isWindowSmall = windowSize === "sm" || windowSize === "md";

  const datesChanged = useCallback(
    (api) => {
      const start = formatToDateString(api.view.activeStart);
      const end = formatToDateString(api.view.activeEnd);

      router.navigate({
        search: () => ({ view: api.view.type, start: start, end: end }),
        replace: true,
      });
      setCalendarTitle(api.view.title);
    },
    [router, setCalendarTitle]
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

  function changeDateHandler(e) {
    const date = e.target.value;
    if (!date) return;

    const api = calendarRef.current.getApi();

    api.gotoDate(date);
    datesChanged(api);
  }

  useEffect(() => {
    const api = calendarRef.current.getApi();
    setCalendarTitle(api.view.title);
  }, []);

  return (
    <div className="flex h-[85svh] flex-col md:h-fit md:max-h-[90svh]">
      <ServerError styles="mt-4 mb-2" error={serverError} />
      <div className="flex flex-col pb-2 md:flex-row md:gap-2">
        <div className="flex w-full flex-col justify-between md:flex-row md:items-center">
          <p className="py-2 text-2xl whitespace-nowrap md:text-3xl">
            {calendarTitle}
          </p>
          <div className="flex flex-row items-center justify-between gap-2">
            <div className="flex flex-row items-center gap-2">
              <input
                className="w-5 dark:[color-scheme:dark]"
                type="date"
                onChange={changeDateHandler}
              />
              <button
                className="hover:bg-hvr_gray cursor-pointer rounded-lg"
                type="button"
                onClick={() => navButtonHandler("prev")}
              >
                <BackArrowIcon styles="w-8 h-8 stroke-current" />
              </button>
              <Button
                variant="primary"
                styles="p-2"
                buttonText="today"
                onClick={todayButtonHandler}
              />
              <button
                className="hover:bg-hvr_gray cursor-pointer rounded-lg"
                type="button"
                onClick={() => navButtonHandler("next")}
              >
                <BackArrowIcon styles="w-8 h-8 stroke-current rotate-180" />
              </button>
            </div>
            <Select
              options={calendarViewOptions}
              value={calendarView}
              onSelect={(option) => changeViewHandler(option.value)}
              styles="w-28"
            />
          </div>
        </div>
      </div>
      <div className="light bg-bg_color text-text_color max-h-full w-full overflow-auto">
        <FullCalendar
          ref={calendarRef}
          plugins={[
            dayGridPlugin,
            interactionPlugin,
            timeGridPlugin,
            listPlugin,
          ]}
          locale="hu"
          timeZone="UTC"
          editable={true}
          eventDurationEditable={true}
          selectable={true}
          initialView={view ? view : "timeGridWeek"}
          // dayGridMonth dates do not start or end with the current month's dates
          initialDate={
            view === "dayGridMonth"
              ? getMonthFromCalendarStart(start)
              : start
                ? start
                : undefined
          }
          height="auto"
          headerToolbar={false}
          events={formatData(eventData)}
          eventClick={(e) => {
            setEventInfo(e.event);
            setIsModalOpen(true);
          }}
          firstDay={preferences.first_day_of_week === "Monday" ? "1" : "0"}
          lazyFetching={true}
          slotLabelFormat={{
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
          titleRangeSeparator=" - "
          fixedWeekCount={false}
          allDaySlot={false}
          displayEventEnd={false}
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
              displayEventTime: false,
            },
            timeGridWeek: {
              displayEventTime: isWindowSmall ? false : undefined,
            },
            timeGridDay: {
              dayHeaderFormat: {
                weekday: "long",
              },
            },
            listWeek: {
              displayEventEnd: true,
            },
          }}
        />
      </div>
      <CalendarModal
        eventInfo={eventInfo}
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false);
        }}
        setError={setServerError}
      />
    </div>
  );
}
