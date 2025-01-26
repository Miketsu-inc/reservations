import Button from "@components/Button";
import Input from "@components/Input";
import ServerError from "@components/ServerError";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { useCallback, useState } from "react";

const defaultFormData = {
  name: "",
  contact_email: "",
};

const defaultMerchantUrl = {
  valid: false,
  url: "",
};

var keyUpTimer;

export default function MerchantInfoForm({ isCompleted }) {
  const [formData, setFormData] = useState(defaultFormData);
  const [isEmpty, setIsEmpty] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [serverError, setServerError] = useState("");
  const [merchantUrl, setMerchantUrl] = useState(defaultMerchantUrl);

  async function handleSubmit(e) {
    e.preventDefault();
    const form = e.target;

    if (!form.checkValidity()) {
      setIsEmpty(true);
      return;
    }

    setIsLoading(true);
    try {
      const response = await fetch("/api/v1/auth/merchant/signup", {
        method: "POST",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify(formData),
      });

      if (!response.ok) {
        invalidateLocalSotrageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        setServerError("");
        isCompleted(true);
      }
    } catch (err) {
      setServerError(err.message);
    } finally {
      setIsLoading(false);
    }
  }

  const checkMerchantUrl = useCallback(async (merchantName) => {
    if (merchantName !== "") {
      try {
        const response = await fetch("/api/v1/merchants/check-url", {
          method: "POST",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify({
            merchant_name: merchantName,
          }),
        });

        const result = await response.json();
        if (response.ok) {
          setMerchantUrl({
            valid: true,
            url: result.data.merchant_url,
          });
        } else {
          invalidateLocalSotrageAuth(response.status);
          setMerchantUrl({
            valid: false,
            url: result.error.merchant_url,
          });
        }
      } catch (err) {
        setServerError(err.message);
      }
    } else {
      setMerchantUrl({
        valid: false,
        url: "",
      });
    }
  }, []);

  function handleInputData(data) {
    setFormData((prevAppData) => ({
      ...prevAppData,
      [data.name]: data.value,
    }));

    if (data.name === "name" && formData.name !== data.value) {
      if (keyUpTimer) {
        clearTimeout(keyUpTimer);
      }

      keyUpTimer = setTimeout(() => checkMerchantUrl(data.value), 600);
    }
  }
  return (
    <>
      <form
        noValidate
        className="flex w-full flex-col items-center justify-center gap-4 *:w-full"
        onSubmit={handleSubmit}
      >
        <h1 className="mb-8 text-center text-xl font-semibold sm:mb-4">
          Start signing up your company
        </h1>
        <p className="mt-4 text-center">
          Something about the data the user gives or idk
        </p>
        <ServerError styles="mt-4 mb-2" error={serverError} />
        <Input
          type="text"
          styles="p-2"
          placeholder="Global Serve kft"
          pattern=".{0,255}"
          name="name"
          id="company_name"
          errorText="Inputs must be 256 character or less!"
          labelText="Company Name"
          inputData={handleInputData}
          hasError={isEmpty}
        />
        <p
          className={`${merchantUrl.url ? "" : "invisible"} text-sm dark:text-gray-400`}
        >
          <span
            className={`${merchantUrl.valid ? "text-text_color" : "text-red-600"}`}
          >
            {merchantUrl.valid
              ? `Your URL will be: https://miketsu.com/m/${merchantUrl.url}`
              : `The name '${merchantUrl.url}' is already taken.`}
          </span>
        </p>
        <Input
          type="email"
          styles="p-2 mt-2"
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
          variant="primary"
          styles="py-2 sm:mt-10 mt-14"
          type="submit"
          buttonText="Continue"
          isLoading={isLoading}
        />
      </form>
    </>
  );
}
