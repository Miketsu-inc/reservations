import { InformationCircleIcon } from "@hugeicons/core-free-icons";
import { Icon, Input, Select } from "@reservations/components";
import ServicePhases from "./ServicePhases";

const priceTypeOptions = [
  { label: "fixed", value: "fixed" },
  { label: "from", value: "from" },
  { label: "free", value: "free" },
];

export function ServicePricingDuration({ service, setService, onUpdate }) {
  const isGroupService = service.booking_type !== "appointment";

  return (
    <div className="flex h-fit flex-col gap-6">
      <div className="flex w-full gap-4 sm:grid sm:grid-cols-2">
        <Input
          styles="peer flex-1 w-full"
          id="price"
          name="price"
          type="number"
          min={0}
          max={1000000}
          labelText={isGroupService ? "Price (per person)" : "Price"}
          placeholder={service.price_type === "free" ? "0" : "1000"}
          required={false}
          value={service.price?.number || ""}
          disabled={service.price_type === "free"}
          inputData={(data) =>
            onUpdate({
              price: {
                number: data.value,
                currency: service.price?.currency || "HUF",
              },
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
        <label className="flex w-auto flex-col">
          <div className="flex items-center gap-2 pb-1">
            <span className="text-sm">Price Note</span>
            <Icon
              icon={InformationCircleIcon}
              styles="size-4 text-gray-500 dark:text-gray-400"
            />
          </div>
          <Select
            value={service.price_type}
            styles="w-28! sm:w-full!"
            options={priceTypeOptions}
            onSelect={(option) => {
              onUpdate({
                price: {
                  number: 0,
                  currency: service.price?.currency || "HUF",
                },
              });
              onUpdate({
                price_type: option.value,
              });
            }}
          />
        </label>
      </div>
      <ServicePhases service={service} setService={setService} />
    </div>
  );
}
