import { Button, ServerError } from "@reservations/components";
import {
  isLocationComplete,
  LocationEditor,
} from "@reservations/jabulani/components";
import { useAuth } from "@reservations/jabulani/lib";
import { invalidateLocalStorageAuth } from "@reservations/lib";
import { useState } from "react";

export default function LocationPicker({
  isSubmitDone,
  isCompleted,
  redirect,
}) {
  const { merchantId } = useAuth();
  const [location, setLocation] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [serverError, setServerError] = useState("");

  async function submitHandler() {
    if (!isLocationComplete(location)) return;

    setIsLoading(true);
    setServerError("");

    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/locations`,
        {
          method: "POST",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify({
            country: location.country,
            city: location.city,
            postal_code: location.postal_code,
            address: location.address,
            geo_point: location.geo_point,
            place_id: location.place_id,
            formatted_location: location.formatted_location.trim(),
            is_primary: true,
            is_active: true,
          }),
        }
      );

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
        return;
      }

      isCompleted(true);
      isSubmitDone(true);
      redirect();
    } catch (error) {
      setServerError(error.message);
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div
      className="bg-layer_bg h-fit max-w-3xl overflow-hidden rounded-lg
        shadow-lg"
    >
      <div className="px-4 py-6">
        <ServerError error={serverError} />
        <h1 className="text-text_color mb-2 text-3xl font-bold">
          Where is your business located?
        </h1>
        <p className="text-gray-700 dark:text-gray-300">
          Enter the address customers should see, then place the pin at the
          exact entrance.
        </p>
      </div>
      <div className="flex flex-col gap-5 p-4">
        <LocationEditor value={location} onChange={setLocation} />
        <Button
          styles="w-full py-2"
          variant="primary"
          buttonText="Continue"
          disabled={!isLocationComplete(location)}
          isLoading={isLoading}
          onClick={submitHandler}
        />
      </div>
    </div>
  );
}
