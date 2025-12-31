import {
  Button,
  CloseButton,
  Input,
  Modal,
  Select,
} from "@reservations/components";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { useState } from "react";

const commonEmojis = [
  "⏰",
  "📚",
  "📅",
  "🍱",
  "☕",
  "💼",
  "🏃",
  "🎯",
  "💡",
  "🔧",
  "📝",
  "🎨",
];

const defaultFormData = {
  id: null,
  icon: "⏰",
  name: "",
  duration: "",
  duration_unit: "min",
};

export default function BlockedTypesModal({
  isOpen,
  onClose,
  onSubmit,
  editData,
}) {
  const { showToast } = useToast();
  const [isSelectOpen, setIsSelectOpen] = useState(false);
  const [formData, setFormData] = useState({
    id: editData?.id || null,
    icon: editData?.icon || "⏰",
    name: editData?.name || "",
    duration: editData?.duration || "",
    duration_unit: editData?.duration_unit || "min",
  });

  function updateFormData(data) {
    setFormData((prev) => ({ ...prev, ...data }));
  }

  async function handleSubmit(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    const durationInMinutes =
      formData.duration_unit === "hour"
        ? formData.duration * 60
        : formData.duration;

    const body = {
      id: formData.id,
      name: formData.name,
      icon: formData.icon,
      duration: Number(durationInMinutes),
    };

    let url = "";
    let method = "";

    if (formData.id != null) {
      url = `/api/v1/merchants/blocked-time-types/${formData.id}`;
      method = "PUT";
    } else {
      // for correct json sending
      delete body.id;
      url = "/api/v1/merchants/blocked-time-types";
      method = "POST";
    }

    try {
      const response = await fetch(url, {
        method: method,
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        showToast({ message: result.error.message, variant: "error" });
      } else {
        showToast({
          message:
            method === "POST"
              ? "Blocked time type added successfully"
              : "Blocked time type modified successfully",
          variant: "success",
        });
        onSubmit();
        setFormData(defaultFormData);
        onClose();
      }
    } catch (err) {
      showToast({ message: err.message, variant: "error" });
    }
  }

  return (
    <Modal
      styles="max-w-md"
      isOpen={isOpen}
      onClose={() => {
        onClose();
        setFormData(defaultFormData);
      }}
      disableFocusTrap={true}
      suspendCloseOnClickOutside={isSelectOpen}
    >
      <form
        className="flex flex-col gap-4 p-6"
        id="BlockedTimeTypeForm"
        onSubmit={handleSubmit}
      >
        <div className="flex items-center justify-between">
          <div className="text-text_color text-lg font-semibold">
            {editData ? "Edit Blocked Time Type" : "New Blocked Time Type"}
          </div>
          <CloseButton
            onClick={() => {
              onClose();
              setFormData(defaultFormData);
            }}
          />
        </div>

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <label className="text-text_color font-medium">Icon</label>
            <div className="flex flex-wrap justify-center gap-2">
              {commonEmojis.map((icon) => (
                <button
                  key={icon}
                  type="button"
                  onClick={() => updateFormData({ icon })}
                  className={`size-13 rounded-md border-2 text-xl transition-all
                  ${
                    formData.icon === icon
                      ? "border-primary"
                      : "border-border_color hover:border-gray-400"
                  }`}
                >
                  {icon}
                </button>
              ))}
            </div>
          </div>

          <Input
            styles="p-2"
            id="name"
            name="name"
            type="text"
            labelText="Name"
            placeholder="e.g., Coffee Break"
            value={formData.name}
            inputData={(data) => updateFormData({ name: data.value })}
          />

          <div className="flex w-full flex-row items-end gap-2">
            <Input
              styles="p-2"
              id="duration"
              name="duration"
              type="number"
              min={1}
              max={formData.duration_unit === "hour" ? 24 : 1440}
              labelText="Duration"
              placeholder="30"
              value={formData.duration}
              inputData={(data) =>
                updateFormData({ duration: Number(data.value) })
              }
            >
              <Select
                styles="w-32! rounded-l-none"
                value={formData.duration_unit || "min"}
                options={[
                  { value: "min", label: "minutes" },
                  { value: "hour", label: "hours" },
                ]}
                onSelect={(option) =>
                  updateFormData({ duration_unit: option.value })
                }
                onOpenChange={(open) => setIsSelectOpen(open)}
              />
            </Input>
          </div>

          <div className="flex gap-3 pt-4">
            <Button
              styles="px-4 py-2 flex-1"
              buttonText="Cancel"
              variant="tertiary"
              type="button"
              onClick={() => {
                onClose();
                setFormData(defaultFormData);
              }}
            />
            <Button
              type="submit"
              variant="primary"
              styles="px-4 py-2 flex-1"
              buttonText={editData ? "Save " : "Add Type"}
            />
          </div>
        </div>
      </form>
    </Modal>
  );
}
