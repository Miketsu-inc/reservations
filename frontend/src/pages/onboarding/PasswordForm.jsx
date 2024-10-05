import { useRef, useState } from "react";
import Button from "../../components/Button";
import Input from "../../components/Input";
import { MAX_PASSWORD_LENGTH, MIN_PASSWORD_LENGTH } from "../../lib/constants";

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
      <h2 className="mt-8 py-2 text-center text-xl sm:mt-4">Password</h2>
      <p className="py-2 text-center">
        Enter a password, which later you can use to login into your account.
        Please try to provide strong passwords
      </p>
      <Input
        styles=""
        ref={passwordRef}
        type="password"
        name="password"
        id="passwordInput"
        ariaLabel="Password"
        autoComplete="new-password"
        labelText="Password"
        labelHtmlFor="passwordInput"
        errorText={errorMessage.password}
        inputValidation={passwordValidation}
        inputData={handleInputData}
      />
      <Input
        styles="mt-4"
        ref={confirmPasswordRef}
        type="password"
        name="confirmPassword"
        id="confirmPasswordInput"
        ariaLabel="Confirm Password"
        autoComplete="new-password"
        labelText="Confirm Password"
        labelHtmlFor="confirmPasswordInput"
        errorText={errorMessage.confirmPassword}
        inputValidation={confirmPasswordValidation}
        inputData={handleInputData}
      />
      <div className="mt-4 flex items-center justify-center">
        <Button
          styles="mt-10 w-full font-semibold mt-4 focus-visible:outline-1 bg-primary
            hover:bg-hvr_primary text-white"
          type="button"
          onClick={handleClick}
          buttonText="Continue"
        ></Button>
      </div>
    </>
  );
}
