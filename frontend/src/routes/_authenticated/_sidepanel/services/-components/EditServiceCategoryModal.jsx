import Button from "@components/Button";
import Input from "@components/Input";
import Modal from "@components/Modal";
import { useToast } from "@lib/hooks";
import { invalidateLocalStorageAuth } from "@lib/lib";
import { useState } from "react";

export default function EditServiceCategoryModal({
  category,
  isOpen,
  onClose,
  onModified,
}) {
  const [categoryData, setCategoryData] = useState({ name: category.name });
  const { showToast } = useToast();

  function updateCategoryData(data) {
    setCategoryData((prev) => ({ ...prev, ...data }));
  }

  async function submitHandler(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    const response = await fetch(
      `/api/v1/merchants/services/categories/${category.id}`,
      {
        method: "PUT",
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

      onModified();
      onClose();
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <form className="p-4 sm:w-lg" onSubmit={submitHandler}>
        <p className="pb-8 text-xl font-semibold">Edit category</p>
        <div className="flex flex-col gap-6">
          <div className="size-18 shrink-0 overflow-hidden rounded-lg">
            <img
              className="size-full object-cover"
              src="https://dummyimage.com/120x120/d156c3/000000.jpg"
              alt="service photo"
            ></img>
          </div>
          <Input
            styles="p-2"
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
              styles="py-2 px-4"
              buttonText="Cancel"
              variant="tertiary"
              type="button"
              onClick={onClose}
            />
            <Button
              styles="py-2 px-4"
              buttonText="Save"
              variant="primary"
              type="submit"
            />
          </div>
        </div>
      </form>
    </Modal>
  );
}
