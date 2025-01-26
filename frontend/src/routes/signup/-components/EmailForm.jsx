import Button from "@components/Button";
import FloatingLabelInput from "@components/FloatingLabelInput";
import { MAX_INPUT_LENGTH } from "@lib/constants";
import { useRef, useState } from "react";

const defaultEmailData = {
  email: {
    value: "",
    isValid: false,
  },
};

export default function EmailForm({ isCompleted, sendInputData }) {
  const emailRef = useRef();
  const [emailData, setEmailData] = useState(defaultEmailData);
  const [errorMessage, setErrorMessage] = useState("Please enter your email!");

  function emailValidation(email) {
    if (email.length > MAX_INPUT_LENGTH) {
      setErrorMessage(`Inputs must be ${MAX_INPUT_LENGTH} characters or less!`);
      return false;
    }
    if (!email.includes("@")) {
      setErrorMessage("Please enter a valid email!");
      return false;
    }
    return true;
  }
  function handleInputData(data) {
    setEmailData((prevEmailData) => ({
      ...prevEmailData,
      [data.name]: {
        value: data.value,
        isValid: data.isValid,
      },
    }));
  }

  function handleClick() {
    if (!emailData.email.isValid) {
      emailRef.current.triggerValidationError();
    } else {
      sendInputData({
        email: emailData.email.value,
      });
      isCompleted(true);
    }
  }

  return (
    <>
      <h2 className="mt-2 text-center text-xl font-semibold">Email</h2>
      <p className="mb-8 py-2 text-center text-gray-600 dark:text-gray-300">
        Enter your email to get started with creating your account
      </p>
      <FloatingLabelInput
        ref={emailRef}
        type="text"
        name="email"
        id="emailInput"
        autoComplete="email"
        labelText="Email"
        errorText={errorMessage}
        inputValidation={emailValidation}
        inputData={handleInputData}
      />
      <div className="mt-10 flex items-center justify-center">
        <Button
          variant="primary"
          styles="w-full py-2"
          type="button"
          onClick={handleClick}
          buttonText="Continue"
        />
      </div>
    </>
  );
}
