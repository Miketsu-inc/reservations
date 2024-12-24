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
      <h2 className="mt-8 py-2 text-center text-xl sm:mt-4">Email</h2>
      <p className="py-2 text-center">
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
      <p className="px-1 pt-4 text-center text-sm tracking-tight">
        After giving your email you'll get a <u>verification email</u>. Please
        check your inbox and verify your email.
      </p>
      <div className="flex items-center justify-center">
        <Button
          styles="mt-6 w-full font-semibold focus-visible:outline-1 hover:bg-hvr_primary
            text-white py-2"
          type="button"
          onClick={handleClick}
          buttonText="Verify email"
        />
      </div>
    </>
  );
}
