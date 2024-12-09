import { useParams } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import ServerError from "../../components/ServerError";
import { useMultiStepForm } from "../../lib/hooks";
import MerchantInfo from "./MerchantInfo";
import SelectDateTime from "./SelectDateTime";

const defaultReservation = {
  merchant_name: "Hair salon",
  service_id: 0,
  location_id: 0,
  timeStamp: "",
};

const defaultMerchantInfo = {
  merchantName: "",
  shortLocation: "",
  contact_email: "",
  shortDescription: "",
  parkingInfo: "",
  aboutUs: "",
  annoucement: "",
  services: [],
};

export default function MerchantPage() {
  const [reservation, setReservation] = useState(defaultReservation);
  const [merchantInfo, setMerchantInfo] = useState(defaultMerchantInfo);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [serverError, setServerError] = useState(undefined);
  const { merchantName } = useParams({ strict: false });
  const { step, nextStep, previousStep } = useMultiStepForm([
    <MerchantInfo
      data={merchantInfo}
      sendServiceId={reservationDataHandler}
      isCompleted={serviceSelectionCompletedHandler}
      key="MerchantInfo"
    />,
    <SelectDateTime
      submit={onSubmitHandler}
      data={reservation}
      backArrowClick={backArrowHandler}
      sendDateTime={reservationDataHandler}
      key="SelectDateTime"
    />,
  ]);

  useEffect(() => {
    async function fetchMerchantInfo() {
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
            timeStamp: 0,
          });

          const shortLocation =
            result.data.address +
            ", " +
            result.data.city +
            " " +
            result.data.postal_code;

          setMerchantInfo({
            merchantName: result.data.merchant_name,
            contact_email: result.data.contact_email,
            shortLocation: shortLocation,
            services: result.data.services,
          });
        }
      } catch (err) {
        setServerError(err.message);
      }
    }

    fetchMerchantInfo();
  }, [merchantName]);

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
            setServerError(result.error.message);
          }
        } catch (err) {
          setServerError(err.message);
        } finally {
          setIsSubmitting(false);
        }
      };

      sendRequest();
    }
  }, [isSubmitting, reservation]);

  function reservationDataHandler(data) {
    setReservation((prev) => {
      return { ...prev, ...data };
    });
  }

  function serviceSelectionCompletedHandler() {
    nextStep();
  }

  function backArrowHandler() {
    previousStep();
  }

  function onSubmitHandler(e) {
    e.preventDefault();
    setServerError("");

    let canSubmit = true;

    console.log(reservation);

    if (reservation.timeStamp === "") {
      setServerError("Please set a reservation date!");
      canSubmit = false;
    }

    if (reservation.service_id === 0) {
      setServerError("Please set a reservation type!");
      canSubmit = false;
    }

    setIsSubmitting(canSubmit);
  }

  return (
    <div className="mx-auto min-h-screen max-w-screen-xl bg-layer_bg px-10">
      <ServerError styles="my-4" error={serverError} />
      {step}
    </div>
  );
}
