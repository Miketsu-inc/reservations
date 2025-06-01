import { useMultiStepForm } from "@lib/hooks";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import LocationForm from "./-components/LocationForm";
import MerchantInfoForm from "./-components/MerchantInfoForm";

export const Route = createFileRoute("/_authenticated/merchantsignup/")({
  component: MerchantSignup,
});

function MerchantSignup() {
  const navigate = useNavigate({ from: Route.fullPath });
  const [isSubmitDone, setIsSubmitDone] = useState(false);
  const { step, _, nextStep } = useMultiStepForm([
    <MerchantInfoForm key="companyInfoForm" isCompleted={isCompletedHandler} />,
    <LocationForm
      key="locationForm"
      isCompleted={isCompletedHandler}
      isSubmitDone={setIsSubmitDone}
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
            ? `sm:bg-layer_bg flex w-full max-w-md flex-col px-8 shadow-sm sm:rounded-xl
              sm:pt-6 sm:pb-10 sm:shadow-lg`
            : ""
          } `}
      >
        {step}
      </div>
    </div>
  );
}
