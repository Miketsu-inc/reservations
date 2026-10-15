import { Alert02Icon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";

export default function TimezoneWarning({ merchantTimeZone }) {
  const browserTimeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;

  if (!merchantTimeZone || browserTimeZone === merchantTimeZone) return null;

  return (
    <div
      className="bord border-border-strong bg-warning mt-4 flex items-start
        gap-2 rounded-md border p-3 text-sm text-black"
      role="status"
    >
      <Icon icon={Alert02Icon} styles="mt-0.5 size-5 shrink-0" />
      <p>
        Times are shown in the merchant's timezone: {merchantTimeZone}. Your
        device is set to {browserTimeZone}.
      </p>
    </div>
  );
}
