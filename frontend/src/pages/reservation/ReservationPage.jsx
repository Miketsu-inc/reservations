import { useEffect, useState } from "react";

const defaultReservation = {
  user: "DEFAULT USER",
  shop: "DEFAULT SHOP",
  reservationType: "",
  date: "",
};

const reservationTypes = ["hair", "nail", "face"];

export default function ReservationPage() {
  const [reservation, setReservation] = useState(defaultReservation);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");

  useEffect(() => {
    if (isSubmitting) {
      setIsSubmitting(false);

      fetch("/api/v1/reservations", {
        method: "POST",
        headers: {
          "Content-type": "application/json; charset=UTF-8",
        },
        body: JSON.stringify({
          user: reservation.user,
          shop: reservation.shop,
          reservationType: reservation.reservationType,
          date: reservation.date,
        }),
      });
    }
  }, [isSubmitting, reservation]);

  function onSubmitHandler(e) {
    e.preventDefault();
    setErrorMessage("");

    if (reservation.date === "") {
      setErrorMessage("Please set a reservation date!");
    }

    if (reservation.reservationType === "") {
      setErrorMessage("Please set a reservation type!");
    }

    setIsSubmitting(true);
  }

  function dateChangeHandler(e) {
    setReservation((prev) => ({ ...prev, date: e.target.value }));
  }

  function typeChangeHandler(e) {
    setReservation((prev) => ({ ...prev, reservationType: e.target.value }));
  }

  return (
    <>
      <div className="flex flex-col items-center gap-6 md:flex-row-reverse">
        <div>
          <img src="https://dummyimage.com/1920x1080/d156c3/000000.jpg"></img>
        </div>
        <div className="flex flex-col gap-2">
          <h1 className="text-center">Company name</h1>
          <p className="text-justify">
            Short description about the core values of the company, maybe also
            what they belive in. How they do their buisness. What they would
            like to achive in the future. I'm basically just bullshiting at this
            point. Should have used lorem ipsum
          </p>
          <p className="text-center">Open hours: 9:00-19:00</p>
        </div>
      </div>
      <form method="POST" onSubmit={onSubmitHandler}>
        <div className="flex flex-col gap-2 pt-10">
          <select
            defaultValue="default"
            onChange={typeChangeHandler}
            className="text-custombg"
          >
            <option value="default" disabled hidden>
              Choose a reservation type
            </option>
            {reservationTypes.map((type, index) => (
              <option key={index} value={type} className="text-custombg">
                {type}
              </option>
            ))}
          </select>
          <input
            type="datetime-local"
            onChange={dateChangeHandler}
            className="text-custombg"
          ></input>
          <button type="submit" className="border-2">
            Reserve
          </button>
          {errorMessage ? (
            <p className="text-center text-red-500">{errorMessage}</p>
          ) : (
            <></>
          )}
        </div>
      </form>
    </>
  );
}
