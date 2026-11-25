import { Clock01Icon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";
import { WEEKDAYS } from "@reservations/lib";

export function Section({ children, name, show = true }) {
  if (!show) return null;

  return (
    <section>
      <h2 className="pb-6 text-xl font-semibold lg:text-2xl">{name}</h2>
      {children}
    </section>
  );
}

export function DirectionsLink({ geoPoint }) {
  return (
    <a
      href={`https://www.google.com/maps/dir/?api=1&destination=${geoPoint.latitude},${geoPoint.longitude}`}
      target="_blank"
      rel="noopener noreferrer"
      className="text-fg-primary text-nowrap hover:underline"
    >
      (Plan a ride)
    </a>
  );
}

function nextOpenText(status) {
  switch (status.days_until_open) {
    case null:
    case undefined:
      return "";
    case 0:
      return "opens later today";
    case 1:
      return "opens tomorrow";
    default:
      return `opens on ${WEEKDAYS[status.next_open_day]}`;
  }
}

export function OpenStatus({ status, styles, iconStyles = "size-5" }) {
  const detail = status.is_business_open
    ? `closes at ${status.close_time}`
    : nextOpenText(status);

  return (
    <div className={`flex items-center gap-2 ${styles ?? ""}`}>
      <Icon icon={Clock01Icon} styles={`shrink-0 ${iconStyles}`} />
      <span className="text-nowrap">
        {status.is_business_open ? (
          <span className="text-fg-success font-medium">Open</span>
        ) : (
          <span className="text-fg-warning font-medium">Closed</span>
        )}
        {detail && ` - ${detail}`}
      </span>
    </div>
  );
}
