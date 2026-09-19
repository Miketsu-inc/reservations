import {
  Button,
  Input,
  ResponsiveDialog,
  ServerError,
} from "@reservations/components";
import { invalidateLocalStorageAuth } from "@reservations/lib";
import { useCallback, useEffect, useRef, useState } from "react";

export default function MerchantNameModal({ isOpen, onClose, onSubmit }) {
  const [newName, setNewName] = useState("");
  const [merchantUrl, setMerchantUrl] = useState({ valid: false, url: "" });
  const [serverError, setServerError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const keyUpTimer = useRef(null);
  const requestId = useRef(0);

  useEffect(() => {
    return () => clearTimeout(keyUpTimer.current);
  }, []);

  const checkMerchantUrl = useCallback(async (name) => {
    const currentRequestId = ++requestId.current;
    if (name !== "") {
      try {
        const response = await fetch("/api/v1/merchants/check-url", {
          method: "POST",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify({ merchant_name: name }),
        });

        const result = await response.json();
        if (currentRequestId !== requestId.current) return;

        if (response.ok) {
          setMerchantUrl({ valid: true, url: result.data.merchant_url });
        } else {
          invalidateLocalStorageAuth(response.status);
          setMerchantUrl({ valid: false, url: result.error.meta.merchant_url });
        }
      } catch (err) {
        if (currentRequestId === requestId.current) {
          setServerError(err.message);
        }
      }
    } else {
      setMerchantUrl({ valid: false, url: "" });
    }
  }, []);

  function handleInputData(data) {
    setNewName(data.value);

    if (keyUpTimer.current) {
      clearTimeout(keyUpTimer.current);
    }

    keyUpTimer.current = setTimeout(() => checkMerchantUrl(data.value), 600);
  }

  async function handleSubmit(e) {
    e.preventDefault();
    const form = e.target;

    if (!form.checkValidity()) {
      return;
    }
    if (!merchantUrl.valid || isSubmitting) return;

    setIsSubmitting(true);
    const didSave = await onSubmit(newName);
    setIsSubmitting(false);
    if (didSave) onClose();
  }

  return (
    <ResponsiveDialog styles="lg:max-w-xl" isOpen={isOpen} onClose={onClose}>
      <form className="m-4 flex flex-col gap-4" onSubmit={handleSubmit}>
        <h2 className="text-xl font-semibold">Change Merchant Name</h2>
        <p className="text-gray-700 dark:text-gray-300">
          Changing your merchant name will also change your booking page URL.
          Any customers with the old URL will no longer be able to access it.
        </p>

        <ServerError styles="mt-4 mb-2" error={serverError} />

        <Input
          type="text"
          labelText="New Merchant Name"
          name="merchant_name"
          value={newName}
          inputData={handleInputData}
          placeholder="My Company Ltd."
        />

        <p
          className={`${merchantUrl.url ? "" : "invisible"} text-sm
            dark:text-gray-400`}
        >
          <span
            className={`${merchantUrl.valid ? "text-text_color" : "text-red-600"}`}
          >
            {merchantUrl.valid
              ? `Your URL will be: https://miketsu.com/m/${merchantUrl.url}`
              : `The name '${merchantUrl.url}' is already taken.`}
          </span>
        </p>

        <div className="mt-1 flex justify-end gap-3">
          <Button
            variant="tertiary"
            styles="p-2 hidden lg:block"
            buttonText="Cancel"
            onClick={() => {
              setNewName("");
              setMerchantUrl({ valid: false, url: "" });
              onClose();
            }}
          />
          <Button
            variant="primary"
            type="submit"
            buttonText="Change Name"
            styles="p-2 w-full lg:w-auto"
            disabled={!merchantUrl.valid || isSubmitting}
            isLoading={isSubmitting}
          />
        </div>
      </form>
    </ResponsiveDialog>
  );
}
