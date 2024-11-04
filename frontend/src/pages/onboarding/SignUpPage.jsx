import { useEffect, useState } from "react";
import ServerError from "../../components/ServerError";
import { useMultiStepForm } from "../../lib/hooks";
import EmailForm from "./EmailForm";
import NameForm from "./NameForm";
import PasswordForm from "./PasswordForm";
import PhoneNumForm from "./PhoneNumForm";
import ProgressBar from "./ProgressBar";
import SubmissionCompleted from "./SubmissionCompleted";

const defaultSignUpData = {
  firstName: "",
  lastName: "",
  email: "",
  phoneNum: "",
  password: "",
};

export default function SingUpPage() {
  const [signUpData, setSignUpData] = useState(defaultSignUpData);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSubmitDone, setIsSubmitDone] = useState(false);
  const [serverError, setServerError] = useState(undefined);
  const { step, stepIndex, nextStep, stepCount } = useMultiStepForm([
    <EmailForm
      key="emailForm"
      sendInputData={signUpDataHandler}
      isCompleted={isCompletedHandler}
    />,
    <PhoneNumForm
      key="phoneNumForm"
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
      console.log(signUpData);
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
          const result = await response.json();
          if (result.error) {
            setServerError(result.error);
            return;
          } else {
            setServerError(undefined);
            setIsSubmitDone(true);
          }
        } catch (err) {
          setServerError("An error occurred. Please try again.");
        } finally {
          setIsSubmitting(false);
        }
      };
      sendRequest();
    }
  }, [signUpData, isSubmitting]);

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
        className="flex min-h-screen w-full max-w-md flex-col px-10 shadow-sm sm:h-4/5 sm:min-h-1.5
          sm:rounded-3xl sm:bg-layer_bg sm:pb-16 sm:pt-6 sm:shadow-lg"
      >
        <ProgressBar
          currentStep={stepIndex}
          stepCount={stepCount}
          isSubmitDone={isSubmitDone}
        />
        <ServerError styles="" error={serverError} />
        {isSubmitDone ? (
          <SubmissionCompleted text="You signed up successfully" />
        ) : (
          step
        )}
      </div>
    </div>
  );
}
