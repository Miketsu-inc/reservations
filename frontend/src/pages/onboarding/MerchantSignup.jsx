import { useEffect, useState } from "react";
import { useMultiStepForm } from "../../lib/hooks";
import AppointmentForm from "./AppointmentForm";
import LocationForm from "./LocationForm";
import MerchantInfoForm from "./MerchantInfoForm";

const defaultMerchantData = {
  company_name: "",
  owner: "",
  contact_email: "",
  appointment_type: "",
  duration: "",
  price: "",
  country: "",
  postal_code: "",
  city: "",
  address: "",
};

export default function MerchantSignup() {
  const [MerchantData, setMerchantData] = useState(defaultMerchantData);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const { step, stepIndex, nextStep } = useMultiStepForm([
    <MerchantInfoForm
      key="companyInfoForm"
      sendInputData={merchantDataHandler}
      isCompleted={isCompletedHandler}
    />,
    <AppointmentForm
      key="appointmentFrom"
      sendInputData={merchantDataHandler}
      isCompleted={isCompletedHandler}
    />,
    <LocationForm
      key="locationForm"
      sendInputData={merchantDataHandler}
      isCompleted={isCompletedHandler}
      submitForm={handleSubmit}
      isSubmitting={isSubmitting}
    />,
  ]);

  useEffect(() => {
    if (isSubmitting) {
      console.log(MerchantData);
      // const sendRequest = async () => {
      //   try {
      //     const response = await fetch("/api/v1/auth/merchantSignup", {
      //       method: "POST",
      //       headers: {
      //         Accept: "application/json",
      //         "content-type": "application/json",
      //       },
      //       body: JSON.stringify(signUpData),
      //     });
      //     const result = await response.json();
      //     if (result.error) {
      //       console.log(result);
      //       return;
      //     } else {
      //       console.log(result);
      //     }
      //   } catch (err) {
      //     console.error("Error messsage from server:", err.message);
      //   } finally {
      //     setIsSubmitting(false);
      //   }
      // };
      // sendRequest();
    }
  }, [MerchantData, isSubmitting]);

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
    <div className="flex min-h-screen min-w-min items-center justify-center">
      <div
        className="mt-8 flex min-h-screen w-full max-w-md flex-col items-center justify-center
          sm:h-4/5 sm:min-h-1.5 sm:rounded-3xl sm:bg-layer_bg sm:pb-16 sm:pt-6"
      >
        {step}
      </div>
    </div>
  );
}
