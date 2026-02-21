import {
  CakeIcon,
  EnvelopeIcon,
  PersonIcon,
  PhoneIcon,
} from "@reservations/assets";
import { Avatar, Button } from "@reservations/components";
import { Link } from "@tanstack/react-router";

function formatBirthday(datestr) {
  const date = new Date(datestr);
  return date.toLocaleDateString("en-US", {
    month: "long",
    day: "numeric",
  });
}

export default function CustomerProfile({ customer, onRemove, styles }) {
  return (
    <div
      className={`flex h-full w-full flex-col items-center gap-5 sm:py-10
        ${styles}`}
    >
      <div
        className="border-border_color flex w-full flex-col items-center gap-3
          border-b"
      >
        <div className="flex flex-col items-center gap-6">
          <Avatar
            styles="size-28! text-[24px]! shrink-0 rounded-full!"
            img={customer?.avatar_url}
            initials={`${customer?.first_name?.[0] || "?"}${customer?.last_name?.[0] || ""}`}
          />
          <div className="text-text_color text-center text-xl font-bold">
            {customer.first_name} {customer.last_name}
          </div>
        </div>

        {/* Details Section */}
        <div className="flex w-full flex-col items-center gap-10">
          <div className="flex w-full flex-col items-center gap-2">
            {customer.email && (
              <div className="flex items-center gap-3 text-sm">
                <EnvelopeIcon styles="size-5 text-text_color/70" />

                <span className="truncate">{customer.email}</span>
              </div>
            )}
            {customer.phone_number && (
              <div className="flex items-center gap-3 text-sm">
                <PhoneIcon
                  styles="size-4 fill-text_color/70 stroke-text_color/10"
                />

                <span>{customer.phone_number}</span>
              </div>
            )}
          </div>

          <div className="flex w-full items-center justify-center gap-5 pb-5">
            <Button
              variant="tertiary"
              buttonText="Remove"
              onClick={onRemove}
              styles="px-2 py-1.5 w-28"
            />
            <Link
              className="text-text_color w-28 rounded-lg border-2
                border-gray-300 bg-transparent px-2 py-1.5 text-center
                shadow-none hover:bg-gray-300 dark:border-gray-800
                dark:hover:bg-gray-800"
              to={`/customers/${customer?.id}`}
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
            <CakeIcon styles="size-5 text-text_color/70" />
            <span>{formatBirthday(customer.birthday)}</span>
          </div>
        )}
        {customer.last_visited && (
          <div className="flex items-center gap-3 text-sm">
            <PersonIcon styles="size-5 fill-text_color/70" />

            <span className="">
              Last visited: {formatBirthday(customer.last_visited)}
            </span>
          </div>
        )}
      </div>
    </div>
  );
}
