import Button from "@components/Button";
import FloatingLabelInput from "@components/FloatingLabelInput";
import { MAX_INPUT_LENGTH } from "@lib/constants";
import { useRef, useState } from "react";

const defaultPhoneNumData = {
  phoneNum: {
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

  function PhoneNumValidation(phoneNum) {
    if (phoneNum.length > MAX_INPUT_LENGTH) {
      setErrorMessage(`Inputs must be ${MAX_INPUT_LENGTH} characters or less!`);
      return false;
    }
    //more validation?
    return phoneNum;
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
    if (!phoneNumData.phoneNum.isValid) {
      phoneNumRef.current.triggerValidationError();
    } else {
      sendInputData({
        phoneNum: phoneNumData.phoneNum.value,
      });
      isCompleted(true);
    }
  }

  return (
    <>
      <h1 className="mt-4 text-center text-xl font-semibold">Phone number</h1>
      <p className="py-2 text-center">
        Enter your phone number to receive reminder notifications about your
        appointment
      </p>
      <FloatingLabelInput
        styles=""
        ref={phoneNumRef}
        type="tel"
        name="phoneNum"
        id="PhoneNumInput"
        ariaLabel="phone number"
        autoComplete="tel"
        labelText="Enter your phone number"
        errorText={errorMessage}
        inputValidation={PhoneNumValidation}
        inputData={handleInputData}
      />

      <p className="mt-4 px-2 pt-3 text-center text-sm tracking-tight">
        After giving your number, you'll receive a{" "}
        <span className="text-nowrap underline"> verification code</span>.
        Please check your messages and verify your phone number.
      </p>
      <div className="flex items-center justify-center">
        <Button
          styles="mt-6 w-full font-semibold focus-visible:outline-1 hover:bg-hvr_primary
            text-white"
          type="button"
          onClick={handleClick}
          buttonText="Verify"
        />
      </div>
    </>
  );
}
