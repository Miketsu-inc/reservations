import { InformationCircleIcon } from "@hugeicons/core-free-icons";
import {
  Icon,
  Input,
  Select,
  Switch,
  Textarea,
  TooltipContent,
  TooltipTrigger,
  Tootlip,
} from "@reservations/components";
import { useMemo } from "react";

const bookingTypeOptions = [
  { label: "appointment", value: "appointment" },
  { label: "class", value: "class" },
];

export function ServiceBasicDetails({ service, categories, onUpdate }) {
  const isGroupService = service.booking_type !== "appointment";

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

  return (
    <div className="flex w-full flex-col gap-6">
      <div
        className="flex w-full flex-col items-center gap-4 md:flex-row md:gap-2"
      >
        <Input
          id="ServiceName"
          name="ServiceName"
          type="text"
          labelText="Service name"
          placeholder="e.g. hair styling"
          childrenSide="left"
          value={service.name}
          inputData={(data) => onUpdate({ name: data.value })}
        >
          <input
            id="color"
            className="border-input_border_color size-10.5 cursor-pointer
              rounded-l-lg border bg-transparent"
            name="color"
            type="color"
            value={service.color}
            onChange={(e) =>
              onUpdate({
                color: e.target.value,
              })
            }
          />
        </Input>
        <Select
          labelText="Booking type"
          value={service.booking_type}
          options={bookingTypeOptions}
          onSelect={(option) => {
            onUpdate({
              booking_type: option.value,
            });
          }}
        />
      </div>
      {isGroupService && (
        <div className="grid grid-cols-2 gap-4">
          <Input
            id="min_participants"
            name="min_participants"
            type="number"
            min={1}
            labelText="Min Participants"
            placeholder="1"
            required={false}
            value={service.min_participants || ""}
            inputData={(data) =>
              onUpdate({
                min_participants: Number(data.value),
              })
            }
          />
          <Input
            id="max_participants"
            name="max_participants"
            type="number"
            min={1}
            labelText="Max Capacity"
            placeholder="10"
            value={service.max_participants || ""}
            inputData={(data) =>
              onUpdate({
                max_participants: Number(data.value),
              })
            }
          />
        </div>
      )}
      <Select
        value={service.category_id}
        labelText="Service Category"
        required={false}
        options={categoryOptions}
        onSelect={(option) =>
          onUpdate({
            category_id: option.value,
          })
        }
      />
      <div className="flex flex-row items-center gap-3">
        <Switch
          defaultValue={service.is_active}
          onSwitch={() =>
            onUpdate({
              is_active: !service.is_active,
            })
          }
        />
        <div className="flex flex-row items-center gap-1">
          <p>Active service</p>
          <span className="hidden items-center md:flex">
            <Tootlip>
              <TooltipTrigger>
                <Icon
                  icon={InformationCircleIcon}
                  styles="size-4 text-gray-500 dark:text-gray-400"
                />
              </TooltipTrigger>
              <TooltipContent side="right">
                <p>Only active services will show up on your booking page</p>
              </TooltipContent>
            </Tootlip>
          </span>
        </div>
      </div>
      <Textarea
        styles="p-2 max-h-20 min-h-20 md:max-h-32 md:min-h-32"
        id="description"
        name="description"
        labelText="Description"
        required={false}
        placeholder="About this service..."
        value={service.description}
        inputData={(data) => onUpdate({ description: data.value })}
      />
    </div>
  );
}
