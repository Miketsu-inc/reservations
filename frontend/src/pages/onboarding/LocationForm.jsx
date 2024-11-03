import { useEffect, useState } from "react";
import Button from "../../components/Button";
import Input from "../../components/Input";

const defaultFormData = {
  country: "",
  postal_code: "",
  city: "",
  address: "",
};
export default function LocationForm({ isSubmitDone, isCompleted }) {
  const [formData, setFormData] = useState(defaultFormData);
  const [isEmpty, setIsEmpty] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

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
          const response = await fetch("/api/v1/merchants/location", {
            method: "POST",
            headers: {
              Accept: "application/json",
              "content-type": "application/json",
            },
            body: JSON.stringify(formData),
          });
          const result = await response.json();
          if (result.error) {
            console.log(result.error);
            return;
          } else {
            isCompleted(true);
            isSubmitDone(true);
          }
        } catch (err) {
          console.error("Error messsage from server:", err.message);
        } finally {
          setIsSubmitting(false);
        }
      };
      sendRequest();
    }
  }, [isSubmitting, formData, isCompleted, isSubmitDone]);

  function handleInputData(data) {
    setFormData((prevFormData) => ({
      ...prevFormData,
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
        <h1 className="text-center text-2xl font-bold">Location</h1>
        <p className="text-center">
          Give us the location of your working space
        </p>
        <div className="flex w-full gap-4">
          <div className="flex-grow">
            <Input
              type="text"
              styles="w-full"
              placeholder="Hungary"
              pattern=".{0,255}"
              name="country"
              id="country"
              autoComplete="country"
              errorText="Inputs must be 256 character or less!"
              labelText="Country"
              inputData={handleInputData}
              hasError={isEmpty}
            />
          </div>
          <div className="w-24">
            <Input
              type="text"
              styles="w-fulls"
              placeholder="2119"
              pattern="^[0-9]{0,255}$"
              name="postal_code"
              id="postal_code"
              autoComplete="postal-code"
              errorText="Postal code should consists of numbers only!"
              labelText="Postal Code"
              inputData={handleInputData}
              hasError={isEmpty}
            />
          </div>
        </div>
        <div className="grid grid-cols-1 gap-4 sm:mt-2 sm:grid-cols-2">
          <Input
            type="text"
            styles="w-full"
            placeholder="Budapest"
            pattern=".{0,255}"
            name="city"
            id="city"
            autoComplete="address-level2"
            errorText="Inputs must be 256 character or less!"
            labelText="City"
            inputData={handleInputData}
            hasError={isEmpty}
          />
          <Input
            type="text"
            styles="w-full"
            placeholder="Állás utca 3"
            pattern=".{0,255}"
            name="address"
            id="address"
            autoComplete="address-line1"
            errorText="Inputs must be 256 character or less!"
            labelText="Address"
            inputData={handleInputData}
            hasError={isEmpty}
          />
        </div>
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
