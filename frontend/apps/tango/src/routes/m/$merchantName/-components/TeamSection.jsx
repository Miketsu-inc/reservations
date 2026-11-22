import { Avatar, Loading, ServerError } from "@reservations/components";
import {
  activeTeamQueryOptions,
  formatShortName,
  toAvatar,
  useWindowSize,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import RatingSummary from "./RatingSummary";

export default function TeamSection({ merchantName, locationId }) {
  const { isWindowSmall } = useWindowSize();

  const {
    data: employees,
    isLoading,
    isError,
    error,
  } = useQuery(activeTeamQueryOptions(merchantName, locationId));

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
        {employees.map((employee) => {
          const { img, initials } = toAvatar(employee);

          return (
            <div
              key={employee.id}
              className="flex w-32 shrink-0 flex-col items-center gap-4
                text-center"
            >
              <Avatar
                styles={`shrink-0 font-semibold
                ${isWindowSmall ? "size-20 text-lg!" : "size-22 text-xl!"}`}
                img={img}
                initials={initials}
              />

              <div className="flex flex-col items-center justify-center gap-1.5">
                <span className="leading-tight font-semibold">
                  {formatShortName(employee)}
                </span>

                {employee.review_count > 0 && (
                  <RatingSummary
                    average={employee.average_rating}
                    styles="text-sm"
                  >
                    <span className="text-fg-muted">
                      ({employee.review_count})
                    </span>
                  </RatingSummary>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
