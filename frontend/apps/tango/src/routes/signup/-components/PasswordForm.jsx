import { Button, FloatingLabelInput } from "@reservations/components";
import { MAX_PASSWORD_LENGTH, MIN_PASSWORD_LENGTH } from "@reservations/lib";
import { useRef, useState } from "react";

const defaultPasswordData = {
  password: {
    value: "",
    isValid: false,
  },
  confirmPassword: {
    value: "",
    isValid: false,
  },
};

const defaultErrorMeassage = {
  password: "Please enter your password!",
  confirmPassword: "Please enter your password again!",
};

export default function PasswordForm({ isCompleted, sendInputData }) {
  const passwordRef = useRef();
  const confirmPasswordRef = useRef();
  const [passwordData, setPasswordData] = useState(defaultPasswordData);
  const [errorMessage, setErrorMessage] = useState(defaultErrorMeassage);

  function updateErrors(key, message) {
    setErrorMessage((prevErrorMessage) => ({
      ...prevErrorMessage,
      [key]: message,
    }));
  }

  function passwordValidation(password) {
    if (password.length < MIN_PASSWORD_LENGTH) {
      updateErrors(
        "password",
        `Password must be ${MIN_PASSWORD_LENGTH} characters or more!`
      );
      return false;
    }
    if (password.length > MAX_PASSWORD_LENGTH) {
      updateErrors(
        "password",
        `Password must be ${MAX_PASSWORD_LENGTH} characters or less!`
      );
      return false;
    }
    return true;
  }

  function confirmPasswordValidation(confirmPassword) {
    if (confirmPassword.length > MAX_PASSWORD_LENGTH) {
      updateErrors(
        "confirmPassword",
        `Password must be ${MAX_PASSWORD_LENGTH} characters or less!`
      );
      return false;
    }
    return true;
  }

  function handleInputData(data) {
    setPasswordData((prevPasswordData) => ({
      ...prevPasswordData,
      [data.name]: {
        value: data.value,
        isValid: data.isValid,
      },
    }));
  }

  function handleClick() {
    let hasError = false;

    if (!passwordData.password.isValid) {
      passwordRef.current.triggerValidationError();
      hasError = true;
    }

    if (
      !passwordData.confirmPassword.isValid ||
      passwordData.confirmPassword.value.length === 0
    ) {
      confirmPasswordRef.current.triggerValidationError();
      hasError = true;
    }

    if (passwordData.password.value !== passwordData.confirmPassword.value) {
      confirmPasswordRef.current.triggerValidationError();
      updateErrors("confirmPassword", "The two passwords should match!");
      hasError = true;
    }

    if (!hasError) {
      sendInputData({
        password: passwordData.password.value,
      });
      isCompleted(true);
    }
  }

  return (
    <>
      <h2 className="mb-8 text-center text-xl font-semibold">Password</h2>
      <FloatingLabelInput
        ref={passwordRef}
        type="password"
        name="password"
        id="passwordInput"
        autoComplete="new-password"
        labelText="Password"
        errorText={errorMessage.password}
        inputValidation={passwordValidation}
        inputData={handleInputData}
      />
      <FloatingLabelInput
        styles="mt-4"
        ref={confirmPasswordRef}
        type="password"
        name="confirmPassword"
        id="confirmPasswordInput"
        autoComplete="new-password"
        labelText="Confirm Password"
        errorText={errorMessage.confirmPassword}
        inputValidation={confirmPasswordValidation}
        inputData={handleInputData}
      />
      <div className="mt-10 flex items-center justify-center">
        <Button
          variant="primary"
          styles="w-full py-2"
          type="button"
          onClick={handleClick}
          buttonText="Continue"
        ></Button>
      </div>
    </>
  );
}
