import { useEffect, useState } from "react";
import Button from "../../components/Button";
import Input from "../../components/Input";
import ServerError from "../../components/ServerError";

const defaultFormData = {
  name: "",
  contact_email: "",
};

export default function MerchantInfoForm({ isCompleted }) {
  const [formData, setFormData] = useState(defaultFormData);
  const [isEmpty, setIsEmpty] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [serverError, setServerError] = useState(undefined);

  function handleSubmit(e) {
    e.preventDefault();
    const form = e.target;
    if (!form.checkValidity()) {
      setIsEmpty(true);
      return;
    }
    setIsSubmitting(true);
  }

  useEffect(() => {
    if (isSubmitting) {
      const sendRequest = async () => {
        try {
          const response = await fetch("/api/v1/auth/merchant/signup", {
            method: "POST",
            headers: {
              Accept: "application/json",
              "content-type": "application/json",
            },
            body: JSON.stringify(formData),
          });
          const result = await response.json();
          if (result.error) {
            setServerError(result.error);
            return;
          } else {
            setServerError(undefined);
            isCompleted(true);
          }
        } catch (err) {
          setServerError("An error occurred. Please try again.");
        } finally {
          setIsSubmitting(false);
        }
      };
      sendRequest();
    }
  }, [isSubmitting, formData, isCompleted]);

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
        <ServerError styles="mt-4 mb-2" error={serverError} />
        <Input
          type="text"
          styles=""
          placeholder="Global Serve kft"
          pattern=".{0,255}"
          name="name"
          id="company_name"
          errorText="Inputs must be 256 character or less!"
          labelText="Company Name"
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
          styles="p-2 w-5/6 mt-10 font-semibold focus-visible:outline-1 bg-primary
            hover:bg-hvr_primary text-white"
          type="submit"
          buttonText="Continue"
          isLoading={isSubmitting}
        />
      </form>
    </>
  );
}
