import { useMultiStepForm } from "@reservations/lib";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import LocationPicker from "./-components/LocationPicker";
import MerchantInfoForm from "./-components/MerchantInfoForm";

export const Route = createFileRoute("/_authenticated/signup/")({
  component: MerchantSignup,
});

function MerchantSignup() {
  const navigate = useNavigate({ from: Route.fullPath });
  const [isSubmitDone, setIsSubmitDone] = useState(false);
  const { step, _, nextStep } = useMultiStepForm([
    <MerchantInfoForm key="companyInfoForm" isCompleted={isCompletedHandler} />,
    <LocationPicker
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
      className={`${!isSubmitDone ? "min-h-screen min-w-min items-center" : ""}
        flex flex-col justify-center px-4`}
    >
      {step}
    </div>
  );
}
