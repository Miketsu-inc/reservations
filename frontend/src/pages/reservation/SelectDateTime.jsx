import { useCallback, useEffect, useState } from "react";

export default function SelectDateTIme({ reservation, setReservation }) {
  const [freeHours, setFreeHours] = useState({ morning: [], afternoon: [] });

  const fetchHours = useCallback(async () => {
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
    await new Promise((resolve) => setTimeout(resolve, 100));
    setFreeHours(mockHours);
  }, []);

  useEffect(() => {
    if (reservation.day !== "" && reservation.service_id != 0) {
      fetchHours();
    }
  }, [reservation, fetchHours]);

  function hourChangeHandler(e) {
    const hour = e.target.value;
    setReservation((prev) => ({ ...prev, from_hour: hour }));
  }

  function dayChnageHandler(e) {
    const fromTimeStamp = Date.parse(e.target.value);
    const day = new Date(fromTimeStamp).toISOString();

    setReservation((prev) => ({
      ...prev,
      day: day,
    }));
  }

  return (
    <>
      <input
        type="date"
        onChange={dayChnageHandler}
        className="mt-4 block w-full rounded-md border border-text_color bg-layer_bg px-4 py-2
          text-base text-text_color shadow-sm hover:bg-hvr_gray focus:bg-hvr_gray
          focus:outline-none dark:[color-scheme:dark]"
      />

      {reservation.day && (
        <div className="flex flex-col gap-3">
          <h3 className="text-lg font-bold">Morning</h3>
          {freeHours.morning.length > 0 ? (
            <div className="grid w-full grid-cols-3 gap-3 rounded-md px-4 sm:grid-cols-5">
              {freeHours.morning.map((hour, index) => (
                <button
                  key={`morning-${index}`}
                  className={`cursor-pointer rounded-md bg-accent/90 py-1 font-bold text-black transition-all
                    hover:bg-accent/80
                    ${reservation.from_hour === hour ? "ring-2 ring-blue-500" : ""}`}
                  onClick={hourChangeHandler}
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
          {freeHours.afternoon.length > 0 ? (
            <div className="grid w-full grid-cols-3 gap-3 rounded-md px-4 sm:grid-cols-5">
              {freeHours.afternoon.map((hour, index) => (
                <button
                  key={`afternoon-${index}`}
                  className={`cursor-pointer rounded-md bg-accent/90 py-1 font-bold text-black transition-all
                    hover:bg-accent/80
                    ${reservation.from_hour === hour ? "ring-2 ring-blue-500" : ""}`}
                  onClick={hourChangeHandler}
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
