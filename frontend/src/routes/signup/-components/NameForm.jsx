import Button from "@components/Button";
import FloatingLabelInput from "@components/FloatingLabelInput";
import { MAX_INPUT_LENGTH } from "@lib/constants";
import { useRef, useState } from "react";

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

const defaultErrorMeassage = {
  firstname: "Please enter your first name",
  lastname: "Please enter your last name",
};

export default function NameForm({ isCompleted, isLoading, SubmitForm }) {
  const firstNameRef = useRef();
  const lastNameRef = useRef();
  const [nameData, setNameData] = useState(defaultNameData);
  const [errorMessage, setErrorMessage] = useState(defaultErrorMeassage);

  function updateErrors(key, message) {
    setErrorMessage((prevErrorMessage) => ({
      ...prevErrorMessage,
      [key]: message,
    }));
  }

  function firstNameValidation(firstName) {
    if (firstName.length > MAX_INPUT_LENGTH) {
      updateErrors(
        "firstname",
        `Inputs must be ${MAX_INPUT_LENGTH} characters or less!`
      );
      return false;
    }
    return true;
  }

  function lastNameValidation(lastName) {
    if (lastName.length > MAX_INPUT_LENGTH) {
      updateErrors(
        "lastname",
        `Inputs must be ${MAX_INPUT_LENGTH} characters or less!`
      );
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

    if (!nameData.firstName.isValid || nameData.firstName.value.length === 0) {
      firstNameRef.current.triggerValidationError();
      hasError = true;
    }

    if (!nameData.lastName.isValid || nameData.lastName.value.length === 0) {
      lastNameRef.current.triggerValidationError();
      hasError = true;
    }

    if (!hasError) {
      SubmitForm(nameData.firstName.value, nameData.lastName.value);
      isCompleted(true);
    }
  }

  return (
    <>
      <h2 className="mt-8 py-2 text-center text-xl sm:mt-4">Username</h2>
      <p className="py-2 text-center">
        Enter your first and last name, which you will use over the application.
      </p>
      <FloatingLabelInput
        ref={firstNameRef}
        type="text"
        name="firstName"
        id="firstNameInput"
        autoComplete="family-name"
        labelText="First Name"
        errorText={errorMessage.firstname}
        inputValidation={firstNameValidation}
        inputData={handleInputData}
      />
      <FloatingLabelInput
        styles="mt-4"
        ref={lastNameRef}
        type="text"
        name="lastName"
        id="lastNameInput"
        autoComplete="given-name"
        labelText="Last Name"
        errorText={errorMessage.lastname}
        inputValidation={lastNameValidation}
        inputData={handleInputData}
      />
      <div className="mt-4 flex items-center justify-center">
        <Button
          styles="mt-6 w-full font-semibold mt-4 focus-visible:outline-1 bg-primary
            hover:bg-hvr_primary text-white py-2"
          type="button"
          onClick={handleClick}
          isLoading={isLoading}
          buttonText="Finish"
        ></Button>
      </div>
    </>
  );
}
