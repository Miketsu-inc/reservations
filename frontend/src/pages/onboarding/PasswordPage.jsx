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
        type="password"
        name="confirmPassword"
        id="confirmPasswordInput"
        ariaLabel="Confirm Password"
        autoComplete="new-password"
        labelText="Confirm Password"
        labelHtmlFor="confirmPasswordInput"
        errorText=""
        inputValidation={confirmPasswordValidation}
        inputData={props.handleInputData}
      />
    </>
  );
}
