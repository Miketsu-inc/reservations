import Input from "../../components/Input";

export default function PersonalInfo(props) {
  function firstNameValidation(firstName) {
    return firstName;
  }
  function lastNameValidation(lastName) {
    return lastName;
  }

  return (
    <>
      <Input
        styles=""
        type="text"
        name="firstName"
        id="firstNameInput"
        ariaLabel="First Name"
        autoComplete="family-name"
        labelText="First Name"
        labelHtmlFor="firstNameInput"
        errorText=""
        inputValidation={firstNameValidation}
        inputData={props.handleInputData}
      />
      <Input
        styles=""
        type="text"
        name="lastName"
        id="lastNameInput"
        ariaLabel="last name"
        autoComplete="given-name"
        labelText="Last Name"
        labelHtmlFor="lastNameInput"
        errorText=""
        inputValidation={lastNameValidation}
        inputData={props.handleInputData}
      />
    </>
  );
}
