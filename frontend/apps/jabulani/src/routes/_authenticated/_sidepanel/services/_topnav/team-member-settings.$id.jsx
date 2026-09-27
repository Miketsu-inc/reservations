import { ArrowDown01Icon, ArrowLeft01Icon } from "@hugeicons/core-free-icons";
import {
  Avatar,
  Button,
  CheckBox,
  Icon,
  Input,
  Loading,
  SearchInput,
  Select,
  ServerError,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  BUFFER_TIME_OPTIONS,
  formatDuration,
  invalidateLocalStorageAuth,
  useToast,
} from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { Block, createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";

const durationUnitOptions = [
  { value: "min", label: "minutes" },
  { value: "hour", label: "hours" },
];

const priceTypeOptions = [
  { label: "fixed", value: "fixed" },
  { label: "from", value: "from" },
  { label: "free", value: "free" },
];

function toMinutes(duration, durationUnit) {
  if (duration === "") return null;

  const value = Number(duration);
  return Math.round(durationUnit === "hour" ? value * 60 : value);
}

function fromMinutes(duration, durationUnit) {
  if (duration === null || duration === undefined) return "";

  return durationUnit === "hour" ? duration / 60 : duration;
}

function getInitialEmployees(data) {
  return data.employees.map((employee) => ({ ...employee }));
}

function getDurationSum(employee, phases) {
  const overrides = new Map(
    employee.phase_overrides.map((phase) => [
      phase.service_phase_id,
      phase.duration,
    ])
  );

  return phases.reduce(
    (total, phase) => total + (overrides.get(phase.id) ?? phase.duration),
    0
  );
}

function hasEmployeeOverrides(employee) {
  return (
    employee.price !== null ||
    employee.price_type !== null ||
    employee.min_participants !== null ||
    employee.max_participants !== null ||
    employee.buffer_time !== null ||
    employee.phase_overrides.length > 0
  );
}

function resetOverrides(employee) {
  return {
    ...employee,
    price: null,
    price_type: null,
    min_participants: null,
    max_participants: null,
    buffer_time: null,
    phase_overrides: [],
  };
}

async function fetchTeamMemberSettings(merchantId, serviceId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/services/${serviceId}/team-member-settings`,
    {
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    }
  );

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  }

  return result.data;
}

function teamMemberSettingsQueryOptions(merchantId, serviceId) {
  return queryOptions({
    queryKey: [merchantId, "service", serviceId, "team-member-settings"],
    queryFn: () => fetchTeamMemberSettings(merchantId, serviceId),
  });
}

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/services/_topnav/team-member-settings/$id"
)({
  component: TeamMemberSettingsPage,
  loader: async ({
    params,
    context: {
      queryClient,
      authContext: { merchantId },
    },
  }) => {
    await queryClient.ensureQueryData(
      teamMemberSettingsQueryOptions(merchantId, params.id)
    );
  },
  errorComponent: ({ error }) => <ServerError error={error.message} />,
});

function TeamMemberSettingsPage() {
  const { id } = Route.useParams({ from: Route.id });
  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { merchantId } = useAuth();

  const { data, isLoading, isError, error } = useQuery(
    teamMemberSettingsQueryOptions(merchantId, id)
  );

  if (isLoading) return <Loading />;
  if (isError) return <ServerError error={error.message || error} />;

  return (
    <TeamMemberSettingsForm
      key={id}
      data={data}
      merchantId={merchantId}
      serviceId={id}
      queryClient={queryClient}
    />
  );
}

function TeamMemberSettingsForm({ data, merchantId, serviceId, queryClient }) {
  const { showToast } = useToast();
  const [searchText, setSearchText] = useState("");
  const [serverError, setServerError] = useState();
  const [isSaving, setIsSaving] = useState(false);
  const [expandedEmployeeIds, setExpandedEmployeeIds] = useState(
    () => new Set()
  );
  const [employees, setEmployees] = useState(() => getInitialEmployees(data));
  const [lastSavedEmployees, setLastSavedEmployees] = useState(() =>
    getInitialEmployees(data)
  );

  const defaultBufferTime = data.default_buffer_time ?? 0;
  const defaultBufferTimeLabel =
    BUFFER_TIME_OPTIONS.find((option) => option.value === defaultBufferTime)
      ?.label ?? formatDuration(defaultBufferTime);
  const isGroupService = data.booking_type !== "appointment";

  const hasUnsavedChanges =
    JSON.stringify(employees) !== JSON.stringify(lastSavedEmployees);
  const hasOverrides = employees.some(
    (employee) => employee.is_assigned && hasEmployeeOverrides(employee)
  );

  const filteredEmployees = employees.filter((employee) =>
    `${employee.first_name} ${employee.last_name}`
      .toLowerCase()
      .includes(searchText.trim().toLowerCase())
  );

  function updateEmployee(employeeId, updater) {
    setEmployees((current) =>
      current.map((employee) => {
        if (employee.employee_id !== employeeId) return employee;
        return typeof updater === "function"
          ? updater(employee)
          : { ...employee, ...updater };
      })
    );
  }

  function updatePhase(employeeId, phaseId, duration) {
    updateEmployee(employeeId, (employee) => {
      const otherOverrides = employee.phase_overrides.filter(
        (phase) => phase.service_phase_id !== phaseId
      );
      return {
        ...employee,
        phase_overrides:
          duration === null
            ? otherOverrides
            : [...otherOverrides, { service_phase_id: phaseId, duration }].sort(
                (a, b) => a.service_phase_id - b.service_phase_id
              ),
      };
    });
  }

  function resetAll() {
    setEmployees((current) =>
      current.map((employee) =>
        employee.is_assigned ? resetOverrides(employee) : employee
      )
    );
  }

  function validateSettings() {
    if (!employees.some((employee) => employee.is_assigned)) {
      return "Assign at least one team member to the service.";
    }

    if (isGroupService) {
      for (const employee of employees.filter(
        (current) => current.is_assigned
      )) {
        const minParticipants =
          employee.min_participants ?? data.default_min_participants;
        const maxParticipants =
          employee.max_participants ?? data.default_max_participants;
        if (minParticipants > maxParticipants) {
          return `Minimum participants cannot exceed maximum participants for ${employee.first_name} ${employee.last_name}.`;
        }
      }
    }

    return null;
  }

  async function saveSettings() {
    const validationError = validateSettings();
    if (validationError) {
      setServerError(validationError);
      return;
    }

    setIsSaving(true);
    setServerError();

    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/services/${serviceId}/team-member-settings`,
        {
          method: "PUT",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify({
            employees: employees.map((employee) => ({
              employee_id: employee.employee_id,
              is_assigned: employee.is_assigned,
              price: employee.price,
              price_type: employee.price_type,
              min_participants: isGroupService
                ? employee.min_participants
                : null,
              max_participants: isGroupService
                ? employee.max_participants
                : null,
              buffer_time: employee.buffer_time,
              phase_overrides: employee.phase_overrides,
            })),
          }),
        }
      );

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        throw new Error(result.error.message);
      }

      queryClient.setQueryData(
        teamMemberSettingsQueryOptions(merchantId, serviceId),
        (current) => ({ ...current, employees })
      );
      queryClient.setQueryData([merchantId, "service", serviceId], (current) =>
        current
          ? {
              ...current,
              employee_ids: employees
                .filter((employee) => employee.is_assigned)
                .map((employee) => employee.employee_id),
            }
          : current
      );
      setLastSavedEmployees(employees);
      showToast({
        message: "Team member settings updated successfully",
        variant: "success",
      });
    } catch (err) {
      setServerError(err.message);
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <Block
      shouldBlockFn={() =>
        hasUnsavedChanges &&
        !confirm("You have unsaved changes, are you sure you want to leave?")
      }
    >
      <div className="mx-auto flex w-full max-w-6xl flex-col px-4 py-6 md:py-8">
        <div className="mb-5 flex items-center justify-between gap-3">
          <Link
            to={`/services/edit/${serviceId}`}
            className="hover:bg-hvr_gray flex min-w-0 items-center gap-2
              rounded-lg py-2 pr-2"
          >
            <Icon icon={ArrowLeft01Icon} styles="size-6 shrink-0" />
            <span className="truncate">Back to {data.service_name}</span>
          </Link>
          <Button
            type="button"
            styles="px-6 py-2 shrink-0"
            buttonText="Save"
            isLoading={isSaving}
            disabled={!hasUnsavedChanges}
            onClick={saveSettings}
          />
        </div>
        <div className="mb-8">
          <h1 className="text-2xl font-semibold sm:text-3xl">
            Team member settings
          </h1>
          <p className="text-text_color/70 mt-2 max-w-2xl">
            Manage assignments and customize pricing, service phases, capacity,
            and scheduling for each team member.
          </p>
        </div>

        <ServerError error={serverError} />

        <div className="overflow-hidden">
          <div
            className="bg-layer_bg flex flex-col gap-3 rounded-xl p-4
              sm:flex-row sm:items-center sm:justify-between"
          >
            <SearchInput
              searchText={searchText}
              onChange={setSearchText}
              placeholder="Search team members"
              styles="w-full! sm:w-72!"
            />
            <Button
              type="button"
              variant="tertiary"
              styles="px-4 py-2 w-full sm:w-auto"
              buttonText="Reset all"
              disabled={!hasOverrides}
              onClick={resetAll}
            />
          </div>
          <div
            className="border-border_color hidden
              grid-cols-[minmax(12.5rem,1.3fr)_minmax(12.5rem,1fr)_minmax(10rem,1fr)_minmax(11rem,1fr)_8rem]
              gap-4 border-b px-5 pt-8 pb-4 text-sm xl:grid"
          >
            <span className="text-text_color flex items-center font-semibold">
              Team member
            </span>
            <span>
              <span className="text-text_color font-semibold">Duration</span>
              <span className="text-text_color/60 mt-1 block text-sm">
                Default: {formatDuration(data.default_duration)}
              </span>
            </span>
            <span>
              <span className="text-text_color font-semibold">Price type</span>
              <span className="text-text_color/60 mt-1 block text-sm">
                Default: {data.default_price_type}
              </span>
            </span>
            <span>
              <span className="text-text_color font-semibold">Price</span>
              <span className="text-text_color/60 mt-1 block text-sm">
                Default: {formatPrice(data.default_price)}
              </span>
            </span>
            <span />
          </div>

          {filteredEmployees.length === 0 ? (
            <div className="text-text_color/60 px-6 py-14 text-center">
              {employees.length === 0
                ? "No active team members are available."
                : "No team members match your search."}
            </div>
          ) : (
            <ul>
              {filteredEmployees.map((employee) => (
                <EmployeeSettingsRow
                  key={employee.employee_id}
                  employee={employee}
                  data={data}
                  isGroupService={isGroupService}
                  isExpanded={expandedEmployeeIds.has(employee.employee_id)}
                  defaultBufferTime={defaultBufferTime}
                  defaultBufferTimeLabel={defaultBufferTimeLabel}
                  onToggleExpanded={() =>
                    setExpandedEmployeeIds((current) => {
                      const next = new Set(current);
                      next.has(employee.employee_id)
                        ? next.delete(employee.employee_id)
                        : next.add(employee.employee_id);
                      return next;
                    })
                  }
                  onUpdate={(updates) =>
                    updateEmployee(employee.employee_id, updates)
                  }
                  onUpdatePhase={(phaseId, duration) =>
                    updatePhase(employee.employee_id, phaseId, duration)
                  }
                  onReset={() =>
                    updateEmployee(employee.employee_id, resetOverrides)
                  }
                />
              ))}
            </ul>
          )}
        </div>
      </div>
    </Block>
  );
}

function EmployeeSettingsRow({
  employee,
  data,
  isGroupService,
  isExpanded,
  defaultBufferTime,
  defaultBufferTimeLabel,
  onToggleExpanded,
  onUpdate,
  onUpdatePhase,
  onReset,
}) {
  const [durationUnits, setDurationUnits] = useState({});
  const hasOverrides = hasEmployeeOverrides(employee);
  const effectivePriceType = employee.price_type || data.default_price_type;
  const currency =
    employee.price?.currency ||
    data.default_price?.currency ||
    data.currency_code;
  const initials = `${employee.first_name?.[0]}${employee.last_name?.[0]}`;
  const name = `${employee.first_name} ${employee.last_name}`;
  const durationSum = getDurationSum(employee, data.phases);
  const isSingleDuration = isGroupService || data.phases.length === 1;
  const singlePhase = data.phases[0];

  function getDurationUnit(phaseId) {
    return durationUnits[phaseId] || "min";
  }

  function setDurationUnit(phaseId, durationUnit) {
    setDurationUnits((current) => ({
      ...current,
      [phaseId]: durationUnit,
    }));
  }

  return (
    <li className={employee.is_assigned ? "" : "bg-bg_color/60"}>
      <div
        className="grid grid-cols-1 gap-4 p-5
          xl:grid-cols-[minmax(12.5rem,1.3fr)_minmax(12.5rem,1fr)_minmax(10rem,1fr)_minmax(11rem,1fr)_8rem]
          xl:items-center"
      >
        <div className="flex min-w-0 items-center gap-3">
          <CheckBox
            checked={employee.is_assigned}
            onChange={(event) =>
              onUpdate(
                event.target.checked
                  ? { is_assigned: true }
                  : resetOverrides({ ...employee, is_assigned: false })
              )
            }
            aria-label={`${employee.is_assigned ? "Remove" : "Assign"} ${name}`}
            styles="shrink-0"
          />
          <Avatar
            styles={`size-11! rounded-full! shrink-0 text-sm ${
              employee.is_assigned ? "" : "opacity-50"
            }`}
            initials={initials}
          />
          <div
            className={`min-w-0 ${employee.is_assigned ? "" : "opacity-50"}`}
          >
            <p className="truncate">{name}</p>
            <p className="text-text_color/60 text-sm">
              {employee.is_assigned
                ? employee.role
                : `${employee.role} · not assigned`}
            </p>
          </div>
        </div>

        <div className={employee.is_assigned ? "" : "opacity-50"}>
          <span className="flex items-center gap-1 pb-1 text-sm xl:hidden">
            Duration (default: {formatDuration(data.default_duration)})
          </span>
          {isSingleDuration && singlePhase ? (
            <PhaseDurationInput
              styles="w-full"
              unitStyles="w-28!"
              employee={employee}
              phase={singlePhase}
              durationUnit={getDurationUnit(singlePhase.id)}
              disabled={!employee.is_assigned}
              onDurationUnitChange={(durationUnit) =>
                setDurationUnit(singlePhase.id, durationUnit)
              }
              onUpdate={(duration) => onUpdatePhase(singlePhase.id, duration)}
            />
          ) : (
            <Input
              id={`employee-${employee.employee_id}-total-duration`}
              name={`employee-${employee.employee_id}-total-duration`}
              aria-label={`Calculated duration for ${name}`}
              type="text"
              required={false}
              value={formatDuration(durationSum)}
              disabled
            />
          )}
        </div>

        <div className={employee.is_assigned ? "" : "opacity-50"}>
          <span className="flex items-center gap-1 pb-1 text-sm xl:hidden">
            Price type (default: {data.default_price_type})
          </span>
          <Select
            options={priceTypeOptions}
            value={effectivePriceType}
            disabled={!employee.is_assigned}
            onSelect={(option) => {
              const priceType =
                option.value === data.default_price_type ? null : option.value;
              onUpdate({
                price_type: priceType,
                ...(option.value === "free"
                  ? {
                      price: priceType ? { number: "0", currency } : null,
                    }
                  : employee.price_type === "free" &&
                      employee.price?.number === "0"
                    ? { price: null }
                    : {}),
              });
            }}
          />
        </div>

        <div className={employee.is_assigned ? "" : "opacity-50"}>
          <span className="flex items-center gap-1 pb-1 text-sm xl:hidden">
            Price (default: {formatPrice(data.default_price)})
          </span>
          <Input
            id={`employee-${employee.employee_id}-price`}
            name={`employee-${employee.employee_id}-price`}
            aria-label={`Price for ${name}`}
            type="number"
            min={0}
            max={1000000}
            required={false}
            placeholder={data.default_price?.number ?? ""}
            value={employee.price?.number ?? ""}
            disabled={!employee.is_assigned || effectivePriceType === "free"}
            inputData={({ value }) =>
              onUpdate({
                price: value === "" ? null : { number: value, currency },
              })
            }
          >
            <span
              className="border-input_border_color
                peer-disabled:text-text_color/70
                peer-disabled:border-input_border_color/60 rounded-r-lg border
                px-4 py-2 peer-disabled:bg-gray-200/60
                peer-disabled:dark:bg-gray-700/20"
            >
              {currency}
            </span>
          </Input>
        </div>

        <div className="flex items-center gap-2 xl:justify-self-end">
          <button
            type="button"
            onClick={onToggleExpanded}
            disabled={!employee.is_assigned}
            aria-expanded={isExpanded}
            aria-label={`${isExpanded ? "Hide" : "Show"} more settings for ${name}`}
            className="hover:bg-hvr_gray flex cursor-pointer items-center gap-1
              rounded-lg px-2 py-1.5 text-sm disabled:cursor-default
              disabled:opacity-30 disabled:hover:bg-transparent"
          >
            <span>{isExpanded ? "Less" : "More"}</span>
            <Icon
              icon={ArrowDown01Icon}
              styles={`size-5 transition-transform ${
                isExpanded ? "rotate-180" : ""
              }`}
            />
          </button>
          <button
            type="button"
            onClick={onReset}
            disabled={!employee.is_assigned || !hasOverrides}
            className="text-primary hover:bg-primary/10 cursor-pointer
              rounded-lg px-2 py-1.5 text-sm disabled:cursor-default
              disabled:opacity-30 disabled:hover:bg-transparent"
          >
            Reset
          </button>
        </div>
      </div>

      {isExpanded && employee.is_assigned && (
        <div
          className="grid grid-cols-1 px-5 py-6
            xl:grid-cols-[minmax(12.5rem,1.3fr)_minmax(12.5rem,1fr)_minmax(10rem,1fr)_minmax(11rem,1fr)_8rem]
            xl:gap-x-4"
        >
          <div className="flex flex-col gap-8 xl:col-start-2 xl:col-end-6">
            {data.phases.length > 1 && (
              <section>
                <h3 className="mb-3 font-semibold">Duration</h3>
                <div className="flex flex-col gap-4 pt-4">
                  {data.phases.map((phase) => (
                    <div key={phase.id} className="flex flex-row gap-4">
                      <PhaseDurationInput
                        unitStyles="w-32! xl:w-52!"
                        employee={employee}
                        phase={phase}
                        durationUnit={getDurationUnit(phase.id)}
                        onDurationUnitChange={(unit) =>
                          setDurationUnit(phase.id, unit)
                        }
                        onUpdate={(duration) =>
                          onUpdatePhase(phase.id, duration)
                        }
                      />
                      <PhaseTypeBadge phaseType={phase.phase_type} />
                    </div>
                  ))}
                </div>
              </section>
            )}

            <section>
              <h3 className="font-semibold">Booking settings</h3>
              <div
                className={`grid gap-4 pt-4 ${
                  isGroupService ? "sm:grid-cols-3" : "sm:grid-cols-1"
                }`}
              >
                {isGroupService && (
                  <>
                    <Input
                      id={`employee-${employee.employee_id}-min-participants`}
                      name={`employee-${employee.employee_id}-min-participants`}
                      type="number"
                      min={1}
                      required={false}
                      labelText={`Minimum participants (default: ${data.default_min_participants})`}
                      placeholder={data.default_min_participants}
                      value={employee.min_participants ?? ""}
                      inputData={({ value }) =>
                        onUpdate({
                          min_participants: value === "" ? null : Number(value),
                        })
                      }
                    />
                    <Input
                      id={`employee-${employee.employee_id}-max-participants`}
                      name={`employee-${employee.employee_id}-max-participants`}
                      type="number"
                      min={1}
                      required={false}
                      labelText={`Maximum participants (default: ${data.default_max_participants})`}
                      placeholder={data.default_max_participants}
                      value={employee.max_participants ?? ""}
                      inputData={({ value }) =>
                        onUpdate({
                          max_participants: value === "" ? null : Number(value),
                        })
                      }
                    />
                  </>
                )}
                <Select
                  options={BUFFER_TIME_OPTIONS}
                  labelText={`Buffer time (default: ${defaultBufferTimeLabel})`}
                  required={false}
                  value={employee.buffer_time ?? defaultBufferTime}
                  onSelect={(option) =>
                    onUpdate({
                      buffer_time:
                        option.value === defaultBufferTime
                          ? null
                          : option.value,
                    })
                  }
                />
              </div>
            </section>
          </div>
        </div>
      )}
    </li>
  );
}

function PhaseTypeBadge({ phaseType }) {
  const isActive = phaseType === "active";

  return (
    <div className="flex h-full w-20 justify-center">
      <div
        className={`h-fit gap-2 rounded-xl px-2 py-2 text-sm ${
          isActive
            ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
            : "bg-amber-500/10 text-amber-700 dark:text-amber-400"
          }`}
      >
        {isActive ? "Active" : "Waiting"}
      </div>
    </div>
  );
}

function PhaseDurationInput({
  employee,
  phase,
  durationUnit,
  disabled = false,
  styles = "",
  unitStyles = "",
  onDurationUnitChange,
  onUpdate,
}) {
  const override = employee.phase_overrides.find(
    (item) => item.service_phase_id === phase.id
  );

  return (
    <Input
      styles={`w-full ${styles}`}
      id={`employee-${employee.employee_id}-phase-${phase.id}-duration`}
      name={`employee-${employee.employee_id}-phase-${phase.id}-duration`}
      type="number"
      min={1}
      max={durationUnit === "hour" ? 24 : 1440}
      required={false}
      disabled={disabled}
      placeholder={fromMinutes(phase.duration, durationUnit)}
      value={fromMinutes(override?.duration, durationUnit)}
      inputData={({ value }) => onUpdate(toMinutes(value, durationUnit))}
    >
      <Select
        styles={`rounded-l-none ${unitStyles}`}
        value={durationUnit}
        options={durationUnitOptions}
        disabled={disabled}
        onSelect={(option) => onDurationUnitChange(option.value)}
      />
    </Input>
  );
}

function formatPrice(price) {
  if (!price) return "Free";
  return `${price.number} ${price.currency}`;
}
