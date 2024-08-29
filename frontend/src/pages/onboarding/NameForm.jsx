import { useRef, useState } from "react";
import Button from "../../components/Button";
import Input from "../../components/Input";
import { MAX_INPUT_LENGTH } from "../../lib/constants";

const defaultNameData = {
  firstName: {
    value: "",
    isValid: false,
  },
  lastName: {
    value: "",
    isValid: false,
  },
};

export default function NameForm({ isCompleted, sendInputData }) {
  const firstNameRef = useRef();
  const lastNameRef = useRef();
  const [nameData, setNameData] = useState(defaultNameData);
  const [errorMessage, setErrorMessage] = useState("Please enter your name!");

  function firstNameValidation(firstName) {
    if (firstName.length > MAX_INPUT_LENGTH) {
      setErrorMessage(`Inputs must be ${MAX_INPUT_LENGTH} characters or less!`);
      return false;
    }
    if (firstName.length === 0) {
      return false;
    }
    return true;
  }

  function lastNameValidation(lastName) {
    if (lastName.length > MAX_INPUT_LENGTH) {
      setErrorMessage(`Inputs must be ${MAX_INPUT_LENGTH} characters or less!`);
      return false;
    }
    if (lastName.length === 0) {
      return false;
    }
    return true;
  }

  function handleInputData(data) {
    setNameData((prevNameData) => ({
      ...prevNameData,
      [data.name]: {
        value: data.value,
        isValid: data.isValid,
      },
    }));
  }

  function handleClick() {
    let hasError = false;

    if (!nameData.firstName.isValid) {
      firstNameRef.current.triggerValidationError();
      hasError = true;
    }

    if (!nameData.lastName.isValid) {
      lastNameRef.current.triggerValidationError();
      hasError = true;
    }

    if (!hasError) {
      sendInputData({
        firstName: nameData.firstName.value,
        lastName: nameData.lastName.value,
      });
      isCompleted(true);
    }
  }

  return (
    <>
      <h2 className="mt-8 py-2 text-2xl sm:mt-4">Enter your name</h2>
      <Input
        styles=""
        ref={firstNameRef}
        type="text"
        name="firstName"
        id="firstNameInput"
        ariaLabel="First Name"
        autoComplete="family-name"
        labelText="First Name"
        labelHtmlFor="firstNameInput"
        errorText={errorMessage}
        inputValidation={firstNameValidation}
        inputData={handleInputData}
      />
      <Input
        styles=""
        ref={lastNameRef}
        type="text"
        name="lastName"
        id="lastNameInput"
        ariaLabel="last name"
        autoComplete="given-name"
        labelText="Last Name"
        labelHtmlFor="lastNameInput"
        errorText={errorMessage}
        inputValidation={lastNameValidation}
        inputData={handleInputData}
      />
      <div className="flex items-center justify-center">
        <Button
          styles="mt-10 w-full font-semibold focus-visible:outline-1 bg-primary
            hover:bg-hvr_primary text-white"
          type="button"
          onClick={handleClick}
          buttonText="Continue"
        />
      </div>
    </>
  );
}
