import { useState } from "react";
import { DayPicker } from "react-day-picker";
import "react-day-picker/style.css";
import BackArrowIcon from "../../assets/icons/BackArrowIcon";
import Button from "../../components/Button";
import ServerError from "../../components/ServerError";
import AvailableTimes from "./AvailableTimes";

export default function SelectDateTime({
  data,
  backArrowClick,
  sendDateTime,
  submit,
  isSubmitting,
}) {
  const [selectedDay, setSelectedDay] = useState();
  const [selectedHour, setSelectedHour] = useState();
  const [serverError, setServerError] = useState();

  function dayChangeHandler(date) {
    setSelectedDay(date.toISOString().split("T")[0]);
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
    <div className="py-5">
      <button
        className="inline-flex gap-1 hover:underline"
        onClick={backArrowClick}
      >
        <BackArrowIcon />
        Back
      </button>
      <ServerError error={serverError} />
      <form method="POST" onSubmit={submit}>
        <div className="flex flex-col pt-5 md:flex-row md:gap-10 lg:pt-10">
          <div className="flex flex-col gap-6 md:w-1/2">
            <p className="py-5 text-xl">Pick a date</p>
            <div className="self-center md:self-auto">
              <DayPicker
                mode="single"
                timeZone="UTC"
                selected={selectedDay}
                onSelect={dayChangeHandler}
              />
            </div>
            {selectedDay && (
              <>
                <hr className="border-gray-500" />
                <p className="py-5 text-xl">Pick a Time</p>
                <AvailableTimes
                  day={selectedDay}
                  serviceId={data.service_id}
                  merchant_name={data.merchant_name}
                  selectHour={selectedHourHandler}
                  clickedHour={selectedHour}
                  setServerError={setServerError}
                />
              </>
            )}
          </div>
          <div className="pt-8 md:flex md:w-1/2 md:flex-col md:pt-0">
            <div className="hidden md:flex md:flex-col md:gap-6">
              <p className="py-5 text-xl">Summary</p>
              <div className="text-lg *:grid *:grid-cols-2">
                <div>
                  <p>Merchant:</p>
                  <p>{data.merchant_name}</p>
                </div>
                <div>
                  <p>Service:</p>
                  <p>{data.service_id}</p>
                </div>
                <div>
                  <p>Location:</p>
                  <p>{data.location_id}</p>
                </div>
                <div className={`${selectedDay ? "" : "invisible"}`}>
                  <p>Date:</p>
                  <p>{selectedDay}</p>
                </div>
                <div className={`${selectedHour ? "" : "invisible"}`}>
                  <p>Time:</p>
                  <p>{selectedHour}</p>
                </div>
              </div>
            </div>
            <div className="md:pt-28">
              <Button
                onClick={reservationClickHandler}
                type="submit"
                disabled={selectedDay && selectedHour ? false : true}
                isLoading={isSubmitting}
                buttonText="Reserve"
                styles="w-full text-white dark:bg-transparent dark:border-2 border-secondary
                  dark:text-secondary dark:hover:border-hvr_secondary
                  dark:hover:text-hvr_secondary font-semibold border-primary hover:bg-hvr_primary
                  dark:focus:outline-none dark:focus:border-hvr_secondary
                  dark:focus:text-hvr_secondary"
              ></Button>
            </div>
          </div>
        </div>
      </form>
    </div>
  );
}
