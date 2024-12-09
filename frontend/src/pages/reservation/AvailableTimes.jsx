import { useEffect, useState } from "react";

const defaultAvailableTimes = {
  morning: [],
  afternoon: [],
};

export default function AvailableTimes({
  day,
  serviceId,
  locationId,
  selectHour,
  clickedHour,
}) {
  const [availableTimes, setAvailableTimes] = useState(defaultAvailableTimes);

  useEffect(() => {
    async function fetchHours() {
      // try {
      //   const response = await fetch(
      //     `/api/v1/merchants/info?merchant=${merchantName}&service=${reservation.service_id}&day=${reservation.day}&location=${reservation.location_id}`,
      //     {
      //       method: "GET",
      //     }
      //   );

      //   const result = await response.json();

      //   if (!response.ok) {
      //     setServerError(result.error.message);
      //   } else {
      //     setServerError(undefined);

      //     setFreeHours(result.data)
      //   }
      // } catch (err) {
      //   setServerError(err.message);
      // }
      const mockHours = {
        morning: [
          "08:30",
          "09:00",
          "09:30",
          "10:00",
          "10:30",
          "11:00",
          "11:30",
          "12:00",
        ],
        afternoon: [
          "12:30",
          "13:00",
          "13:30",
          "14:00",
          "14:30",
          "15:00",
          "15:30",
          "16:00",
        ],
      };

      setAvailableTimes(mockHours);
    }

    fetchHours();
  }, [day, serviceId, locationId]);

  function hourClickHandler(e) {
    selectHour(e.target.value);
  }

  return (
    <>
      {day && (
        <div className="flex flex-col gap-3">
          <h3 className="text-lg font-bold">Morning</h3>
          {availableTimes.morning.length > 0 ? (
            <div className="grid w-full grid-cols-2 gap-3 rounded-md px-4 sm:grid-cols-5">
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

          <h3 className="text-lg font-bold">Afternoon</h3>
          {availableTimes.afternoon.length > 0 ? (
            <div className="grid w-full grid-cols-2 gap-3 rounded-md px-4 sm:grid-cols-5">
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
