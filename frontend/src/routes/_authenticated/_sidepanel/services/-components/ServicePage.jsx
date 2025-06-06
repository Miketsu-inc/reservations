import Button from "@components/Button";
import Card from "@components/Card";
import Input from "@components/Input";
import Select from "@components/Select";
import Switch from "@components/Switch";
import { Block } from "@tanstack/react-router";
import { useEffect, useMemo, useState } from "react";
import ServicePhases from "./ServicePhases";
import { useServicePhases } from "./servicehooks";

export default function ServicePage({ service, onSave }) {
  const originalData = useMemo(
    () => ({
      id: service?.id,
      name: service?.name || "",
      description: service?.description || "",
      color: service?.color || "#2334b8",
      price: service?.price || "",
      price_note: service?.price_note || "",
      cost: service?.cost || "",
      category_id: service?.category_id || null,
      is_active: service?.is_active || true,
      phases: service?.phases || [],
      products: service?.products || [],
      categories: service?.categories || [],
    }),
    [service]
  );
  const [serviceData, setServiceData] = useState(originalData);
  const [lastSavedData, setLastSavedData] = useState(originalData);

  useEffect(() => {
    setServiceData(originalData);
    setLastSavedData(originalData);
  }, [originalData]);

  const phaseHandlers = useServicePhases(setServiceData);

  function updateServiceData(data) {
    setServiceData((prev) => ({ ...prev, ...data }));
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
      <div className="flex h-screen px-4 py-2 md:px-0 md:py-0">
        <div className="my-6 w-full">
          <div className="flex flex-col gap-4">
            <Card styles="sticky top-14 md:top-0 z-10 flex flex-row items-center justify-between">
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
                      options={[
                        { value: null, label: "No category" },
                        ...serviceData.categories.map((category) => ({
                          value: category.id,
                          label: category.name,
                        })),
                      ]}
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
                    <p>Show on booking page</p>
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
          </div>
        </div>
      </div>
    </Block>
  );
}
