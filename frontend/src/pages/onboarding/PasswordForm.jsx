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

export default function PasswordForm({
  isCompleted,
  sendInputData,
  isSubmitting,
}) {
  const passwordRef = useRef();
  const confirmPasswordRef = useRef();
  const [passwordData, setPasswordData] = useState(defaultPasswordData);
  const [errorMessage, setErrorMessage] = useState(
    "Please enter your password!"
  );

  function passwordValidation(password) {
    if (password.length < MIN_PASSWORD_LENGTH) {
      setErrorMessage(
        `Password must be ${MIN_PASSWORD_LENGTH} characters or more!`
      );
      return false;
    }
    if (password.length > MAX_PASSWORD_LENGTH) {
      setErrorMessage(
        `Password must be ${MAX_PASSWORD_LENGTH} characters or less!`
      );
      return false;
    }
    return true;
  }

  function confirmPasswordValidation(confirmPassword) {
    if (confirmPassword.length > MAX_PASSWORD_LENGTH) {
      setErrorMessage(
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

  function handleClick(e) {
    let hasError = false;

    if (!passwordData.password.isValid) {
      passwordRef.current.triggerValidationError();
      hasError = true;
    }

    if (!passwordData.confirmPassword.isValid) {
      confirmPasswordRef.current.triggerValidationError();
      hasError = true;
    }

    if (passwordData.password.value !== passwordData.confirmPassword.value) {
      confirmPasswordRef.current.triggerValidationError();
      hasError = true;
    }

    if (!hasError) {
      sendInputData({
        password: passwordData.password.value,
        //confirmPassword: passwordData.confirmPassword.value,
      });
      e.target.form.requestSubmit();
      isCompleted(true);
    }
  }

  return (
    <>
      <h2 className="mt-8 py-2 text-2xl sm:mt-4">Enter your password</h2>
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
        errorText={errorMessage}
        inputValidation={passwordValidation}
        inputData={handleInputData}
      />
      <Input
        styles=""
        ref={confirmPasswordRef}
        type="password"
        name="confirmPassword"
        id="confirmPasswordInput"
        ariaLabel="Confirm Password"
        autoComplete="new-password"
        labelText="Confirm Password"
        labelHtmlFor="confirmPasswordInput"
        errorText={errorMessage}
        inputValidation={confirmPasswordValidation}
        inputData={handleInputData}
      />
      <div className="flex items-center justify-center">
        <Button
          styles="mt-10 w-full font-semibold mt-4 focus-visible:outline-1 bg-primary
            hover:bg-hvr-primary text-white"
          type="button"
          onClick={handleClick}
          isLoading={isSubmitting}
          buttonText="Finish"
        ></Button>
      </div>
    </>
  );
}
