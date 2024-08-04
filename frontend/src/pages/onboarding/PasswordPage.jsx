import Input from "../../components/Input";
import { MIN_PASSWORD_LENGTH } from "../../lib/constants";

export default function PasswordPage({
  passwordRef,
  confirmPasswordRef,
  handleInputData,
}) {
  function passwordValidation(password) {
    return password.length > MIN_PASSWORD_LENGTH;
  }
  function confirmPasswordValidation(confirmPassword) {
    return confirmPassword;
  }

  //Validation when submiting

  return (
    <>
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
        errorText="Please enter a valid password!"
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
        errorText="The two password should match"
        inputValidation={confirmPasswordValidation}
        inputData={handleInputData}
      />
    </>
  );
}
