import { ArrowLeft01Icon, Person } from "@hugeicons/core-free-icons";
import {
  Avatar,
  CloseButton,
  Icon,
  Loading,
  ResponsiveDialog,
  ResponsiveDialogClose,
  ResponsiveDialogContent,
  ServerError,
} from "@reservations/components";
import { activeTeamQueryOptions, useWindowSize } from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { EmployeeItem } from "./EmployeeSelectionStep";

export default function EmployeePicker({
  merchantName,
  serviceId,
  employeeId,
  employee,
  onSelectEmployee,
}) {
  const [isOpen, setIsOpen] = useState(false);
  const { isWindowSmall } = useWindowSize();

  const {
    data: employees,
    isLoading,
    isError,
    error,
  } = useQuery({
    ...activeTeamQueryOptions(merchantName, serviceId),
    enabled: isOpen && Boolean(merchantName && serviceId),
  });

  const noPrefEmployee = { id: "no-pref", first_name: "No preference" };
  const isSingleEmployee = employees?.length === 1;

  const currentEmployee = employees?.find(
    (emp) => String(emp.id) === String(employeeId)
  );

  const isNoPref = employeeId === "no-pref";
  const empFirstName = currentEmployee?.first_name ?? employee?.first_name;
  const empLastName = currentEmployee?.last_name ?? employee?.last_name;
  const avatarUrl = currentEmployee?.avatar_url ?? employee?.avatar_url;

  function handleEmployeeSelect(emp) {
    if (String(employeeId) !== String(emp.id)) {
      onSelectEmployee?.(emp);
    }
    setIsOpen(false);
  }

  return (
    <>
      <button
        type="button"
        onClick={() => setIsOpen(true)}
        className="bg-layer_bg border-border_color focus-visible:ring-primary
          flex w-fit cursor-pointer items-center gap-2 rounded-full border
          py-1.5 pr-3 pl-2 transition-all duration-200 hover:bg-gray-50
          focus-visible:ring-2 focus-visible:outline-none
          dark:hover:bg-gray-200/5"
      >
        {isNoPref ? (
          <>
            <div
              className="bg-primary/80 flex size-8 items-center justify-center
                rounded-full"
            >
              <Icon icon={Person} styles="size-5 text-white" />
            </div>
            <span className="text-sm font-medium">No Preference</span>
          </>
        ) : (
          <>
            <Avatar
              styles="size-8! text-[11px]! shrink-0 rounded-full!"
              img={avatarUrl}
              initials={`${empFirstName?.[0]}${empLastName?.[0]}`}
            />
            <span className="text-sm font-medium">
              {`${empFirstName} ${empLastName}`}
            </span>
          </>
        )}
        <Icon
          icon={ArrowLeft01Icon}
          styles={`text-text_color/60 -rotate-90 size-5 shrink-0
            transition-transform duration-200 ${isOpen ? "rotate-90" : ""}`}
        />
      </button>

      <ResponsiveDialog open={isOpen} onOpenChange={(open) => setIsOpen(open)}>
        <ResponsiveDialogContent
          styles="w-full lg:p-6"
          popUpStyles="h-[calc(80vh+3rem)]!"
        >
          <div className="flex w-full justify-end">
            <ResponsiveDialogClose asChild>
              <CloseButton styles="hidden lg:block" />
            </ResponsiveDialogClose>
          </div>
          <div
            className="flex h-full w-full flex-col gap-8 p-3 lg:h-auto lg:px-6"
          >
            <div className="flex flex-col gap-1">
              <h2 className="text-2xl font-bold">Select an Employee</h2>
              {isSingleEmployee && (
                <p className="text-text_color/70 text-sm">
                  Only this employee provides this service.
                </p>
              )}
            </div>

            {isError ? (
              <ServerError
                error={error?.message || "Failed to load team members"}
              />
            ) : isLoading ? (
              <div className="flex w-full flex-col lg:max-h-[60vh] lg:w-140">
                <Loading />
              </div>
            ) : (
              <div
                className={`scrollbar-thin overflow-y-auto pr-1 pb-8
                  dark:scheme-dark
                  ${isWindowSmall ? "flex-1" : "max-h-[60vh] w-145"}`}
              >
                <ul className="flex flex-col gap-4">
                  {!isSingleEmployee && (
                    <EmployeeItem
                      employee={noPrefEmployee}
                      isSelected={employeeId === "no-pref"}
                      onSelect={handleEmployeeSelect}
                      noPreference={true}
                    />
                  )}
                  {employees?.map((emp) => {
                    const isSelected =
                      String(employeeId) === String(emp.id) || isSingleEmployee;
                    return (
                      <EmployeeItem
                        key={emp.id}
                        employee={emp}
                        isSelected={isSelected}
                        onSelect={handleEmployeeSelect}
                      />
                    );
                  })}
                </ul>
              </div>
            )}
          </div>
        </ResponsiveDialogContent>
      </ResponsiveDialog>
    </>
  );
}
