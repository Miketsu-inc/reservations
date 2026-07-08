import { Button, Input, Select } from "@reservations/components";
import { useState } from "react";

const priceTypeOptions = [
  { label: "fixed", value: "fixed" },
  { label: "from", value: "from" },
  { label: "free", value: "free" },
];

export default function ServiceEditor({ serviceData, onSave }) {
  const [service, setService] = useState(serviceData);

  function updateService(data) {
    setService((prev) => ({ ...prev, ...data }));
  }

  function handleSubmit(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    onSave(service);
  }

  return (
    <form className="relative h-full w-full" onSubmit={handleSubmit}>
      <div className="flex flex-col gap-8 px-6">
        <div className="flex items-center justify-between">
          <p className="text-2xl font-semibold">Edit Service</p>
          <div
            className="w-1/3 rounded-full py-2"
            style={{ backgroundColor: service.color }}
          ></div>
        </div>
        <div className="flex w-full flex-row items-end gap-2">
          <Input
            id="ServiceName"
            name="ServiceName"
            type="text"
            labelText="Service name"
            placeholder="e.g. hair styling"
            childrenSide="left"
            value={service.name}
            inputData={(data) => updateService({ name: data.value })}
          />
        </div>
        <Input
          id="duration"
          name="duration"
          type="number"
          min={1}
          max={service.duration_unit === "hour" ? 24 : 1440}
          labelText="Duration"
          placeholder="30"
          value={service.duration}
          inputData={(data) => updateService({ duration: Number(data.value) })}
          disabled={true}
        >
          <Select
            styles="rounded-l-none w-32!"
            value={service.duration_unit || "min"}
            options={[
              { value: "min", label: "minutes" },
              { value: "hour", label: "hours" },
            ]}
            onSelect={(option) =>
              updateService({ duration_unit: option.value })
            }
            disabled={true}
          />
        </Input>
        <div className="flex w-full gap-4">
          <Input
            styles="peer flex-1"
            id="price"
            name="price"
            type="number"
            min={0}
            max={1000000}
            labelText="Price"
            placeholder={service.price_type === "free" ? "0" : "1000"}
            required={false}
            value={service.price?.number || ""}
            disabled={service.price_type === "free"}
            inputData={(data) =>
              updateService({
                price: {
                  number: data.value,
                  currency: data.price?.currency || "HUF",
                },
                // if currency is chnageable this needs a helper function
                formatted_price: `${data.value || 0} Ft`,
              })
            }
          >
            <p
              className={`border-input_border_color
                peer-disabled:text-text_color/70
                peer-disabled:border-input_border_color/60 rounded-r-lg border
                px-4 py-2 peer-disabled:bg-gray-200/60
                peer-disabled:dark:bg-gray-700/20`}
            >
              {service.price?.currency || "HUF"}
            </p>
          </Input>
          <label className="flex w-auto flex-col gap-1">
            <span className="text-sm">Price Note</span>

            <Select
              value={service.price_type}
              styles="w-28!"
              options={priceTypeOptions}
              onSelect={(option) => {
                if (option.value === "free") {
                  updateService({
                    price: {
                      number: 0,
                      currency: service.price?.currency || "HUF",
                    },
                    formatted_price: "0 Ft",
                  });
                }
                updateService({ price_type: option.value });
              }}
            />
          </label>
        </div>
      </div>
      <div
        className="border-border_color bg-layer_bg items center fixed right-0
          bottom-0 left-0 flex w-full border-t px-6 py-4"
      >
        <Button
          styles="py-2 px-4 w-full"
          variant="primary"
          name="createButton"
          buttonText="Save"
        />
      </div>
    </form>
  );
}
