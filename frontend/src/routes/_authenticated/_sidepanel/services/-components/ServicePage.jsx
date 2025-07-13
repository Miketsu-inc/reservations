import Button from "@components/Button";
import Card from "@components/Card";
import DeleteModal from "@components/DeleteModal";
import Input from "@components/Input";
import Select from "@components/Select";
import Switch from "@components/Switch";
import { TooltipContent, TooltipTrigger, Tootlip } from "@components/Tooltip";
import InfoIcon from "@icons/InfoIcon";
import { useToast } from "@lib/hooks";
import { invalidateLocalStorageAuth } from "@lib/lib";
import { Block, useRouter } from "@tanstack/react-router";
import { useEffect, useMemo, useState } from "react";
import ProductAdder from "./ProductAdder";
import ServicePhases from "./ServicePhases";
import { useServicePhases } from "./servicehooks";

export default function ServicePage({
  service,
  categories,
  products,
  onSave,
  route,
}) {
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
      invalidateLocalStorageAuth(response.status);
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
                  if (serviceData.phases.length > 0) {
                    setLastSavedData(serviceData);
                    onSave(serviceData);
                  } else {
                    showToast({
                      message: "Please add at least one service phase",
                      variant: "error",
                    });
                  }
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
                    <Input
                      styles="p-2"
                      id="ServiceName"
                      name="ServiceName"
                      type="text"
                      labelText="Service name"
                      placeholder="e.g. hair styling"
                      childrenSide="left"
                      value={serviceData.name}
                      inputData={(data) =>
                        updateServiceData({ name: data.value })
                      }
                    >
                      <input
                        id="color"
                        className="border-input_border_color size-10.5 cursor-pointer rounded-l-lg border
                          bg-transparent"
                        name="color"
                        type="color"
                        value={serviceData.color}
                        onChange={(e) =>
                          updateServiceData({ color: e.target.value })
                        }
                      />
                    </Input>
                  </div>
                  <div className="flex flex-row gap-4">
                    <Input
                      styles="p-2"
                      id="price"
                      name="price"
                      type="number"
                      min={0}
                      max={1000000}
                      labelText="Price"
                      placeholder="1000"
                      value={serviceData.price}
                      inputData={(data) =>
                        updateServiceData({ price: Number(data.value) })
                      }
                    >
                      <p className="border-input_border_color rounded-r-lg border px-4 py-2">
                        HUF
                      </p>
                    </Input>
                    <Input
                      styles="p-2"
                      id="PriceNote"
                      name="PriceNote"
                      type="text"
                      labelText="Price note"
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
                    max={1000000}
                    labelText="Cost"
                    placeholder="500"
                    required={false}
                    value={serviceData.cost}
                    inputData={(data) =>
                      updateServiceData({ cost: Number(data.value) })
                    }
                  >
                    <p className="border-input_border_color rounded-r-lg border px-4 py-2">
                      HUF
                    </p>
                  </Input>
                  <label className="flex flex-col gap-1">
                    <span className="text-sm">Service category</span>
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
                      <span className="hidden items-center md:flex">
                        <Tootlip>
                          <TooltipTrigger>
                            <InfoIcon styles="size-4 stroke-gray-500 dark:stroke-gray-400" />
                          </TooltipTrigger>
                          <TooltipContent side="right">
                            <p>
                              Only active services will show up on your booking
                              page
                            </p>
                          </TooltipContent>
                        </Tootlip>
                      </span>
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
            <ProductAdder
              availableProducts={products}
              usedProducts={serviceData.used_products}
              onUpdate={(updated) =>
                updateServiceData({ used_products: updated })
              }
            />
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
