import { ArrowLeft01Icon } from "@hugeicons/core-free-icons";
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
  formatDuration,
  invalidateLocalStorageAuth,
  useToast,
} from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { Block, createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";

const DEFAULT_VALUE = "default";

const priceTypeOptions = [
  { label: "Fixed", value: "fixed" },
  { label: "From", value: "from" },
  { label: "Free", value: "free" },
];

function buildDurationOptions(defaultDuration, employees) {
  const values = new Set([defaultDuration]);

  for (let duration = 5; duration <= 240; duration += 5) {
    values.add(duration);
  }
  for (let duration = 270; duration <= 1440; duration += 30) {
    values.add(duration);
  }
  employees.forEach((employee) => {
    if (employee.duration) values.add(employee.duration);
  });

  return [
    {
      label: `${formatDuration(defaultDuration)} (Default)`,
      value: DEFAULT_VALUE,
    },
    ...Array.from(values)
      .sort((a, b) => a - b)
      .map((duration) => ({
        label: formatDuration(duration),
        value: duration,
      })),
  ];
}

function getInitialEmployees(data) {
  return data.employees.map((employee) => ({
    ...employee,
    price: employee.price ? { ...employee.price } : null,
  }));
}

async function fetchEmployeePricing(merchantId, serviceId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/services/${serviceId}/advanced-pricing-duration`,
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

function employeePricingQueryOptions(merchantId, serviceId) {
  return queryOptions({
    queryKey: [merchantId, "service", serviceId, "advanced-pricing-duration"],
    queryFn: () => fetchEmployeePricing(merchantId, serviceId),
  });
}

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/services/_topnav/advanced-pricing-duration/$id"
)({
  component: AdvancedPricingDurationPage,
  loader: async ({
    params,
    context: {
      queryClient,
      authContext: { merchantId },
    },
  }) => {
    await queryClient.ensureQueryData(
      employeePricingQueryOptions(merchantId, params.id)
    );
  },
  errorComponent: ({ error }) => <ServerError error={error.message} />,
});

function AdvancedPricingDurationPage() {
  const { id } = Route.useParams({ from: Route.id });
  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { merchantId } = useAuth();

  const { data, isLoading, isError, error } = useQuery(
    employeePricingQueryOptions(merchantId, id)
  );

  if (isLoading) return <Loading />;
  if (isError) return <ServerError error={error.message || error} />;

  return (
    <AdvancedPricingDurationForm
      key={id}
      data={data}
      merchantId={merchantId}
      serviceId={id}
      queryClient={queryClient}
    />
  );
}

function AdvancedPricingDurationForm({
  data,
  merchantId,
  serviceId,
  queryClient,
}) {
  const { showToast } = useToast();
  const [searchText, setSearchText] = useState("");
  const [serverError, setServerError] = useState();
  const [isSaving, setIsSaving] = useState(false);
  const [employees, setEmployees] = useState(() => getInitialEmployees(data));
  const [lastSavedEmployees, setLastSavedEmployees] = useState(() =>
    getInitialEmployees(data)
  );

  const durationOptions = buildDurationOptions(
    data.default_duration,
    employees
  );
  const memberPriceTypeOptions = [
    {
      label: `${capitalize(data.default_price_type)} (Default)`,
      value: DEFAULT_VALUE,
    },
    ...priceTypeOptions,
  ];
  const hasUnsavedChanges =
    JSON.stringify(employees) !== JSON.stringify(lastSavedEmployees);
  const hasOverrides = employees.some(
    (employee) =>
      employee.is_assigned &&
      (employee.duration !== null ||
        employee.price !== null ||
        employee.price_type !== null)
  );
  const filteredEmployees = employees.filter((employee) =>
    `${employee.first_name || ""} ${employee.last_name || ""}`
      .toLowerCase()
      .includes(searchText.trim().toLowerCase())
  );

  function updateEmployee(employeeId, updates) {
    setEmployees((current) =>
      current.map((employee) =>
        employee.employee_id === employeeId
          ? { ...employee, ...updates }
          : employee
      )
    );
  }

  function resetEmployee(employeeId) {
    updateEmployee(employeeId, {
      duration: null,
      price: null,
      price_type: null,
    });
  }

  function resetAll() {
    setEmployees((current) =>
      current.map((employee) => ({
        ...employee,
        ...(employee.is_assigned
          ? { duration: null, price: null, price_type: null }
          : {}),
      }))
    );
  }

  async function savePricing() {
    setIsSaving(true);
    setServerError();

    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/services/${serviceId}/advanced-pricing-duration`,
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
              duration: employee.duration,
              price: employee.price,
              price_type: employee.price_type,
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
        [merchantId, "service", serviceId, "advanced-pricing-duration"],
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
        message: "Team member pricing and duration updated successfully",
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
      shouldBlockFn={() => {
        if (!hasUnsavedChanges) return false;
        return !confirm(
          "You have unsaved changes, are you sure you want to leave?"
        );
      }}
    >
      <div className="mx-auto flex w-full max-w-6xl flex-col px-4 py-6 md:py-8">
        <div
          className="mb-8 flex flex-col gap-5 sm:flex-row sm:items-start
            sm:justify-between"
        >
          <div>
            <Link
              to={`/services/edit/${serviceId}`}
              className="text-text_color/70 hover:text-text_color mb-4 flex
                w-fit items-center gap-2 text-sm"
            >
              <Icon icon={ArrowLeft01Icon} styles="size-4" />
              Back to {data.service_name}
            </Link>
            <h1 className="text-2xl font-semibold sm:text-3xl">
              Advanced pricing and duration
            </h1>
            <p className="text-text_color/70 mt-2">
              Assign team members and set their specific pricing and duration.
            </p>
          </div>
          <Button
            type="button"
            styles="px-6 py-2 w-full sm:w-auto shrink-0"
            buttonText="Save changes"
            isLoading={isSaving}
            disabled={!hasUnsavedChanges}
            onClick={savePricing}
          />
        </div>

        <ServerError error={serverError} />

        <div
          className="border-border_color bg-layer_bg overflow-hidden rounded-xl
            border shadow-sm"
        >
          <div
            className="border-border_color bg-bg_color flex flex-col gap-3
              border-b p-4 sm:flex-row sm:items-center sm:justify-between"
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
            className="border-border_color text-text_color/70 hidden
              grid-cols-[minmax(15rem,1.35fr)_minmax(10rem,1fr)_minmax(10rem,1fr)_minmax(11rem,1fr)_4.5rem]
              gap-4 border-b px-5 py-4 text-sm font-medium lg:grid"
          >
            <span>Team member</span>
            <span>
              Duration
              <small className="block font-normal">
                Default: {formatDuration(data.default_duration)}
              </small>
            </span>
            <span>
              Price type
              <small className="block font-normal">
                Default: {capitalize(data.default_price_type)}
              </small>
            </span>
            <span>
              Price
              <small className="block font-normal">
                Default: {formatPrice(data.default_price)}
              </small>
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
            <ul className="divide-border_color divide-y">
              {filteredEmployees.map((employee) => (
                <EmployeePricingRow
                  key={employee.employee_id}
                  employee={employee}
                  defaultPrice={data.default_price}
                  defaultPriceType={data.default_price_type}
                  durationOptions={durationOptions}
                  priceTypeOptions={memberPriceTypeOptions}
                  onUpdate={(updates) =>
                    updateEmployee(employee.employee_id, updates)
                  }
                  onReset={() => resetEmployee(employee.employee_id)}
                />
              ))}
            </ul>
          )}
        </div>
      </div>
    </Block>
  );
}

function EmployeePricingRow({
  employee,
  defaultPrice,
  defaultPriceType,
  durationOptions,
  priceTypeOptions,
  onUpdate,
  onReset,
}) {
  const hasOverride =
    employee.duration !== null ||
    employee.price !== null ||
    employee.price_type !== null;
  const effectivePriceType = employee.price_type || defaultPriceType;
  const currency = employee.price?.currency || defaultPrice?.currency || "HUF";
  const initials = `${employee.first_name?.[0] || ""}${employee.last_name?.[0] || ""}`;

  return (
    <li
      className={`grid grid-cols-1 gap-4 p-5 transition-colors
        ${employee.is_assigned ? "" : "bg-bg_color/60"}
        lg:grid-cols-[minmax(15rem,1.35fr)_minmax(10rem,1fr)_minmax(10rem,1fr)_minmax(11rem,1fr)_4.5rem]
        lg:items-center`}
    >
      <div className="flex min-w-0 items-center gap-3">
        <CheckBox
          checked={employee.is_assigned}
          onChange={(event) =>
            onUpdate({
              is_assigned: event.target.checked,
              ...(!event.target.checked
                ? { duration: null, price: null, price_type: null }
                : {}),
            })
          }
          aria-label={`${employee.is_assigned ? "Remove" : "Assign"} ${
            [employee.first_name, employee.last_name]
              .filter(Boolean)
              .join(" ") || "team member"
          }`}
          styles="shrink-0"
        />
        <Avatar
          styles={`size-11! shrink-0 text-sm ${
            employee.is_assigned ? "" : "opacity-50"
          }`}
          initials={initials || "?"}
        />
        <div className={`min-w-0 ${employee.is_assigned ? "" : "opacity-50"}`}>
          <p className="truncate font-medium">
            {[employee.first_name, employee.last_name]
              .filter(Boolean)
              .join(" ") || "Unnamed team member"}
          </p>
          <p className="text-text_color/60 text-sm">
            {employee.is_assigned
              ? capitalize(employee.role)
              : `${capitalize(employee.role)} · Not assigned`}
          </p>
        </div>
      </div>

      <div className={employee.is_assigned ? "" : "opacity-50"}>
        <span className="text-text_color/60 mb-1 block text-xs lg:hidden">
          Duration
        </span>
        <Select
          options={durationOptions}
          value={employee.duration ?? DEFAULT_VALUE}
          disabled={!employee.is_assigned}
          onSelect={(option) =>
            onUpdate({
              duration: option.value === DEFAULT_VALUE ? null : option.value,
            })
          }
        />
      </div>

      <div className={employee.is_assigned ? "" : "opacity-50"}>
        <span className="text-text_color/60 mb-1 block text-xs lg:hidden">
          Price type
        </span>
        <Select
          options={priceTypeOptions}
          value={employee.price_type ?? DEFAULT_VALUE}
          disabled={!employee.is_assigned}
          onSelect={(option) => {
            const priceType =
              option.value === DEFAULT_VALUE ? null : option.value;
            onUpdate({
              price_type: priceType,
              ...(priceType === "free"
                ? { price: { number: "0", currency } }
                : {}),
            });
          }}
        />
      </div>

      <div className={employee.is_assigned ? "" : "opacity-50"}>
        <span className="text-text_color/60 mb-1 block text-xs lg:hidden">
          Price
        </span>
        <Input
          id={`employee-${employee.employee_id}-price`}
          name={`employee-${employee.employee_id}-price`}
          aria-label={`Price for ${employee.first_name || "team member"}`}
          type="number"
          min={0}
          max={1000000}
          required={false}
          value={employee.price?.number ?? defaultPrice?.number ?? ""}
          disabled={!employee.is_assigned || effectivePriceType === "free"}
          inputData={({ value }) =>
            onUpdate({
              price: value === "" ? null : { number: value, currency },
            })
          }
        >
          <span
            className="border-input_border_color bg-bg_color rounded-r-lg border
              px-3 py-2 text-sm"
          >
            {currency}
          </span>
        </Input>
      </div>

      <button
        type="button"
        onClick={onReset}
        disabled={!employee.is_assigned || !hasOverride}
        className="text-primary hover:bg-hvr_gray w-fit cursor-pointer
          rounded-lg px-2 py-2 text-sm font-medium disabled:cursor-default
          disabled:opacity-30 lg:justify-self-end"
      >
        Reset
      </button>
    </li>
  );
}

function capitalize(value = "") {
  return value ? value[0].toUpperCase() + value.slice(1) : "";
}

function formatPrice(price) {
  if (!price) return "Free";
  return `${price.number} ${price.currency}`;
}
