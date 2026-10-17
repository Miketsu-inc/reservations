import { StarIcon } from "@hugeicons/core-free-icons";
import { Avatar, Icon, Loading, ServerError } from "@reservations/components";
import { activeTeamQueryOptions, useWindowSize } from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";

export default function TeamSection({ merchantName }) {
  const { isWindowSmall } = useWindowSize();

  const {
    data: employees,
    isLoading,
    isError,
    error,
  } = useQuery({ ...activeTeamQueryOptions(merchantName) });

  if (isError) {
    return <ServerError error={error.message} />;
  }

  if (isLoading) {
    return <Loading />;
  }

  return (
    <div className="flex flex-col">
      <div
        className="flex w-full scrollbar-thin gap-8 overflow-x-auto lg:flex-wrap
          lg:gap-x-16 lg:gap-y-12 dark:scheme-dark"
      >
        {employees.map((employee) => (
          <div
            key={employee.id}
            className="flex w-32 shrink-0 flex-col items-center gap-4
              text-center"
          >
            <Avatar
              styles={`shrink-0 font-semibold
              ${isWindowSmall ? "size-20 text-lg!" : "size-22 text-xl!"}`}
              initials={`${employee.first_name.charAt(0)}${employee.last_name.charAt(0)}`}
            />

            <div className="flex flex-col items-center justify-center gap-1.5">
              <span className="leading-tight font-semibold">
                {employee.first_name} {employee.last_name.charAt(0)}.
              </span>

              <div className="flex items-center gap-1.5">
                <Icon
                  icon={StarIcon}
                  styles="fill-warning size-4 text-fg-warning"
                />
                <div className="flex items-center gap-1 text-sm">
                  <span className="font-medium">4.6</span>
                  <span className="text-fg-muted">(9)</span>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
