import Input from "../../components/Input";

export default function PersonalInfo({
  firstNameRef,
  lastNameRef,
  handleInputData,
}) {
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
        ref={firstNameRef}
        type="text"
        name="firstName"
        id="firstNameInput"
        ariaLabel="First Name"
        autoComplete="family-name"
        labelText="First Name"
        labelHtmlFor="firstNameInput"
        errorText="Please enter your first name"
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
        errorText="Please enter your last name"
        inputValidation={lastNameValidation}
        inputData={handleInputData}
      />
    </>
  );
}
