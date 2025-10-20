import { BackArrowIcon } from "@reservations/assets";
import { formatDuration } from "@reservations/lib";
import { Link } from "@tanstack/react-router";

export default function ServiceItem({ children, service, router, locationId }) {
  return (
    <div className="flex flex-row gap-3">
      <div
        className="mt-1 size-12 shrink-0 overflow-hidden rounded-lg md:size-20"
      >
        <img
          className="size-full object-cover"
          src="https://dummyimage.com/80x80/d156c3/000000.jpg"
          alt="service photo"
        />
      </div>
      <div className="flex w-full flex-col gap-3 md:flex-row md:justify-between">
        <div
          className="flex flex-row items-center md:flex-col md:justify-center"
        >
          <div className="flex flex-col">
            <p className="text-lg font-semibold">{service.name}</p>
            <p className="hidden text-sm md:block">
              {formatDuration(service.total_duration)}
            </p>

            <Link
              from={router.fullPath}
              to={`services/${service.id}`}
              search={{
                locationId: locationId,
              }}
              className="hover:text-text_color hover:stroke-text_color flex
                items-center gap-1 pt-2 text-sm text-gray-700
                dark:text-gray-300"
            >
              See service details
              <BackArrowIcon
                styles="size-4 rotate-180 stroke-gray-700 dark:stroke-gray-300"
              />
            </Link>
          </div>
        </div>
        <div
          className="flex flex-row items-center justify-between gap-2 md:gap-4"
        >
          <p className="w-full font-semibold">
            {service.price && `${service.price} ${service.price_note}`}
          </p>
          {children}
        </div>
      </div>
    </div>
  );
}
