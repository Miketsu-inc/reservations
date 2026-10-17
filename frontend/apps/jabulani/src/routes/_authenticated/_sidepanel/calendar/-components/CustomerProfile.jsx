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
        <div className="flex flex-col items-center gap-6">
          <Avatar
            styles="size-28! text-[24px]! shrink-0"
            img={customer?.avatar_url}
            initials={`${customer?.first_name?.[0] || "?"}${customer?.last_name?.[0] || ""}`}
          />
          <div className="text-fg text-center text-xl font-bold">
            {customer?.first_name} {customer?.last_name}
          </div>
        </div>

        <div className="flex w-full flex-col items-center gap-8">
          <div className="flex w-full flex-col items-center gap-2">
            {customer.email && (
              <div className="flex items-center gap-3 text-sm">
                <Icon icon={Mail01Icon} styles="size-5 text-fg-muted" />
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

          <div className="flex w-full items-center justify-center gap-5 pb-5">
            {!disabled && (
              <Button
                variant="tertiary"
                buttonText="Remove"
                onClick={onRemove}
                styles="px-2 py-1.5 w-28"
              />
            )}
            <Link
              className="text-fg border-border hover:bg-hover w-28 rounded-lg
                border-2 bg-transparent px-2 py-1.5 text-center shadow-none"
              to={`/customers/${customer?.customer_id}`}
              from=""
            >
              View Profile
            </Link>
          </div>
        </div>
      </div>
      <div className="flex w-full flex-col justify-start gap-3 px-8">
        {customer.birthday && (
          <div className="flex items-center gap-3 text-sm">
            <Icon icon={BirthdayCakeIcon} styles="size-5 text-fg-muted" />
            <span>{formatBirthday(customer.birthday)}</span>
          </div>
        )}
        {customer.last_visited && (
          <div className="flex items-center gap-3 text-sm">
            <Icon icon={User03Icon} styles="size-5 text-fg-muted" />

            <span className="">
              Last visited: {formatBirthday(customer.last_visited)}
            </span>
          </div>
        )}
      </div>
      {customer.customer_note && (
        <div className="w-full px-4 pb-4">
          <div className="bg-hover rounded-md px-4 py-4">
            <div className="mb-2 flex items-center gap-3">
              <Icon icon={Message01Icon} styles="size-4 mt-1 fill-fg" />
              <span className="text-sm font-medium">
                Note from {customer.first_name}
              </span>
            </div>
            <p className="text-fg-muted text-sm leading-relaxed">
              {customer.customer_note}
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
