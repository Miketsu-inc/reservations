import { meQueryOptions, useMultiStepForm } from "@reservations/lib";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import LocationPicker from "./-components/LocationPicker";
import MerchantInfoForm from "./-components/MerchantInfoForm";

export const Route = createFileRoute("/_authenticated/signup/")({
  component: MerchantSignup,
});

function MerchantSignup() {
  const navigate = useNavigate({ from: Route.fullPath });
  const { queryClient } = Route.useRouteContext();
  const [isSubmitDone, setIsSubmitDone] = useState(false);
  const { step, stepIndex, nextStep } = useMultiStepForm([
    <MerchantInfoForm key="companyInfoForm" isCompleted={isCompletedHandler} />,
    <LocationPicker
      key="locationForm"
      isCompleted={isCompletedHandler}
      isSubmitDone={setIsSubmitDone}
      redirect={() => navigate({ to: "/dashboard" })}
    />,
  ]);

  async function isCompletedHandler(isCompleted) {
    if (isCompleted) {
      if (stepIndex === 0) {
        await queryClient.invalidateQueries(meQueryOptions());
      }
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
