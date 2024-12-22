import dayGridPlugin from "@fullcalendar/daygrid";
import interactionPlugin from "@fullcalendar/interaction";
import listPlugin from "@fullcalendar/list";
import FullCalendar from "@fullcalendar/react";
import timeGridPlugin from "@fullcalendar/timegrid";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { useCallback, useRef, useState } from "react";
import ServerError from "../../../../../components/ServerError";
import { useClickOutside } from "../../../../../lib/hooks";
import CalendarModal from "./CalendarModal";

const defaultEventInfo = {
  extendedProps: {
    first_name: "",
    last_name: "",
    phone_number: "",
    user_comment: "",
    merchant_comment: "",
    price: "",
  },
};

export default function Calendar() {
  const modalRef = useRef();
  const [isLoading, setIsLoading] = useState(false);
  const [serverError, setServerError] = useState("");
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [eventInfo, setEventInfo] = useState(defaultEventInfo);
  useClickOutside(modalRef, () => setIsModalOpen(false));

  const formatData = (data) => {
    return data.map((event) => ({
      id: event.id,
      title: event.service_name,
      start: event.from_date,
      end: event.to_date,
      extendedProps: {
        first_name: event.first_name,
        last_name: event.last_name,
        phone_number: event.phone_number,
        user_comment: event.user_comment,
        merchant_comment: event.merchant_comment,
        price: event.price,
      },
    }));
  };

  const fetchEvents = useCallback(
    async (fetchInfo, successCallback, failureCallback) => {
      setIsLoading(true);
      try {
        const response = await fetch(
          `/api/v1/appointments/calendar?start=${fetchInfo.startStr}&end=${fetchInfo.endStr}`,
          {
            method: "GET",
          }
        );

        const result = await response.json();

        if (!response.ok) {
          invalidateLocalSotrageAuth(response.status);
          setServerError(result.error.message);
          failureCallback(result.error.message);
        } else {
          setServerError(undefined);

          if (result.data !== null) {
            const events = formatData(result.data);
            successCallback(events);
          } else {
            failureCallback("No appointments found");
          }
        }
      } catch (err) {
        setServerError(err.message);
        failureCallback(err.message);
      } finally {
        setIsLoading(false);
      }
    },
    []
  );

  function handleClick(e) {
    setEventInfo(e.event);
    setIsModalOpen(true);
  }

  return (
    <>
      <ServerError styles="mt-4 mb-2" error={serverError} />
      <div className="flex items-center justify-center">
        <div className="w-1/2">
          {isLoading && <div>Loading the calendar</div>}
        </div>
        <div className="bg-bg_color text-text_color">
          <FullCalendar
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
            initialView="timeGridWeek"
            weekNumbers={true}
            navLinks={true}
            height="auto"
            events={fetchEvents}
            eventClick={handleClick}
            lazyFetching={true}
            // views={{
            //   dayGridMonth: {
            //     fixedWeekCount: false,
            //   },
            //   timeGridWeek: {
            //     titleFormat: {
            //       year: "numeric",
            //       month: "long",
            //       day: "2-digit",
            //     },
            //     slotLabelFormat: {
            //       hour: "numeric",
            //       minute: "2-digit",
            //     },
            //     slotDuration: "00:15:00",
            //     slotMinTime: "08:00:00",
            //     slotMaxTime: "17:30:00",
            //     nowIndicator: true,
            //   },
            //   timeGridDay: {
            //     slotLabelFormat: {
            //       hour: "numeric",
            //       minute: "2-digit",
            //     },
            //     slotDuration: "00:15:00",
            //     slotMinTime: "08:00:00",
            //     slotMaxTime: "17:30:00",
            //     nowIndicator: true,
            //   },
            // }}
            // headerToolbar={{
            //   left: "dayGridMonth,timeGridWeek,timeGridDay,list",
            //   center: "title",
            //   right: "today,prev,next",
            // }}
            // allDaySlot={false}
            // eventTimeFormat={{
            //   hour: "numeric",
            //   minute: "2-digit",
            //   second: "2-digit",
            //   meridiem: false,
            // }}
            // buttonText={{
            //   month: "hónap",
            //   today: "ma",
            //   week: "hét",
            //   day: "nap",
            //   list: "lista",
            // }}
          />
        </div>
        <span ref={modalRef}>
          <CalendarModal
ref={modalRef}
            eventInfo={eventInfo}
            isOpen={isModalOpen}
            close={() => {
              setIsModalOpen(false);
            }}
error={setServerError}
          />
        </span>
      </div>
    </>
  );
}
