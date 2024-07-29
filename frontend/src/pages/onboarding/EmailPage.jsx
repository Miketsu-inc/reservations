import Input from "../../components/Input";

export default function EmailPage(props) {
  function emailValidation(email) {
    return email.includes("@");
  }
  return (
    <Input
      styles=""
      type="text"
      name="email"
      id="emailInput"
      ariaLabel="Email"
      autoComplete="email"
      labelText="Email"
      labelHtmlFor="emailInput"
      errorText="Please enter a valid email!"
      inputValidation={emailValidation}
      inputData={props.handleInputData}
    />
  );
}
