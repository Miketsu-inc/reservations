import Button from "@components/Button";
import Card from "@components/Card";
import DeleteModal from "@components/DeleteModal";
import Input from "@components/Input";
import Select from "@components/Select";
import Switch from "@components/Switch";
import InfoIcon from "@icons/InfoIcon";
import { useToast } from "@lib/hooks";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { Block, useRouter } from "@tanstack/react-router";
import { useEffect, useMemo, useState } from "react";
import ServicePhases from "./ServicePhases";
import { useServicePhases } from "./servicehooks";

export default function ServicePage({ service, categories, onSave, route }) {
  const originalData = useMemo(
    () => ({
      id: service?.id,
      name: service?.name || "",
      description: service?.description || "",
      color: service?.color || "#2334b8",
      price: service?.price ?? "",
      price_note: service?.price_note || "",
      cost: service?.cost ?? "",
      category_id: service?.category_id || null,
      is_active: service?.is_active ?? true,
      phases: service?.phases || [],
      used_products: service?.used_products || [],
    }),
    [service]
  );
  const router = useRouter();
  const [serviceData, setServiceData] = useState(originalData);
  const [lastSavedData, setLastSavedData] = useState(originalData);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const { showToast } = useToast();

  useEffect(() => {
    setServiceData(originalData);
    setLastSavedData(originalData);
  }, [originalData]);

  const phaseHandlers = useServicePhases(setServiceData);

  const categoryOptions = useMemo(
    () => [
      { value: null, label: "No category" },
      ...categories.map((category) => ({
        value: category.id,
        label: category.name,
      })),
    ],
    [categories]
  );

  function updateServiceData(data) {
    setServiceData((prev) => ({ ...prev, ...data }));
  }

  async function deleteHandler() {
    const response = await fetch(`/api/v1/merchants/services/${service.id}`, {
      method: "DELETE",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    });

    if (!response.ok) {
      invalidateLocalSotrageAuth(response.status);
      const result = await response.json();
      showToast({ message: result.error.message, variant: "error" });
    } else {
      router.navigate({
        from: route.fullPath,
        to: "/services",
      });
      showToast({
        message: "Service deleted successfully",
        variant: "success",
      });
    }
  }

  return (
    <Block
      shouldBlockFn={() => {
        if (JSON.stringify(serviceData) === JSON.stringify(lastSavedData))
          return false;

        const canLeave = confirm(
          "You have unsaved changes, are you sure you want to leave?"
        );
        return !canLeave;
      }}
    >
      {service && (
        <DeleteModal
          isOpen={showDeleteModal}
          onClose={() => setShowDeleteModal(false)}
          onDelete={deleteHandler}
          itemName={service.name}
        />
      )}
      <div className="flex h-screen px-4 py-2 md:px-0 md:py-0">
        <div className="my-6 w-full">
          <div className="flex flex-col gap-4">
            <Card styles="sticky top-14 md:top-0 z-10 flex flex-row items-center justify-between gap-2">
              <p className="text-xl">{serviceData.name || "New service"}</p>
              <Button
                styles="py-2 px-6"
                variant="primary"
                buttonText="Save"
                onClick={() => {
                  setLastSavedData(serviceData);
                  onSave(serviceData);
                }}
              />
            </Card>
            <div className="flex flex-col gap-4 md:flex-row">
              <Card styles="flex flex-col gap-4 md:w-1/2 md:flex-row">
                <div
                  style={{ backgroundColor: serviceData.color }}
                  className="size-28 shrink-0 overflow-hidden rounded-lg xl:size-[120px]"
                >
                  <img
                    className="size-full object-cover"
                    src="https://dummyimage.com/120x120/d156c3/000000.jpg"
                    alt="service photo"
                  ></img>
                </div>
                <div className="flex w-full flex-col gap-6">
                  <div className="flex w-full flex-row items-end gap-2">
                    <input
                      id="color"
                      className="size-11 cursor-pointer bg-transparent"
                      name="color"
                      type="color"
                      value={serviceData.color}
                      onChange={(e) =>
                        updateServiceData({ color: e.target.value })
                      }
                    />
                    <Input
                      styles="p-2"
                      id="ServiceName"
                      name="ServiceName"
                      type="text"
                      labelText="Service name"
                      hasError={false}
                      placeholder="e.g. hair styling"
                      value={serviceData.name}
                      inputData={(data) =>
                        updateServiceData({ name: data.value })
                      }
                    />
                  </div>
                  <div className="flex flex-row items-center gap-4">
                    <Input
                      styles="p-2"
                      id="price"
                      name="price"
                      type="number"
                      min={0}
                      labelText="Price (HUF)"
                      hasError={false}
                      placeholder="1000"
                      errorText="Price must be between 1 and 1,000,000"
                      value={serviceData.price}
                      inputData={(data) =>
                        updateServiceData({ price: Number(data.value) })
                      }
                    />
                    <Input
                      styles="p-2"
                      id="PriceNote"
                      name="PriceNote"
                      type="text"
                      labelText="Price note"
                      hasError={false}
                      placeholder="e.g. -tol"
                      value={serviceData.price_note}
                      inputData={(data) =>
                        updateServiceData({ price_note: data.value })
                      }
                    />
                  </div>
                  <Input
                    styles="p-2"
                    id="cost"
                    name="cost"
                    type="number"
                    min={0}
                    labelText="Cost (HUF)"
                    hasError={false}
                    placeholder="500"
                    required={false}
                    errorText="Cost must be between 0 and 1,000,000"
                    value={serviceData.cost}
                    inputData={(data) =>
                      updateServiceData({ cost: Number(data.value) })
                    }
                  />
                  <label>
                    Service category
                    <Select
                      value={serviceData.category_id}
                      options={categoryOptions}
                      onSelect={(option) =>
                        updateServiceData({ category_id: option.value })
                      }
                    />
                  </label>
                  <div className="flex flex-row items-center gap-3">
                    <Switch
                      defaultValue={serviceData.is_active}
                      onSwitch={() =>
                        updateServiceData({ is_active: !serviceData.is_active })
                      }
                    />
                    <div className="flex flex-row items-center gap-1">
                      <p>Active service</p>
                      <InfoIcon styles="size-4 stroke-gray-500 dark:stroke-gray-400" />
                    </div>
                  </div>
                  <div className="flex flex-col gap-1">
                    <label htmlFor="description">Description</label>
                    <textarea
                      id="description"
                      name="description"
                      placeholder="About this service..."
                      className="bg-bg_color focus:border-primary max-h-20 min-h-20 w-full rounded-lg border
                        border-gray-300 p-2 text-sm outline-hidden md:max-h-32 md:min-h-32"
                      value={serviceData.description}
                      onChange={(e) =>
                        updateServiceData({ description: e.target.value })
                      }
                    />
                  </div>
                </div>
              </Card>
              <div className="flex h-fit flex-col gap-2 md:w-1/2">
                <ServicePhases
                  phases={serviceData.phases}
                  onAddPhase={phaseHandlers.addPhase}
                  onUpdatePhase={phaseHandlers.updatePhase}
                  onRemovePhase={phaseHandlers.removePhase}
                />
              </div>
            </div>
            <Card>
              <p>Products</p>
            </Card>
            <Card>
              <p>Employees</p>
            </Card>
            {service && (
              <Button
                type="button"
                styles="py-4 mb-2 shadow-none bg-transparent hover:bg-transparent !text-red-500"
                buttonText="Delete service"
                onClick={() => setShowDeleteModal(true)}
              ></Button>
            )}
          </div>
        </div>
      </div>
    </Block>
  );
}
