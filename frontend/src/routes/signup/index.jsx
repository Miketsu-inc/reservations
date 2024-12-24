import ProgressBar from "@components/ProgressBar";
import ServerError from "@components/ServerError";
import { useMultiStepForm } from "@lib/hooks";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import EmailForm from "./-components/EmailForm";
import NameForm from "./-components/NameForm";
import PasswordForm from "./-components/PasswordForm";
import PhoneNumberForm from "./-components/PhoneNumberForm";
import SubmissionCompleted from "./-components/SubmissionCompleted";

const defaultSignUpData = {
  firstName: "",
  lastName: "",
  email: "",
  phoneNum: "",
  password: "",
};

export const Route = createFileRoute("/signup/")({
  component: SingUpPage,
});

function SingUpPage() {
  const router = useRouter();
  const search = Route.useSearch();
  const [signUpData, setSignUpData] = useState(defaultSignUpData);
  const [isSubmitting, setIsSubmitting] = useState(false);
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
      sendInputData={signUpDataHandler}
      SubmitForm={handleSubmit}
      isCompleted={isCompletedHandler}
      isSubmitting={isSubmitting}
    />,
  ]);

  useEffect(() => {
    if (isSubmitting) {
      const sendRequest = async () => {
        try {
          const response = await fetch("/api/v1/auth/user/signup", {
            method: "POST",
            headers: {
              Accept: "application/json",
              "content-type": "application/json",
            },
            body: JSON.stringify(signUpData),
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
          setIsSubmitting(false);
        }
      };
      sendRequest();
    }
  }, [signUpData, isSubmitting, router, search]);

  function handleSubmit() {
    setIsSubmitting(true);
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
        className="flex w-full max-w-md flex-col px-10 shadow-sm sm:h-4/5 sm:min-h-1.5
          sm:rounded-xl sm:bg-layer_bg sm:pb-16 sm:pt-6 sm:shadow-lg"
      >
        <ProgressBar
          currentStep={stepIndex}
          stepCount={stepCount}
          isSubmitDone={isSubmitDone}
        />
        <ServerError error={serverError} />
        {isSubmitDone ? (
          <SubmissionCompleted text="You signed up successfully" />
        ) : (
          step
        )}
      </div>
    </div>
  );
}
