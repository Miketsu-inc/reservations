import { useCallback, useEffect, useState } from "react";

const defaultAvailableTimes = {
  morning: [],
  afternoon: [],
};

export default function AvailableTimes({
  day,
  serviceId,
  selectHour,
  clickedHour,
  merchant_name,
  setServerError,
}) {
  const [availableTimes, setAvailableTimes] = useState(defaultAvailableTimes);

  function convertTimes(times) {
    const morning = times.filter((time) => {
      const hour = parseInt(time.split(":")[0]);
      return hour < 12;
    });

    const afternoon = times.filter((time) => {
      const hour = parseInt(time.split(":")[0]);
      return hour >= 12;
    });

    setAvailableTimes({ morning, afternoon });
  }

  const fetchHours = useCallback(
    async (day, serviceId, merchant_name) => {
      try {
        const response = await fetch(
          `/api/v1/merchants/times?name=${merchant_name}&service_id=${serviceId}&day=${day}`,
          {
            method: "GET",
          }
        );

        const result = await response.json();

        if (!response.ok) {
          setServerError(result.error.message);
        } else {
          setServerError(undefined);
          convertTimes(result.data);
        }
      } catch (err) {
        setServerError(err.message);
      }
    },
    [setServerError]
  );

  useEffect(() => {
    if (day !== undefined) {
      fetchHours(day, serviceId, merchant_name);
    }
  }, [day, serviceId, merchant_name, fetchHours]);

  function hourClickHandler(e) {
    selectHour(e.target.value);
  }

  return (
    <>
      {day && (
        <div className="flex flex-col gap-3">
          <p className="text-lg font-bold">Morning</p>
          {availableTimes.morning.length > 0 ? (
            <div className="grid w-full grid-cols-2 gap-3 rounded-md sm:grid-cols-5">
              {availableTimes.morning.map((hour, index) => (
                <button
                  key={`morning-${index}`}
                  className={`cursor-pointer rounded-md bg-accent/90 py-1 font-bold text-black transition-all
                    hover:bg-accent/80 ${clickedHour === hour ? "ring-2 ring-blue-500" : ""}`}
                  onClick={hourClickHandler}
                  value={hour}
                  type="button"
                >
                  {hour}
                </button>
              ))}
            </div>
          ) : (
            <p className="text-md flex items-center justify-center font-bold">
              No available morning hours for this day
            </p>
          )}

          <p className="text-lg font-bold">Afternoon</p>
          {availableTimes.afternoon.length > 0 ? (
            <div className="grid w-full grid-cols-2 gap-3 rounded-md sm:grid-cols-5">
              {availableTimes.afternoon.map((hour, index) => (
                <button
                  key={`afternoon-${index}`}
                  className={`cursor-pointer rounded-md bg-accent/90 py-1 font-bold text-black transition-all
                    hover:bg-accent/80 ${clickedHour === hour ? "ring-2 ring-blue-500" : ""}`}
                  onClick={hourClickHandler}
                  value={hour}
                  type="button"
                >
                  {hour}
                </button>
              ))}
            </div>
          ) : (
            <p className="text-md flex items-center justify-center font-bold">
              No available afternoon hours for this day
            </p>
          )}
        </div>
      )}
    </>
  );
}
