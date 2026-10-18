import { Alert02Icon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";

export default function TimezoneWarning({ merchantTimeZone }) {
  const browserTimeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;

  if (!merchantTimeZone || browserTimeZone === merchantTimeZone) return null;

  return (
    <div
      className="border-border-warning bg-warning-subtle text-fg mt-4 flex
        items-start gap-2 rounded-lg border p-3 text-sm"
      role="status"
    >
      <Icon
        icon={Alert02Icon}
        styles="text-fg-warning mt-0.5 size-5 shrink-0"
      />
      <p>
        Times are shown in the merchant's timezone: {merchantTimeZone}. Your
        device is set to {browserTimeZone}.
      </p>
    </div>
  );
}
