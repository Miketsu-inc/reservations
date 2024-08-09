import { useEffect, useState } from "react";
import { useMultiStepForm } from "../../lib/hooks";
import EmailPage from "./EmailPage";
import PasswordPage from "./PasswordPage";
import PersonalInfo from "./PersonalInfo";
import ProgressBar from "./ProgressBar";
import SubmissionCompleted from "./SubmissionCompleted";

const defaultSignUpData = {
  firstName: "",
  lastName: "",
  email: "",
  password: "",
  confirmPassword: "",
};

export default function SingUp() {
  const [signUpData, setSignUpData] = useState(defaultSignUpData);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSubmitDone, setIsSubmitDone] = useState(false);
  const { step, stepIndex, nextStep } = useMultiStepForm([
    <PersonalInfo
      sendInputData={signUpDataHandler}
      isCompleted={isCompletedHandler}
    />,
    <EmailPage
      sendInputData={signUpDataHandler}
      isCompleted={isCompletedHandler}
    />,
    <PasswordPage
      sendInputData={signUpDataHandler}
      isCompleted={isCompletedHandler}
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
        className="flex min-h-screen w-full max-w-md flex-col px-10 shadow-sm sm:h-4/5 sm:min-h-1.5
          sm:rounded-md sm:bg-slate-400 sm:bg-opacity-5 sm:pb-16 sm:pt-6 sm:shadow-lg
          lg:px-8"
      >
        <ProgressBar step={stepIndex} />

        <form
          className="flex flex-col"
          method="POST"
          action=""
          autoComplete="on"
          onSubmit={handleSubmit}
        >
          {step}
        </form>
        {isSubmitDone ? (
          <SubmissionCompleted text="You signed up successfully" />
        ) : (
          <></>
        )}
      </div>
    </div>
  );
}
