import { useEffect, useState } from "react";
import { useMultiStepForm } from "../../lib/hooks";
import EmailForm from "./EmailForm";
import NameForm from "./NameForm";
import PasswordForm from "./PasswordForm";
import ProgressBar from "./ProgressBar";
import SubmissionCompleted from "./SubmissionCompleted";

const defaultSignUpData = {
  firstName: "",
  lastName: "",
  email: "",
  password: "",
  confirmPassword: "",
};

export default function SingUpPage() {
  const [signUpData, setSignUpData] = useState(defaultSignUpData);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSubmitDone, setIsSubmitDone] = useState(false);
  const { step, stepIndex, nextStep } = useMultiStepForm([
    <NameForm
      sendInputData={signUpDataHandler}
      isCompleted={isCompletedHandler}
    />,
    <EmailForm
      sendInputData={signUpDataHandler}
      isCompleted={isCompletedHandler}
    />,
    <PasswordForm
      sendInputData={signUpDataHandler}
      isCompleted={isCompletedHandler}
      isSubmitting={isSubmitting}
    />,
    <></>,
  ]);

  useEffect(() => {
    if (isSubmitting) {
      setIsSubmitting(false);
      // send POST request
      console.log(signUpData);
      setIsSubmitDone(true);
    }
  }, [signUpData, isSubmitting]);

  function handleSubmit(e) {
    e.preventDefault();
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
        className="sm:bg-layer-bg flex min-h-screen w-full max-w-md flex-col px-10 shadow-sm
sm:h-4/5 sm:min-h-1.5 sm:rounded-md sm:pb-16 sm:pt-6 sm:shadow-lg lg:px-8"
      >
        <ProgressBar isSubmitDone={isSubmitDone} step={stepIndex} />

        <form
          className="flex flex-col"
          method="POST"
          action=""
          autoComplete="on"
          onSubmit={handleSubmit}
        >
          {isSubmitDone ? (
            <SubmissionCompleted text="You signed up successfully" />
          ) : (
            step
          )}
        </form>
      </div>
    </div>
  );
}
