import Button from "@components/Button";
import FloatingLabelInput from "@components/FloatingLabelInput";
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
    if (phoneNum.length > 12) {
      setErrorMessage("Inputs must be 12 characters or less!");
      return false;
    }

    if (phoneNum[0] !== "+") {
      setErrorMessage("Phone number should start with a '+' character!");
      return false;
    }

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
        ref={phoneNumRef}
        type="tel"
        name="phoneNum"
        id="PhoneNumInput"
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
            text-white py-2"
          type="button"
          onClick={handleClick}
          buttonText="Verify"
        />
      </div>
    </>
  );
}
