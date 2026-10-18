import {
  BirthdayCakeIcon,
  Call02Icon,
  Mail01Icon,
  Message01Icon,
  User03Icon,
} from "@hugeicons/core-free-icons";
import { Avatar, Button, Icon } from "@reservations/components";
import { Link } from "@tanstack/react-router";

function formatBirthday(datestr) {
  const date = new Date(datestr);
  return date.toLocaleDateString("en-US", {
    month: "long",
    day: "numeric",
  });
}

export default function CustomerProfile({
  customer,
  onRemove,
  styles,
  disabled,
}) {
  return (
    <div
      className={`flex h-full w-full flex-col items-center gap-5 sm:py-10
        ${styles}`}
    >
      <div
        className="border-border flex w-full flex-col items-center gap-3
          border-b"
      >
        <div className="flex flex-col items-center gap-4">
          <Avatar
            styles="size-24! text-xl! shrink-0"
            img={customer?.avatar_url}
            initials={`${customer?.first_name?.[0] || "?"}${customer?.last_name?.[0] || ""}`}
          />
          <div className="text-fg text-center text-lg font-bold">
            {customer?.first_name} {customer?.last_name}
          </div>
        </div>

        <div className="flex w-full flex-col items-center gap-6">
          <div className="flex w-full flex-col items-center gap-2">
            {customer.email && (
              <div className="flex items-center gap-3 text-sm">
                <Icon icon={Mail01Icon} styles="size-4 text-fg-muted" />
                <span className="truncate">{customer.email}</span>
              </div>
            )}
            {customer.phone_number && (
              <div className="flex items-center gap-3 text-sm">
                <Icon icon={Call02Icon} styles="size-4 text-fg-muted" />

                <span>{customer.phone_number}</span>
              </div>
            )}
          </div>

          <div className="flex w-full items-center justify-center gap-2 pb-5">
            {!disabled && (
              <Button
                variant="secondary"
                buttonText="Remove"
                onClick={onRemove}
                styles="px-3 py-1.5 w-28 text-sm"
              />
            )}
            <Link
              className="text-fg border-border-strong bg-bg-input hover:bg-hover
                w-28 rounded-lg border px-3 py-1.5 text-center text-sm"
              to={`/customers/${customer?.customer_id}`}
              from=""
            >
              View Profile
            </Link>
          </div>
        </div>
      </div>
      {customer.birthday && (
        <div className="flex items-center gap-3 px-6 pb-3 text-sm">
          <Icon icon={BirthdayCakeIcon} styles="size-4 text-fg-muted" />
          <span>{formatBirthday(customer.birthday)}</span>
        </div>
      )}
      {customer.last_visited && (
        <div className="flex items-center gap-3 px-6 pb-3 text-sm">
          <Icon icon={User03Icon} styles="size-4 text-fg-muted" />

          <span>Last visited: {formatBirthday(customer.last_visited)}</span>
        </div>
      )}
      {customer.customer_note && (
        <div className="w-full px-4 pb-4">
          <div className="bg-bg-sunken rounded-lg px-4 py-3">
            <div className="mb-2 flex items-center gap-2">
              <Icon icon={Message01Icon} styles="size-4.5" />
              <span className="text-sm">Note from {customer.first_name}</span>
            </div>
            <p className="text-fg-muted text-sm">{customer.customer_note}</p>
          </div>
        </div>
      )}
    </div>
  );
}
