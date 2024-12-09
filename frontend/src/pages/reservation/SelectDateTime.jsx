import { useState } from "react";
import BackArrowIcon from "../../assets/icons/BackArrowIcon";
import Button from "../../components/Button";
import AvailableTimes from "./AvailableTimes";

export default function SelectDateTime({
  data,
  backArrowClick,
  sendDateTime,
  submit,
}) {
  const [selectedDay, setSelectedDay] = useState();
  const [selectedHour, setSelectedHour] = useState();

  function dayChangeHandler(e) {
    const fromTimeStamp = Date.parse(e.target.value);
    const isoString = new Date(fromTimeStamp).toISOString();

    setSelectedDay(isoString);
  }

  function reservationClickHandler() {
    const date = new Date(selectedDay);

    const [hours, minutes] = selectedHour.split(":").map(Number);
    date.setUTCHours(hours, minutes, 0, 0);

    const timeStamp = date.toISOString();

    sendDateTime({
      timeStamp: timeStamp,
    });
  }

  function selectedHourHandler(hour) {
    setSelectedHour(hour);
  }

  return (
    <>
      <button onClick={backArrowClick}>
        <BackArrowIcon />
      </button>
      <p className="pt-5 text-xl">Pick a date</p>
      <form method="POST" onSubmit={submit}>
        <div className="flex flex-col gap-6 pt-5 lg:pt-10">
          <input
            type="date"
            onChange={dayChangeHandler}
            className="mt-4 block w-full rounded-md border border-text_color bg-layer_bg px-4 py-2
              text-base text-text_color shadow-sm hover:bg-hvr_gray focus:bg-hvr_gray
              focus:outline-none dark:[color-scheme:dark]"
          />
          <AvailableTimes
            day={selectedDay}
            serviceId={data.service_id}
            locationId={data.location_id}
            selectHour={selectedHourHandler}
            clickedHour={selectedHour}
          />
          <Button
            onClick={reservationClickHandler}
            type="submit"
            styles="text-white dark:bg-transparent dark:border-2 border-secondary
              dark:text-secondary dark:hover:border-hvr_secondary
              dark:hover:text-hvr_secondary mt-6 font-semibold border-primary
              hover:bg-hvr_primary dark:focus:outline-none dark:focus:border-hvr_secondary
              dark:focus:text-hvr_secondary"
          >
            Reserve
          </Button>
        </div>
      </form>
    </>
  );
}
