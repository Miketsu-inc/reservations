import { useMultiStepForm } from "@lib/hooks";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import AppointmentsAdder from "./-components/AppointmentsAdder";
import LocationForm from "./-components/LocationForm";
import MerchantInfoForm from "./-components/MerchantInfoForm";

export const Route = createFileRoute("/_authenticated/merchantsignup/")({
  component: MerchantSignup,
});

function MerchantSignup() {
  const navigate = useNavigate({ from: "/merchantsignup" });
  const [isSubmitDone, setIsSubmitDone] = useState(false);
  const { step, _, nextStep } = useMultiStepForm([
    <MerchantInfoForm key="companyInfoForm" isCompleted={isCompletedHandler} />,
    <LocationForm
      key="locationForm"
      isCompleted={isCompletedHandler}
      isSubmitDone={setIsSubmitDone}
    />,
    <AppointmentsAdder
      key="appointmentsAdder"
      redirect={() => navigate({ to: "/dashboard" })}
    />,
  ]);

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
