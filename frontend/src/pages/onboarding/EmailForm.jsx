import { useRef, useState } from "react";
import Button from "../../components/Button";
import Input from "../../components/Input";
import { MAX_INPUT_LENGTH } from "../../lib/constants";

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
      <h2 className="mt-8 py-2 text-2xl sm:mt-4">Enter your email</h2>
      <Input
        styles=""
        ref={emailRef}
        type="text"
        name="email"
        id="emailInput"
        ariaLabel="Email"
        autoComplete="email"
        labelText="Email"
        labelHtmlFor="emailInput"
        errorText={errorMessage}
        inputValidation={emailValidation}
        inputData={handleInputData}
      />
      <div className="flex items-center justify-center">
        <Button
          styles="mt-10 w-full font-semibold focus-visible:outline-1 hover:bg-hvr_primary
            text-white"
          type="button"
          onClick={handleClick}
          buttonText="Continue"
        />
      </div>
    </>
  );
}
