import { useEffect, useState } from "react";
import { useMultiStepForm } from "../../lib/hooks";
import AppointmentsAdder from "./AppointmentsAdder";
import LocationForm from "./LocationForm";
import MerchantInfoForm from "./MerchantInfoForm";

const defaultMerchantData = {
  company_name: "",
  owner: "",
  contact_email: "",
  country: "",
  postal_code: "",
  city: "",
  address: "",
};

export default function MerchantSignup() {
  const [MerchantData, setMerchantData] = useState(defaultMerchantData);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSubmitDone, setIsSubmitDone] = useState(false);

  const { step, _, nextStep } = useMultiStepForm([
    <MerchantInfoForm
      key="companyInfoForm"
      sendInputData={merchantDataHandler}
      isCompleted={isCompletedHandler}
    />,
    <LocationForm
      key="locationForm"
      sendInputData={merchantDataHandler}
      submitForm={handleSubmit}
      isSubmitting={isSubmitting}
    />,
    <AppointmentsAdder key="appointmentsAdder" />,
  ]);

  useEffect(() => {
    if (isSubmitting) {
      const sendRequest = async () => {
        try {
          const response = await fetch("/api/v1/auth/merchantSignup", {
            method: "POST",
            headers: {
              Accept: "application/json",
              "content-type": "application/json",
            },
            body: JSON.stringify(MerchantData),
          });
          const result = await response.json();
          if (result.error) {
            return;
          } else {
            nextStep();
            setIsSubmitDone(true);
          }
        } catch (err) {
          console.error("Error messsage from server:", err.message);
          return;
        } finally {
          setIsSubmitting(false);
        }
      };
      sendRequest();
    }
  }, [MerchantData, isSubmitting, nextStep]);

  function handleSubmit() {
    setIsSubmitting(true);
  }

  function merchantDataHandler(data) {
    setMerchantData((prev) => {
      return { ...prev, ...data };
    });
  }
  function isCompletedHandler(isCompleted) {
    if (isCompleted) {
      nextStep();
    }
  }

  return (
    <div
      className={`${!isSubmitDone ? "min-h-screen min-w-min items-center" : ""} flex flex-col
        justify-center`}
    >
      <div
        className={`${
          !isSubmitDone
            ? `flex min-h-screen w-full max-w-md flex-col px-10 shadow-sm sm:h-4/5 sm:min-h-1.5
              sm:rounded-3xl sm:bg-layer_bg sm:pb-16 sm:pt-6 sm:shadow-lg`
            : ""
          } `}
      >
        {step}
      </div>
    </div>
  );
}
