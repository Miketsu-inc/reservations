import Button from "@components/Button";
import FloatingLabelInput from "@components/FloatingLabelInput";
import { useRef, useState } from "react";

const defaultPhoneNumData = {
  phone_number: {
    value: "",
    isValid: false,
  },
};

export default function PhoneNumberForm({ isCompleted, sendInputData }) {
  const phoneNumRef = useRef();
  const [phoneNumData, setPhoneNumData] = useState(defaultPhoneNumData);
  const [errorMessage, setErrorMessage] = useState(
    "Please enter your phone number!"
  );

  function PhoneNumValidation(phone_number) {
    if (phone_number.length > 12) {
      setErrorMessage("Inputs must be 12 characters or less!");
      return false;
    }

    if (phone_number[0] !== "+") {
      setErrorMessage("Phone number should start with a '+' character!");
      return false;
    }

    return phone_number;
  }

  function handleInputData(data) {
    setPhoneNumData((prevPhoneNumData) => ({
      ...prevPhoneNumData,
      [data.name]: {
        value: data.value,
        isValid: data.isValid,
      },
    }));
  }

  function handleClick() {
    if (!phoneNumData.phone_number.isValid) {
      phoneNumRef.current.triggerValidationError();
    } else {
      sendInputData({
        phone_number: phoneNumData.phone_number.value,
      });
      isCompleted(true);
    }
  }

  return (
    <>
      <h1 className="mt-2 text-center text-xl font-semibold">Phone number</h1>
      <p className="mb-8 py-2 text-center text-gray-600 dark:text-gray-300">
        Enter your phone number to receive notifications about your appointments
      </p>
      <FloatingLabelInput
        ref={phoneNumRef}
        type="tel"
        name="phone_number"
        id="PhoneNumInput"
        autoComplete="tel"
        labelText="Enter your phone number"
        errorText={errorMessage}
        inputValidation={PhoneNumValidation}
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
