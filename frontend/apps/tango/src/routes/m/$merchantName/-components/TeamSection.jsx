import { StarIcon } from "@hugeicons/core-free-icons";
import { Avatar, Icon, Loading, ServerError } from "@reservations/components";
import { activeTeamQueryOptions } from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";

export default function TeamSection({ isWindowSmall, merchantName }) {
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
        className={`flex w-full ${
          isWindowSmall
            ? "scrollbar-thin gap-8 overflow-x-auto"
            : "flex-wrap gap-x-16 gap-y-12"
          } dark:scheme-dark`}
      >
        {employees.map((employee) => (
          <div
            key={employee.id}
            className="flex w-32 shrink-0 flex-col items-center gap-4
              text-center"
          >
            <Avatar
              styles={`rounded-full! shrink-0 font-semibold
              ${isWindowSmall ? "size-23 text-xl!" : "size-26 text-2xl!"}`}
              initials={`${employee.first_name.charAt(0)}${employee.last_name.charAt(0)}`}
            />

            <div className="flex flex-col items-center justify-center gap-1.5">
              <span className="text-lg leading-tight font-medium">
                {employee.first_name} {employee.last_name.charAt(0)}.
              </span>

              <div className="flex items-center gap-1.5">
                <Icon
                  icon={StarIcon}
                  styles="fill-yellow-500 size-4.5 text-yellow-500"
                />
                <div className="flex items-center gap-1 text-sm">
                  <span className="font-medium">4.6</span>
                  <span className="text-gray-500 dark:text-gray-400">(9)</span>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
