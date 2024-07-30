import Input from "../../components/Input";

export default function PasswordPage(props) {
  function passwordValidation(password) {
    return password.length > 6;
  }
  function confirmPasswordValidation(confirmPassword) {
    return confirmPassword;
  }

  //Validation when submiting

  return (
    <>
      <Input
        styles=""
        ref={props.passwordRef}
        type="password"
        name="password"
        id="passwordInput"
        ariaLabel="Password"
        autoComplete="new-password"
        labelText="Password"
        labelHtmlFor="passwordInput"
        errorText="Please enter a valid password!"
        inputValidation={passwordValidation}
        inputData={props.handleInputData}
      />
      <Input
        styles=""
        ref={props.confirmPasswordRef}
        type="password"
        name="confirmPassword"
        id="confirmPasswordInput"
        ariaLabel="Confirm Password"
        autoComplete="new-password"
        labelText="Confirm Password"
        labelHtmlFor="confirmPasswordInput"
        errorText="The two password should match"
        inputValidation={confirmPasswordValidation}
        inputData={props.handleInputData}
      />
    </>
  );
}
