import { formatDuration } from "@lib/datetime";

export default function ServiceItem({ children, service }) {
  return (
    <div className="flex flex-row gap-3">
      <div className="mt-1 size-12 shrink-0 overflow-hidden rounded-lg md:size-20">
        <img
          className="size-full object-cover"
          src="https://dummyimage.com/80x80/d156c3/000000.jpg"
          alt="service photo"
        />
      </div>
      <div className="flex w-full flex-col gap-3 md:flex-row md:justify-between">
        <div className="flex flex-row items-center md:flex-col md:justify-center">
          <div className="flex flex-col">
            <p className="text-lg font-semibold">{service.name}</p>
            <p
              className="line-clamp-2 text-sm text-gray-600 md:line-clamp-none md:max-w-70 md:truncate
                dark:text-gray-400"
            >
              {service.description}
            </p>
            <p className="hidden pt-2 text-sm md:block">
              {formatDuration(service.total_duration)}
            </p>
          </div>
        </div>
        <div className="flex flex-row items-center justify-between gap-2 md:gap-4">
          <p className="w-full font-semibold">
            {`${parseFloat(service.price).toLocaleString()} HUF ${service.price_note}`}
          </p>
          {children}
        </div>
      </div>
    </div>
  );
}
