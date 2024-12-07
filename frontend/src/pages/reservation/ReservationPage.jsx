import { useParams } from "@tanstack/react-router";
import { useCallback, useEffect, useState } from "react";
import Button from "../../components/Button";
import Selector from "../../components/Selector";
import SelectorItem from "../../components/SelectorItem";
import ServerError from "../../components/ServerError";
import SelectDateTime from "./SelectDateTime";

const defaultReservation = {
  merchant_name: "Hair salon",
  service_id: 0,
  location_id: 0,
  day: "",
  from_hour: "",
};

export default function ReservationPage() {
  const [reservation, setReservation] = useState(defaultReservation);
  const [services, setServices] = useState({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");
  const [serverError, setServerError] = useState(undefined);
  const { merchantName } = useParams({ strict: false });

  const fetchMerchantInfo = useCallback(async () => {
    try {
      const response = await fetch(
        `/api/v1/merchants/info?name=${merchantName}`,
        {
          method: "GET",
        }
      );

      const result = await response.json();

      if (!response.ok) {
        setServerError(result.error.message);
      } else {
        setServerError(undefined);

        setReservation({
          merchant_name: result.data.merchant_name,
          service_id: 1,
          location_id: result.data.location_id,
          from_date: "",
          to_date: "",
        });

        result.data.services.forEach((s) => {
          setServices((prev) => ({
            ...prev,
            [s.name]: s.ID,
          }));
        });
      }
    } catch (err) {
      setServerError(err.message);
    }
  }, [merchantName]);

  useEffect(() => {
    fetchMerchantInfo();
  }, [fetchMerchantInfo]);

  useEffect(() => {
    if (isSubmitting) {
      const sendRequest = async () => {
        try {
          const response = await fetch("/api/v1/appointments", {
            method: "POST",
            headers: {
              "Content-type": "application/json; charset=UTF-8",
            },
            body: JSON.stringify(reservation),
          });

          if (!response.ok) {
            const result = await response.json();
            setErrorMessage(result.error.message);
          }
        } catch (err) {
          setErrorMessage(err.message);
        } finally {
          setIsSubmitting(false);
        }
      };

      sendRequest();
    }
  }, [isSubmitting, reservation]);

  function onSubmitHandler(e) {
    e.preventDefault();
    setErrorMessage("");

    let canSubmit = true;

    if (reservation.day === "") {
      setErrorMessage("Please set a reservation date!");
      canSubmit = false;
    }

    if (reservation.from_hour === "") {
      setErrorMessage("Please set a reservation date");
      canSubmit = false;
    }

    if (reservation.service_id === 0) {
      setErrorMessage("Please set a reservation type!");
      canSubmit = false;
    }

    setIsSubmitting(canSubmit);
  }

  function serviceChangeHandler(value) {
    setReservation((prev) => ({ ...prev, service_id: services[value] }));
  }

  return (
    <>
      <div className="flex flex-col items-center gap-6 bg-layer_bg md:flex-row-reverse">
        <div>
          <img src="https://dummyimage.com/1920x1080/d156c3/000000.jpg"></img>
        </div>
        <div className="flex flex-col gap-2 px-4 pb-4">
          <h1 className="text-center">{merchantName}</h1>
          <p className="text-justify">
            Short description about the core values of the company, maybe also
            what they belive in. How they do their buisness. What they would
            like to achive in the future. I'm basically just bullshiting at this
            point. Should have used lorem ipsum
          </p>
          <p className="text-center">Open hours: 9:00-19:00</p>
          <p className="text-center">Email: </p>
        </div>
      </div>
      <form method="POST" onSubmit={onSubmitHandler}>
        <div className="flex flex-col gap-2 px-10 pt-10">
          <ServerError styles="mt-4 mb-2" error={serverError} />
          <Selector
            defaultValue="Choose a reservation type"
            styles="text-base p-1 rounded-lg border-text_color border px-3"
            dropdownStyles="rounded-md bg-layer_bg border border-gray-600 w-full translate-y-[2.9rem] absolute sm:h-32 h-28 md:h-44"
            onSelect={serviceChangeHandler}
          >
            {Object.entries(services).map(([service, id]) => (
              <SelectorItem key={id} value={service} styles="text-base py-1">
                {service}
              </SelectorItem>
            ))}
          </Selector>

          <SelectDateTime
            reservation={reservation}
            setReservation={setReservation}
          />

          <Button
            type="submit"
            styles="text-white dark:bg-transparent dark:border-2 border-secondary
              dark:text-secondary dark:hover:border-hvr_secondary
              dark:hover:text-hvr_secondary mt-6 font-semibold border-primary
              hover:bg-hvr_primary dark:focus:outline-none dark:focus:border-hvr_secondary
              dark:focus:text-hvr_secondary"
          >
            Reserve
          </Button>
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
