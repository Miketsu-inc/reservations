import Button from "@components/Button";
import ServerError from "@components/ServerError";
import dayGridPlugin from "@fullcalendar/daygrid";
import interactionPlugin from "@fullcalendar/interaction";
import listPlugin from "@fullcalendar/list";
import FullCalendar from "@fullcalendar/react";
import timeGridPlugin from "@fullcalendar/timegrid";
import BackArrowIcon from "@icons/BackArrowIcon";
import { formatToDateString, isoToDateString } from "@lib/datetime";
import { useCallback, useEffect, useRef, useState } from "react";
import CalendarModal from "./CalendarModal";

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

export default function Calendar({ router, view, start, eventData }) {
  const [calendarTitle, setCalendarTitle] = useState("");
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [eventInfo, setEventInfo] = useState(defaultEventInfo);
  const [serverError, setServerError] = useState();
  const calendarRef = useRef();

  const datesChanged = useCallback(
    (api) => {
      const start = formatToDateString(api.view.currentStart);
      const end = isoToDateString(api.view.currentEnd.toISOString());

      router.navigate({
        search: () => ({ view: api.view.type, start: start, end: end }),
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
  }

  useEffect(() => {
    const api = calendarRef.current.getApi();
    setCalendarTitle(api.view.title);
  }, [calendarRef, setCalendarTitle]);

  return (
    <div className="flex h-screen flex-col">
      <ServerError styles="mt-4 mb-2" error={serverError} />
      <div className="py-2 md:flex md:flex-row md:gap-2">
        <p className="whitespace-nowrap py-2 text-xl md:text-3xl">
          {calendarTitle}
        </p>
        <div className="flex w-full flex-row justify-between">
          <div className="flex flex-row">
            <button
              className="rounded-lg hover:bg-hvr_gray"
              type="button"
              onClick={() => navButtonHandler("prev")}
            >
              <BackArrowIcon styles="w-6 h-6 md:w-8 md:h-8 stroke-current" />
            </button>
            <button
              className="rounded-lg hover:bg-hvr_gray"
              type="button"
              onClick={() => navButtonHandler("next")}
            >
              <BackArrowIcon styles="w-6 h-6 md:w-8 md:h-8 stroke-current rotate-180" />
            </button>
            <Button
              variant="primary"
              styles="p-1 md:p-2"
              buttonText="today"
              onClick={todayButtonHandler}
            />
          </div>
          <div className="flex flex-row gap-1">
            <Button
              id="month"
              variant="primary"
              styles="p-2"
              buttonText="month"
              onClick={() => changeViewHandler("dayGridMonth")}
            />
            <Button
              variant="primary"
              styles="p-2"
              buttonText="week"
              onClick={() => changeViewHandler("timeGridWeek")}
            />
            <Button
              variant="primary"
              styles="p-2"
              buttonText="day"
              onClick={() => changeViewHandler("timeGridDay")}
            />
            <Button
              variant="primary"
              styles="p-2"
              buttonText="list"
              onClick={() => changeViewHandler("listWeek")}
            />
          </div>
        </div>
      </div>
      <div className="flex grow items-center justify-center">
        <div className="light h-full w-full bg-bg_color text-text_color">
          <FullCalendar
            ref={calendarRef}
            plugins={[
              dayGridPlugin,
              interactionPlugin,
              timeGridPlugin,
              listPlugin,
            ]}
            weekNumberCalculation="ISO"
            locale="hu"
            timeZone="UTC"
            editable={true}
            eventDurationEditable={true}
            selectable={true}
            initialView={view ? view : "timeGridWeek"}
            initialDate={start ? start : undefined}
            height="100%"
            headerToolbar={false}
            events={formatData(eventData)}
            eventClick={(e) => {
              setEventInfo(e.event);
              setIsModalOpen(true);
            }}
            lazyFetching={true}
            views={{
              dayGridMonth: {
                fixedWeekCount: false,
              },
              timeGridWeek: {
                titleFormat: {
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                },
                slotLabelFormat: {
                  hour: "numeric",
                  minute: "2-digit",
                },
                slotDuration: "00:15:00",
                slotMinTime: "08:00:00",
                slotMaxTime: "17:00:00",
                nowIndicator: true,
              },
              timeGridDay: {
                slotLabelFormat: {
                  hour: "numeric",
                  minute: "2-digit",
                },
                slotDuration: "00:15:00",
                slotMinTime: "08:00:00",
                slotMaxTime: "17:00:00",
                nowIndicator: true,
              },
            }}
            allDaySlot={false}
            eventTimeFormat={{
              hour: "numeric",
              minute: "2-digit",
              second: "2-digit",
              meridiem: false,
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
    </div>
  );
}
