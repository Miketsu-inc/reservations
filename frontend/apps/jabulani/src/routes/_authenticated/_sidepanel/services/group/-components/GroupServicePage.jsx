import { InfoIcon } from "@reservations/assets";
import {
  Button,
  Card,
  DeleteModal,
  Input,
  Select,
  Switch,
  Textarea,
  TooltipContent,
  TooltipTrigger,
  Tootlip,
} from "@reservations/components";

import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { Block, useRouter } from "@tanstack/react-router";
import { useEffect, useMemo, useState } from "react";

import ProductAdder from "../../-components/ProductAdder";
import ServiceSchedulingSettings from "../../-components/ServiceSchedulingSettings";

const priceTypeOptions = [
  { label: "fixed", value: "fixed" },
  { label: "from", value: "from" },
  { label: "free", value: "free" },
];

export default function GroupServicePage({
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
      price: service?.price ?? null,
      price_type: service?.price_type || "fixed",
      cost: service?.cost ?? null,
      category_id: service?.category_id || null,
      is_active: service?.is_active ?? true,
      duration: service?.duration || "",
      duration_unit: service?.duration_unit || "min",
      min_participants: service?.min_participants || undefined,
      max_participants: service?.max_participants || "",
      settings: {
        cancel_deadline: service?.settings?.cancel_deadline || null,
        booking_window_min: service?.settings?.booking_window_min || null,
        booking_window_max: service?.settings?.booking_window_max || null,
        buffer_time: service?.settings?.buffer_time || null,
      },
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
        message: "Group Service deleted successfully",
        variant: "success",
      });
    }
  }

  function submitHandler(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    const durationInMinutes =
      serviceData.duration_unit === "hour"
        ? serviceData.duration * 60
        : serviceData.duration;
    const data = {
      ...serviceData,
      duration: Number(durationInMinutes),
    };
    delete data.duration_unit;

    setLastSavedData(serviceData);
    onSave(data);
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
            <form className="flex flex-col gap-4" onSubmit={submitHandler}>
              <Card
                styles="sticky top-14 md:top-0 z-10 flex flex-row items-center
                  justify-between gap-2"
              >
                <div className="flex items-center gap-2">
                  <p className="text-xl">
                    {serviceData.name || "New Group Service"}
                  </p>
                </div>
                <Button
                  styles="py-2 px-6"
                  variant="primary"
                  buttonText="Save"
                  type="submit"
                />
              </Card>

              <Card styles="flex flex-col md:flex-row gap-8">
                <div className="flex flex-col gap-6 md:w-1/2">
                  <div className="flex flex-row gap-4">
                    <div
                      style={{ backgroundColor: serviceData.color }}
                      className="size-28 shrink-0 overflow-hidden rounded-lg"
                    >
                      <img
                        className="size-full object-cover"
                        src="https://dummyimage.com/120x120/d156c3/000000.jpg"
                        alt="service photo"
                      />
                    </div>

                    <div className="flex w-full flex-col justify-between">
                      <Input
                        styles="p-2"
                        id="ServiceName"
                        name="ServiceName"
                        type="text"
                        labelText="Service name"
                        placeholder="e.g. Yoga Class"
                        childrenSide="left"
                        value={serviceData.name}
                        inputData={(data) =>
                          updateServiceData({ name: data.value })
                        }
                      >
                        <input
                          id="color"
                          className="border-input_border_color size-10.5
                            cursor-pointer rounded-l-lg border bg-transparent"
                          name="color"
                          type="color"
                          value={serviceData.color}
                          onChange={(e) =>
                            updateServiceData({ color: e.target.value })
                          }
                        />
                      </Input>

                      <div className="flex flex-row items-center gap-3">
                        <Switch
                          defaultValue={serviceData.is_active}
                          onSwitch={() =>
                            updateServiceData({
                              is_active: !serviceData.is_active,
                            })
                          }
                        />
                        <div className="flex flex-row items-center gap-1">
                          <p>Active service</p>
                          <span className="hidden items-center md:flex">
                            <Tootlip>
                              <TooltipTrigger type="button">
                                <InfoIcon
                                  styles="size-4 stroke-gray-500
                                    dark:stroke-gray-400"
                                />
                              </TooltipTrigger>
                              <TooltipContent side="right">
                                <p>
                                  Only active services will show up on your
                                  booking page
                                </p>
                              </TooltipContent>
                            </Tootlip>
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div className="flex flex-col gap-4">
                    <Input
                      styles="p-2"
                      id="duration"
                      name="duration"
                      type="number"
                      min={1}
                      max={serviceData.duration_unit === "hour" ? 24 : 1440}
                      labelText="Duration"
                      placeholder="30"
                      value={serviceData.duration}
                      inputData={(data) =>
                        updateServiceData({ duration: Number(data.value) })
                      }
                    >
                      <Select
                        styles="rounded-l-none w-32! xl:w-52!"
                        value={serviceData.duration_unit || "min"}
                        options={[
                          { value: "min", label: "minutes" },
                          { value: "hour", label: "hours" },
                        ]}
                        onSelect={(option) =>
                          updateServiceData({ duration_unit: option.value })
                        }
                      />
                    </Input>
                    <div className="flex w-full gap-4 sm:grid sm:grid-cols-2">
                      <Input
                        styles="p-2 peer flex-1 w-full"
                        id="price"
                        name="price"
                        type="number"
                        min={0}
                        max={1000000}
                        labelText="Price (Per Person)"
                        placeholder={
                          serviceData.price_type === "free" ? "0" : "1000"
                        }
                        required={false}
                        value={serviceData.price?.number || ""}
                        disabled={serviceData.price_type === "free"}
                        inputData={(data) =>
                          updateServiceData({
                            price: {
                              number: data.value,
                              currency: serviceData.price?.currency || "HUF",
                            },
                          })
                        }
                      >
                        <p
                          className={`border-input_border_color
                            peer-disabled:text-text_color/70
                            peer-disabled:border-input_border_color/60
                            rounded-r-lg border px-4 py-2
                            peer-disabled:bg-gray-200/60
                            peer-disabled:dark:bg-gray-700/20`}
                        >
                          {serviceData.price?.currency || "HUF"}
                        </p>
                      </Input>
                      <label className="flex w-auto flex-col">
                        <div className="flex items-center gap-2 pb-1">
                          <span className="text-sm">Price Note</span>
                          <InfoIcon styles="size-4 stroke-gray-500
                            dark:stroke-gray-400" />
                        </div>
                        <Select
                          value={serviceData.price_type}
                          styles="w-28! sm:w-full!"
                          options={priceTypeOptions}
                          onSelect={(option) => {
                            updateServiceData({
                              price: {
                                number: 0,
                                currency: serviceData.price?.currency || "HUF",
                              },
                            });
                            updateServiceData({ price_type: option.value });
                          }}
                        />
                      </label>
                    </div>

                    <Input
                      styles="p-2 peer"
                      id="cost"
                      name="cost"
                      type="number"
                      min={0}
                      required={false}
                      labelText="Cost (Per Person)"
                      placeholder="0"
                      value={serviceData.cost?.number || ""}
                      inputData={(data) =>
                        updateServiceData({
                          cost: {
                            number: data.value,
                            currency: serviceData.cost?.currency || "HUF",
                          },
                        })
                      }
                    >
                      <p
                        className="border-input_border_color rounded-r-lg border
                          px-4 py-2"
                      >
                        {serviceData.cost?.currency || "HUF"}
                      </p>
                    </Input>
                  </div>
                </div>

                <div className="flex flex-1 flex-col gap-5 md:w-1/2">
                  <Select
                    value={serviceData.category_id}
                    labelText="Service Category"
                    required={false}
                    options={categoryOptions}
                    onSelect={(option) =>
                      updateServiceData({ category_id: option.value })
                    }
                  />

                  <div className="grid grid-cols-2 gap-4">
                    <Input
                      styles="p-2"
                      id="min_participants"
                      name="min_participants"
                      type="number"
                      min={1}
                      labelText="Min Participants"
                      placeholder="1"
                      required={false}
                      value={serviceData.min_participants}
                      inputData={(data) =>
                        updateServiceData({
                          min_participants: Number(data.value),
                        })
                      }
                    />
                    <Input
                      styles="p-2"
                      id="max_participants"
                      name="max_participants"
                      type="number"
                      min={1}
                      labelText="Max Capacity"
                      placeholder="10"
                      value={serviceData.max_participants}
                      inputData={(data) =>
                        updateServiceData({
                          max_participants: Number(data.value),
                        })
                      }
                    />
                  </div>

                  <Textarea
                    styles="p-2 h-full min-h-35"
                    id="description"
                    name="description"
                    labelText="Description"
                    required={false}
                    placeholder="About this group session..."
                    value={serviceData.description}
                    inputData={(data) =>
                      updateServiceData({ description: data.value })
                    }
                  />
                </div>
              </Card>
            </form>
            <ProductAdder
              availableProducts={products}
              usedProducts={serviceData.used_products}
              onUpdate={(updated) =>
                updateServiceData({ used_products: updated })
              }
            />
            <ServiceSchedulingSettings
              onUpdate={updateServiceData}
              settings={serviceData.settings}
            />
            {service && (
              <Button
                type="button"
                styles="py-4 mb-2 shadow-none bg-transparent
                  hover:bg-transparent text-red-500!"
                buttonText="Delete Group Service"
                onClick={() => setShowDeleteModal(true)}
              />
            )}
          </div>
        </div>
      </div>
    </Block>
  );
}
