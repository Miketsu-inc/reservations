import { useState } from "react";
import Button from "../../components/Button";
import Input from "../../components/Input";

const defaultFormData = {
  company_name: "",
  owner: "",
  contact_email: "",
};

export default function MerchantInfoForm({ sendInputData, isCompleted }) {
  const [formData, setFormData] = useState(defaultFormData);
  const [isEmpty, setIsEmpty] = useState(false);

  function handleSubmit(e) {
    e.preventDefault();
    const form = e.target;
    if (!form.checkValidity()) {
      setIsEmpty(true);
      return;
    }

    sendInputData({
      company_name: formData.company_name,
      owner: formData.owner,
      contact_email: formData.contact_email,
    });
    isCompleted(true);
  }

  function handleInputData(data) {
    setFormData((prevAppData) => ({
      ...prevAppData,
      [data.name]: data.value,
    }));
  }
  return (
    <>
      <form
        noValidate
        className="flex w-full flex-col items-center justify-center *:w-full"
        onSubmit={handleSubmit}
      >
        <h1 className="text-center text-2xl font-semibold">
          Start signing up your company
        </h1>
        <p className="mt-4 text-center">
          Something about the data the user gives or idk
        </p>
        <Input
          type="text"
          styles=""
          placeholder="Global Serve kft"
          pattern=".{0,255}"
          name="company_name"
          id="company_name"
          errorText="Inputs must be 256 character or less!"
          labelText="Company Name"
          inputData={handleInputData}
          hasError={isEmpty}
        />
        <Input
          type="text"
          styles=""
          placeholder="Marcell Mikes"
          pattern=".{0,255}"
          name="owner"
          id="owner"
          autoComplete="name"
          errorText="Inputs must be 256 character or less!"
          labelText="Owner"
          inputData={handleInputData}
          hasError={isEmpty}
        />
        <Input
          type="email"
          styles=""
          placeholder="mycompany@gmail.com"
          pattern=".{0,254}@.*"
          name="contact_email"
          id="contact_email"
          autoComplete="email"
          errorText="Please eneter a valid email"
          labelText="Contact Email"
          inputData={handleInputData}
          hasError={isEmpty}
        />
        <Button
          onCLick={""}
          styles="p-2 w-5/6 mt-10 font-semibold focus-visible:outline-1 bg-primary
            hover:bg-hvr_primary text-white"
          name=""
          type="submit"
          buttonText="Continue"
        />
      </form>
    </>
  );
}
