import Button from "@components/Button";
import Input from "@components/Input";
import ServerError from "@components/ServerError";
import { invalidateLocalStorageAuth } from "@lib/lib";
import { useState } from "react";

const defaultFormData = {
  country: "",
  postal_code: "",
  city: "",
  address: "",
};
export default function LocationForm({ isSubmitDone, isCompleted, redirect }) {
  const [formData, setFormData] = useState(defaultFormData);
  const [isLoading, setIsLoading] = useState(false);
  const [serverError, setServerError] = useState("");

  async function handleSubmit(e) {
    e.preventDefault();
    const form = e.target;

    if (!form.checkValidity()) {
      return;
    }

    setIsLoading(true);
    try {
      const response = await fetch("/api/v1/merchants/location", {
        method: "POST",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify(formData),
      });

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        redirect();
        setServerError("");
        isCompleted(true);
        isSubmitDone(true);
      }
    } catch (err) {
      setServerError(err.message);
    } finally {
      setIsLoading(false);
    }
  }

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
        className="flex w-full flex-col items-center justify-center gap-4 *:w-full"
        onSubmit={handleSubmit}
      >
        <h1 className="mb-4 text-center text-xl font-semibold sm:mb-6">
          Location
        </h1>
        <ServerError styles="mt-4 mb-2" error={serverError} />
        <div className="flex w-full gap-4">
          <div className="grow">
            <Input
              type="text"
              styles="w-full p-2"
              placeholder="Hungary"
              pattern=".{0,255}"
              name="country"
              id="country"
              autoComplete="country"
              labelText="Country"
              inputData={handleInputData}
            />
          </div>
          <div className="w-24">
            <Input
              type="text"
              styles="w-full p-2"
              placeholder="2119"
              pattern="^[0-9]{0,255}$"
              name="postal_code"
              id="postal_code"
              autoComplete="postal-code"
              labelText="Postal Code"
              inputData={handleInputData}
            />
          </div>
        </div>
        <div className="grid grid-cols-1 gap-4 sm:mt-6 sm:grid-cols-2">
          <Input
            type="text"
            styles="w-full p-2"
            placeholder="Budapest"
            pattern=".{0,255}"
            name="city"
            id="city"
            autoComplete="address-level2"
            labelText="City"
            inputData={handleInputData}
          />
          <Input
            type="text"
            styles="w-full p-2"
            placeholder="Állás utca 3"
            pattern=".{0,255}"
            name="address"
            id="address"
            autoComplete="address-line1"
            labelText="Address"
            inputData={handleInputData}
          />
        </div>
        <Button
          variant="primary"
          styles="p-2 sm:mt-10 mt-8"
          type="submit"
          buttonText="Continue"
          isLoading={isLoading}
        />
      </form>
    </>
  );
}
