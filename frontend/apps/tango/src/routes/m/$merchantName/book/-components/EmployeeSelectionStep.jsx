import {
  Clock01Icon,
  PlusSignIcon,
  Tick02Icon,
  UserQuestion02Icon,
} from "@hugeicons/core-free-icons";
import { Avatar, Icon, ServerError } from "@reservations/components";
import {
  activeTeamQueryOptions,
  formatDuration,
  getDisplayPrice,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { StepContentSkeleton } from "./StepContentSkeleton";

export default function EmployeeSelectionStep({
  merchantName,
  _locationId,
  serviceId,
  _serviceType,
  onSelect,
  onAutoSkip,
}) {
  const {
    data: employees,
    isLoading,
    isError,
    error,
  } = useQuery({ ...activeTeamQueryOptions(merchantName, serviceId) });

  const [selectedEmployee, setSelectedEmployee] = useState();

  const noPrefEmployee = { id: "no-pref", first_name: "No preference" };

  function handleEmployeeSelect(emp) {
    if (selectedEmployee?.id === emp.id) {
      setSelectedEmployee(null);
      onSelect(null);
    } else {
      setSelectedEmployee(emp);
      onSelect(emp);
    }
  }

  useEffect(() => {
    if (!isLoading && employees?.length === 1) {
      onAutoSkip(employees[0]);
    }
  }, [employees, isLoading, onAutoSkip]);

  if (isError) {
    return <ServerError error={error.message} />;
  }

  if (isLoading) {
    return <StepContentSkeleton />;
  }

  return (
    <div className="flex h-full w-full flex-col gap-6">
      <h1 className="text-2xl font-bold">Select an Employee</h1>
      <div className="flex flex-col">
        <ul className="flex flex-col gap-3">
          <EmployeeItem
            employee={noPrefEmployee}
            isSelected={selectedEmployee?.id === "no-pref"}
            onSelect={handleEmployeeSelect}
            noPreference={true}
          />
          {employees?.map((employee) => {
            const isSelected = selectedEmployee?.id === employee?.id;

            return (
              <EmployeeItem
                key={employee.id}
                employee={employee}
                isSelected={isSelected}
                onSelect={handleEmployeeSelect}
              />
            );
          })}
        </ul>
      </div>
    </div>
  );
}

export function EmployeeItem({
  employee,
  isSelected,
  onSelect,
  noPreference = false,
}) {
  return (
    <li
      role="radio"
      aria-checked={isSelected}
      onClick={() => onSelect(employee)}
      className={`group bg-bg-raised shadow-card flex w-full cursor-pointer
        items-center justify-between rounded-xl border px-4 py-3.5 text-left ${
          isSelected
            ? "border-primary ring-primary-focus ring-2"
            : "border-border hover:border-border-strong"
        } `}
    >
      <div className="flex items-center gap-4">
        {noPreference ? (
          <div
            className="text-fg-primary bg-primary-subtle flex size-16 shrink-0
              items-center justify-center rounded-full"
          >
            <Icon icon={UserQuestion02Icon} styles="size-8" />
          </div>
        ) : (
          <Avatar
            styles="size-16! text-[16px]! shrink-0"
            img={employee?.avatar_url}
            initials={`${employee?.first_name?.[0] ?? ""}${employee?.last_name?.[0] ?? ""}`}
          />
        )}
        <div className="flex flex-col gap-1">
          <span className="font-semibold">{employee.first_name}</span>
          {noPreference ? (
            <span className="text-fg-muted text-sm">Maximal availability</span>
          ) : (
            <div className="flex items-center gap-2">
              {employee.total_duration && (
                <div className="text-fg-muted flex items-center gap-1 text-sm">
                  <Icon icon={Clock01Icon} styles="size-3.5" />
                  <span>{formatDuration(employee.total_duration)}</span>
                </div>
              )}
              <span className="bg-fg-muted size-1 rounded-full"></span>
              {(employee.price || employee.price_type) && (
                <span className="text-fg text-sm font-medium">
                  {getDisplayPrice(employee.price, employee.price_type)}
                </span>
              )}
            </div>
          )}
        </div>
      </div>
      <div
        className={`flex size-8 shrink-0 items-center justify-center
          rounded-full border transition-colors ${
            isSelected
              ? "border-primary bg-primary"
              : "border-border-strong group-hover:border-border-hover"
          } `}
      >
        <Icon
          icon={isSelected ? Tick02Icon : PlusSignIcon}
          styles={`
            ${isSelected ? "text-fg-on-primary size-5" : "text-fg-muted size-5"}`}
        />
      </div>
    </li>
  );
}
