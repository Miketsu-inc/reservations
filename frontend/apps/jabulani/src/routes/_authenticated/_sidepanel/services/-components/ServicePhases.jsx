import {
  AddCircleIcon,
  ArrowLeft02Icon,
  Delete02Icon,
  MoreVerticalIcon,
} from "@hugeicons/core-free-icons";
import {
  Button,
  Icon,
  Input,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
  Select,
} from "@reservations/components";
import { formatDuration } from "@reservations/lib";
import { useMemo } from "react";
import { useServicePhases } from "./servicehooks";

const durationOptions = [
  { value: "min", label: "minutes" },
  { value: "hour", label: "hours" },
];

const phaseTypeOptions = [
  { value: "active", label: "active" },
  { value: "wait", label: "wait" },
];

function toMinutes(duration, durationUnit) {
  if (duration === "") return "";

  const value = Number(duration);
  return Math.round(durationUnit === "hour" ? value * 60 : value);
}

function fromMinutes(duration, durationUnit) {
  if (duration === "") return "";

  return durationUnit === "hour" ? duration / 60 : duration;
}

export default function ServicePhases({ service, setService }) {
  const phases = service.phases;
  const {
    addPhase: appendPhase,
    updatePhase,
    removePhase,
    reorderPhase,
  } = useServicePhases(setService);

  const isGroupService = service.booking_type !== "appointment";

  const durationSum = useMemo(() => {
    return phases.reduce((total, phase) => {
      return total + (Number(phase.duration) || 0);
    }, 0);
  }, [phases]);

  const singlePhase = phases[0];
  const singleDurationUnit =
    singlePhase?.duration_unit || service.duration_unit || "min";
  const singleDuration = singlePhase
    ? fromMinutes(singlePhase.duration, singleDurationUnit)
    : service.duration;

  function updateSinglePhase(updates) {
    const phase = {
      ...singlePhase,
      id: singlePhase?.id ?? -1,
      name: singlePhase?.name || "",
      sequence: 1,
      duration: singlePhase
        ? singlePhase.duration
        : toMinutes(service.duration, service.duration_unit || "min"),
      duration_unit: singleDurationUnit,
      phase_type: isGroupService
        ? "active"
        : singlePhase?.phase_type || "active",
      ...updates,
    };

    singlePhase ? updatePhase(phase) : appendPhase(phase);
  }

  function addPhase() {
    if (phases.length === 0) {
      appendPhase({
        name: "",
        duration: toMinutes(service.duration, service.duration_unit || "min"),
        duration_unit: service.duration_unit || "min",
        phase_type: "active",
      });
    }

    appendPhase({
      name: "",
      duration: "",
      duration_unit: "min",
      phase_type: "active",
    });
  }

  return (
    <div className="flex flex-col gap-4">
      {!isGroupService && phases.length > 1 ? (
        <div className="flex flex-col gap-1">
          <span className="flex items-center gap-1 text-sm">
            Duration
            <span className="text-base leading-none text-red-500">*</span>
          </span>
          <div className="flex flex-col gap-4">
            {phases.map((phase, index) => (
              <PhaseInput
                key={phase.sequence}
                phase={phase}
                isFirst={index === 0}
                isLast={index === phases.length - 1}
                onDelete={removePhase}
                onUpdate={updatePhase}
                onMoveUp={() => reorderPhase(phase.sequence, -1)}
                onMoveDown={() => reorderPhase(phase.sequence, 1)}
              />
            ))}
          </div>
          <p className="text-text_color/70 text-sm">
            Total duration: {formatDuration(durationSum)}
          </p>
        </div>
      ) : (
        <Input
          id="duration"
          name="duration"
          type="number"
          min={1}
          max={singleDurationUnit === "hour" ? 24 : 1440}
          labelText="Duration"
          placeholder="30"
          value={singleDuration}
          inputData={(data) =>
            updateSinglePhase({
              duration: toMinutes(data.value, singleDurationUnit),
            })
          }
        >
          <Select
            styles="rounded-l-none w-32! xl:w-52!"
            value={singleDurationUnit}
            options={durationOptions}
            onSelect={(option) =>
              updateSinglePhase({
                duration: toMinutes(singleDuration, option.value),
                duration_unit: option.value,
              })
            }
          />
        </Input>
      )}
      {!isGroupService && (
        <Button
          variant="tertiary"
          buttonText="Add extra time"
          styles="w-fit px-4 py-2"
          type="button"
          onClick={addPhase}
        >
          <Icon icon={AddCircleIcon} styles="size-5 mr-2" />
        </Button>
      )}
    </div>
  );
}

function PhaseInput({
  phase,
  isFirst,
  isLast,
  onDelete,
  onUpdate,
  onMoveUp,
  onMoveDown,
}) {
  const durationUnit = phase.duration_unit || "min";

  return (
    <div className="flex flex-row items-center gap-2">
      <Input
        id={"phase-" + phase.sequence + "-duration"}
        name={"phase-" + phase.sequence + "-duration"}
        type="number"
        min={1}
        max={durationUnit === "hour" ? 24 : 1440}
        placeholder="30"
        value={fromMinutes(phase.duration, durationUnit)}
        inputData={(data) =>
          onUpdate({
            ...phase,
            duration: toMinutes(data.value, durationUnit),
          })
        }
      >
        <Select
          styles="rounded-l-none max-w-52"
          value={durationUnit}
          options={durationOptions}
          onSelect={(option) =>
            onUpdate({
              ...phase,
              duration: toMinutes(
                fromMinutes(phase.duration, durationUnit),
                option.value
              ),
              duration_unit: option.value,
            })
          }
        />
      </Input>
      <Select
        styles="max-w-28"
        value={phase.phase_type || "active"}
        options={phaseTypeOptions}
        onSelect={(option) =>
          onUpdate({
            ...phase,
            phase_type: option.value,
          })
        }
      />
      <Popover>
        <PopoverTrigger asChild>
          <button
            type="button"
            className="hover:bg-hvr_gray cursor-pointer rounded-lg p-1"
          >
            <Icon icon={MoreVerticalIcon} styles="size-8 rotate-90" />
          </button>
        </PopoverTrigger>
        <PopoverContent side="left">
          <div
            className="*:hover:bg-hvr_gray flex flex-col items-start *:flex
              *:w-full *:cursor-pointer *:flex-row *:items-center *:gap-4
              *:rounded-lg *:p-2"
          >
            <PopoverClose asChild>
              <button
                type="button"
                disabled={isLast}
                onClick={onMoveDown}
                className={`${
                  isLast ? "opacity-35" : "hover:bg-hvr_gray cursor-pointer"
                }`}
              >
                <Icon icon={ArrowLeft02Icon} styles="size-6 -rotate-90" />
                <p>Move down</p>
              </button>
            </PopoverClose>
            <PopoverClose asChild>
              <button
                type="button"
                disabled={isFirst}
                onClick={onMoveUp}
                className={`${isFirst ? "opacity-35" : "hover:bg-hvr_gray cursor-pointer"}`}
              >
                <Icon icon={ArrowLeft02Icon} styles="size-6 rotate-90" />
                <p>Move up</p>
              </button>
            </PopoverClose>
            <PopoverClose asChild>
              <button
                type="button"
                onClick={() => onDelete(phase.sequence)}
                className="text-red-600 dark:text-red-500"
              >
                <Icon icon={Delete02Icon} styles="size-6" />
                <p>Delete</p>
              </button>
            </PopoverClose>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
}
