import { Alert02Icon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";

export default function TimezoneWarning({ merchantTimeZone }) {
  const browserTimeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;

  if (!merchantTimeZone || browserTimeZone === merchantTimeZone) return null;

  return (
    <div
      className="bord mt-4 flex items-start gap-2 rounded-md border
        border-stone-400 bg-amber-400 p-3 text-sm text-black
        dark:border-amber-800 dark:bg-amber-500"
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
