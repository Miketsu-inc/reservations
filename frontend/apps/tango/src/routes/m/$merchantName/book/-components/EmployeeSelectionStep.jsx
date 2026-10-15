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
    <div className="flex h-full w-full flex-col gap-10">
      <h1 className="text-3xl font-bold">Select an Employee</h1>
      <div className="flex flex-col">
        <ul className="flex flex-col gap-4">
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
      className={`bg-bg-raised hover:bg-hover flex w-full cursor-pointer
        items-center justify-between rounded-md border px-6 py-4.5
        transition-all duration-200 ${
          isSelected ? "border-primary" : "border-border"
        } `}
    >
      <div className="flex items-center gap-4">
        {noPreference ? (
          <div
            className="text-fg-primary bg-primary-subtle flex size-20 shrink-0
              items-center justify-center rounded-full"
          >
            <Icon icon={UserQuestion02Icon} styles="size-10" />
          </div>
        ) : (
          <Avatar
            styles="size-20! text-[20px]! shrink-0 rounded-full!"
            img={employee?.avatar_url}
            initials={`${employee?.first_name?.[0] ?? ""}${employee?.last_name?.[0] ?? ""}`}
          />
        )}
        <div className="flex flex-col gap-1">
          <span className="text-lg font-medium">{employee.first_name}</span>
          {noPreference ? (
            <span className="text-fg text-sm">Maximal availability</span>
          ) : (
            <div className="flex items-center gap-2">
              {employee.total_duration && (
                <div className="text-fg flex items-center gap-1 text-sm">
                  <Icon icon={Clock01Icon} styles="size-3.5" />
                  <span>{formatDuration(employee.total_duration)}</span>
                </div>
              )}
              <span className="bg-fg size-1 rounded-full"></span>
              {(employee.price || employee.price_type) && (
                <span className="text-fg text-sm">
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
            isSelected ? "border-primary bg-primary" : "border-border-strong"
          } `}
      >
        <Icon
          icon={isSelected ? Tick02Icon : PlusSignIcon}
          styles={`
            ${isSelected ? "text-white size-6" : "text-fg-muted size-5"}`}
        />
      </div>
    </li>
  );
}
