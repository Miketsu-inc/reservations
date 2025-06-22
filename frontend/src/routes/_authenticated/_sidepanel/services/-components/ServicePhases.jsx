import Button from "@components/Button";
import Card from "@components/Card";
import Input from "@components/Input";
import Select from "@components/Select";
import Switch from "@components/Switch";
import ClockIcon from "@icons/ClockIcon";
import EditIcon from "@icons/EditIcon";
import HourGlassIcon from "@icons/HourGlassIcon";
import InfoIcon from "@icons/InfoIcon";
import PlusIcon from "@icons/PlusIcon";
import ServicesIcon from "@icons/ServicesIcon";
import TrashBinIcon from "@icons/TrashBinIcon";
import { formatDuration } from "@lib/datetime";
import { useWindowSize } from "@lib/hooks";
import { useMemo, useState } from "react";

export default function ServicePhases({
  phases = [],
  onAddPhase,
  onUpdatePhase,
  onRemovePhase,
}) {
  const [showAddForm, setShowAddForm] = useState(false);
  const [editPhase, setEditPhase] = useState(null);
  const windowSize = useWindowSize();

  const isWindowSmall = windowSize === "sm";

  // Sort phases by sequence to ensure proper display order
  const sortedPhases = useMemo(() => {
    return [...phases].sort((a, b) => (a.sequence || 0) - (b.sequence || 0));
  }, [phases]);

  const durationSum = useMemo(() => {
    return phases.reduce((total, phase) => {
      return total + phase.duration || 0;
    }, 0);
  }, [phases]);

  return (
    <>
      <div className="flex flex-row items-center gap-1">
        <p className="text-lg">Service phases</p>
        <InfoIcon styles="size-4 stroke-gray-500 dark:stroke-gray-400" />
      </div>
      {!phases || phases.length === 0 ? (
        <div className="bg-layer_bg border-border_color rounded-lg border p-4">
          <PhaseForm
            phase={{}}
            showCancel={false}
            onSubmit={(phase) => {
              onAddPhase(phase);
              setShowAddForm(false);
            }}
          />
        </div>
      ) : (
        <Card styles="px-0 py-0">
          <div
            className="border-border_color flex flex-row items-center justify-between border-b px-4
              py-4"
          >
            <div className="flex flex-row items-center gap-3">
              <div className="bg-primary/20 rounded-lg p-2">
                <ClockIcon styles="size-5 fill-primary" />
              </div>
              <div>
                <p className="font-semibold">Total duration</p>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  {sortedPhases.length} phases • {formatDuration(durationSum)}
                </p>
              </div>
            </div>
            {!showAddForm && (
              <Button
                styles="py-2 sm:px-4 px-2 text-sm"
                variant="secondary"
                buttonText={!isWindowSmall ? "Add phase" : ""}
                onClick={() => setShowAddForm(true)}
              >
                <PlusIcon styles="size-5 sm:mr-1" />
              </Button>
            )}
          </div>
          <div
            className={`${showAddForm ? "max-h-90 opacity-100" : "max-h-0 opacity-0"} overflow-hidden
              transition-[max-height,opacity] duration-300 ease-in-out`}
          >
            <div className="border-border_color border-b p-4">
              <PhaseForm
                phase={{}}
                showCancel={true}
                onSubmit={(phase) => {
                  onAddPhase(phase);
                  setShowAddForm(false);
                }}
                onCancel={() => setShowAddForm(false)}
              />
            </div>
          </div>
          <ul className="divide-border_color divide-y">
            {sortedPhases.map((phase) => (
              <li className="p-4" key={phase.sequence}>
                {editPhase === phase.sequence ? (
                  <PhaseForm
                    phase={phase}
                    isEdit={true}
                    showCancel={true}
                    onCancel={() => setEditPhase(null)}
                    onSubmit={(phase) => {
                      onUpdatePhase(phase);
                      setEditPhase(null);
                    }}
                  />
                ) : (
                  <div className="flex flex-row items-center justify-between">
                    <div className="flex flex-row items-center gap-3">
                      <div
                        className={`${phase.phase_type === "wait" ? "bg-accent/20" : "bg-secondary/20"} rounded-lg
                          p-2`}
                      >
                        {phase.phase_type === "wait" ? (
                          <HourGlassIcon styles="size-5 stroke-accent" />
                        ) : (
                          <ServicesIcon styles="size-5 text-secondary" />
                        )}
                      </div>
                      <div>
                        <p>{phase.name || `Phase ${phase.sequence}`}</p>
                        <p className="text-sm text-gray-500 dark:text-gray-400">
                          {formatDuration(phase.duration)}
                        </p>
                      </div>
                    </div>
                    <div className="flex flex-row items-center gap-4">
                      <EditIcon
                        onClick={() => setEditPhase(phase.sequence)}
                        styles="size-4 cursor-pointer"
                      />
                      <TrashBinIcon
                        onClick={() => onRemovePhase(phase.sequence)}
                        styles="size-5 cursor-pointer"
                      />
                    </div>
                  </div>
                )}
              </li>
            ))}
          </ul>
        </Card>
      )}
    </>
  );
}

function PhaseForm({ phase, showCancel, onSubmit, onCancel, isEdit }) {
  const [phaseData, setPhaseData] = useState({
    id: phase?.id || 0,
    sequence: phase?.sequence || null,
    name: phase?.name || "",
    duration: phase?.duration || "",
    duration_unit: phase?.duration_unit || "min",
    phase_type: phase?.phase_type || "active",
  });

  function updatePhaseData(data) {
    setPhaseData((prev) => ({ ...prev, ...data }));
  }

  function submitHandler(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    // shouldn't be a float but just in case
    const duration =
      phaseData.duration_unit === "hour"
        ? Math.round(phaseData.duration * 60)
        : phaseData.duration;

    onSubmit({
      id: phaseData.id,
      sequence: phaseData.sequence,
      name: phaseData.name,
      duration: duration,
      phase_type: phaseData.phase_type,
    });

    if (!isEdit) {
      setPhaseData({
        id: 0,
        sequence: null,
        name: "",
        duration: "",
        duration_unit: "min",
        phase_type: "active",
      });
    }
  }

  return (
    <form onSubmit={submitHandler} className="flex flex-col gap-4">
      <p className="pb-2 text-lg">Add new phase</p>
      <Input
        styles="p-2"
        id="phase_name"
        name="phase_name"
        type="text"
        labelText="Phase name (optional)"
        hasError={false}
        placeholder="e.g. hair wash"
        required={false}
        value={phaseData.name}
        inputData={(data) => updatePhaseData({ name: data.value })}
      />
      <div className="flex w-full flex-row items-end gap-2">
        <Input
          styles="p-2"
          id="duration"
          name="duration"
          type="number"
          min={1}
          max={phaseData.duration_unit === "hours" ? 24 : 1440}
          labelText="Duration"
          hasError={false}
          placeholder="30"
          value={phaseData.duration}
          inputData={(data) =>
            updatePhaseData({ [data.name]: Number(data.value) })
          }
        />
        <Select
          styles="w-32"
          value={phaseData.duration_unit || "min"}
          options={[
            { value: "min", label: "minutes" },
            { value: "hour", label: "hour" },
          ]}
          onSelect={(option) =>
            updatePhaseData({ duration_unit: option.value })
          }
        />
      </div>
      <div className="flex flex-row items-center gap-1">
        <Switch
          defaultValue={phaseData.phase_type === "wait"}
          onSwitch={() =>
            updatePhaseData({
              phase_type: phaseData.phase_type === "active" ? "wait" : "active",
            })
          }
        />
        <p className="pl-2">Waiting phase</p>
        <InfoIcon styles="size-4 stroke-gray-500 dark:stroke-gray-400" />
      </div>
      <div className="flex flex-row gap-2 pt-2">
        {isEdit ? (
          <Button
            styles="py-2 px-4 text-sm"
            variant="primary"
            buttonText="Save"
            type="submit"
          />
        ) : (
          <Button
            styles="py-2 px-4 text-sm"
            variant="secondary"
            buttonText="Add phase"
            type="submit"
          >
            <PlusIcon styles="size-5 mr-1" />
          </Button>
        )}
        {showCancel && (
          <Button
            styles="py-2 px-4 text-sm"
            variant="tertiary"
            buttonText="Cancel"
            type="button"
            onClick={onCancel}
          />
        )}
      </div>
    </form>
  );
}
