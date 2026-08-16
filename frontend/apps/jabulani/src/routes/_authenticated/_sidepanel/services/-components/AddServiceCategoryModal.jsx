import { Button, Input, ResponsiveDialog } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { useState } from "react";

export default function AddServiceCategoryModal({ isOpen, onClose, onAdded }) {
  const [categoryData, setCategoryData] = useState({ name: "" });
  const { showToast } = useToast();
  const { merchantId } = useAuth();

  function updateCategoryData(data) {
    setCategoryData((prev) => ({ ...prev, ...data }));
  }

  async function submitHandler(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    const response = await fetch(
      `/api/v1/merchants/${merchantId}/service-categories`,
      {
        method: "POST",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          name: categoryData.name,
        }),
      }
    );

    if (!response.ok) {
      const result = await response.json();
      invalidateLocalStorageAuth(response.status);
      showToast({
        variant: "error",
        message: `Something went wrong while creating a new category ${result.error}`,
      });
    } else {
      showToast({
        variant: "success",
        message: `New service category added successfully`,
      });

      onAdded();
      onClose();
      setCategoryData({ name: "" });
    }
  }

  return (
    <ResponsiveDialog isOpen={isOpen} onClose={onClose}>
      <form className="p-4 md:w-lg" onSubmit={submitHandler}>
        <p className="pb-8 text-xl font-semibold">Create a new category</p>
        <div className="flex flex-col gap-6">
          <p>
            Order your services by categorizing them. The services will get
            displayed under their categories.
          </p>
          <Input
            id="CategoryName"
            name="CategoryName"
            type="text"
            labelText="Category name"
            placeholder="e.g. hair"
            value={categoryData.name}
            inputData={(data) => updateCategoryData({ name: data.value })}
          />
          <div className="flex items-center justify-end gap-2">
            <Button
              styles="py-2 px-4 hidden lg:block"
              buttonText="Cancel"
              variant="tertiary"
              type="button"
              onClick={onClose}
            />
            <Button
              styles="py-2 px-4 w-full lg:w-auto"
              buttonText="Create"
              variant="primary"
              type="submit"
            />
          </div>
        </div>
      </form>
    </ResponsiveDialog>
  );
}
