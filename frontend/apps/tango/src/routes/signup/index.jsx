import { ProgressBar, ServerError } from "@reservations/components";
import { useMultiStepForm } from "@reservations/lib";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import EmailForm from "./-components/EmailForm";
import NameForm from "./-components/NameForm";
import PasswordForm from "./-components/PasswordForm";
import PhoneNumberForm from "./-components/PhoneNumberForm";
import SubmissionCompleted from "./-components/SubmissionCompleted";

const defaultSignUpData = {
  first_name: "",
  last_name: "",
  email: "",
  phone_number: "",
  password: "",
};

export const Route = createFileRoute("/signup/")({
  component: SingUpPage,
});

function SingUpPage() {
  const router = useRouter();
  const search = Route.useSearch();
  const [signUpData, setSignUpData] = useState(defaultSignUpData);
  const [isLoading, setIsLoading] = useState(false);
  const [isSubmitDone, setIsSubmitDone] = useState(false);
  const [serverError, setServerError] = useState("");
  const { step, stepIndex, nextStep, stepCount } = useMultiStepForm([
    <EmailForm
      key="emailForm"
      sendInputData={signUpDataHandler}
      isCompleted={isCompletedHandler}
    />,
    <PhoneNumberForm
      key="phoneNumberForm"
      sendInputData={signUpDataHandler}
      isCompleted={isCompletedHandler}
    />,
    <PasswordForm
      key="passwordForm"
      sendInputData={signUpDataHandler}
      isCompleted={isCompletedHandler}
    />,
    <NameForm
      key="nameForm"
      SubmitForm={handleSubmit}
      isCompleted={isCompletedHandler}
      isLoading={isLoading}
    />,
  ]);

  async function handleSubmit(firstName, lastName) {
    setIsLoading(true);
    try {
      const response = await fetch("/api/v1/auth/signup/user", {
        method: "POST",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          first_name: firstName,
          last_name: lastName,
          email: signUpData.email,
          password: signUpData.password,
          phone_number: signUpData.phone_number,
        }),
      });

      if (!response.ok) {
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        setServerError();
        setIsSubmitDone(true);

        if (search.redirect) {
          router.history.push(search.redirect);
        } else {
          router.navigate({ from: Route.fullPath, to: "/" });
        }
      }
    } catch (err) {
      setServerError(err.message);
    } finally {
      setIsLoading(false);
    }
  }

  function signUpDataHandler(data) {
    setSignUpData((prev) => {
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
        className="sm:bg-layer_bg flex w-full max-w-md flex-col px-8 shadow-sm
          sm:rounded-xl sm:pt-6 sm:pb-16 sm:shadow-lg"
      >
        <ProgressBar
          currentStep={stepIndex}
          stepCount={stepCount}
          isSubmitDone={isSubmitDone}
        />
        <ServerError error={serverError} styles="mb-4" />
        {isSubmitDone ? (
          <SubmissionCompleted text="You signed up successfully" />
        ) : (
          step
        )}
      </div>
    </div>
  );
}
