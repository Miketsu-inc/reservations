import { PlusSignIcon, WalkingIcon } from "@hugeicons/core-free-icons";
import {
  Avatar,
  Button,
  CheckBox,
  Icon,
  SearchInput,
} from "@reservations/components";
import { useState } from "react";
import NewCustomerOverlay from "./NewCutomerForm";

export default function CustomerSelector({
  customers,
  onSave,
  isGroupMode = true,
  styles,
  selected = [],
}) {
  const [searchText, setSearchText] = useState("");
  const [selectedCustomers, setSelectedCustomers] = useState(selected);

  const filteredCustomers = customers.filter(
    (customer) =>
      customer.first_name.toLowerCase().includes(searchText.toLowerCase()) ||
      customer.last_name.toLowerCase().includes(searchText.toLowerCase())
  );

  const handleCustomerClick = (customer) => {
    if (!isGroupMode) {
      onSave([customer]);
      return;
    }

    const isSelected = selectedCustomers.some(
      (c) => c.customer_id === customer.customer_id
    );

    if (isSelected) {
      setSelectedCustomers((prev) =>
        prev.filter((c) => c.customer_id !== customer.customer_id)
      );
    } else {
      setSelectedCustomers((prev) => [...prev, customer]);
    }
  };

  function handleNewCustomer(customer) {
    onSave([customer]);
  }

  return (
    <div className={"relative flex h-full w-full flex-col px-4"}>
      <div className={`flex flex-col gap-5 pt-6 pb-3 ${styles}`}>
        <div className="flex items-center justify-between">
          <h2 className="text-fg text-lg font-semibold">
            {isGroupMode ? "Select clients" : "Select a client"}
          </h2>
        </div>

        <SearchInput
          searchText={searchText}
          onChange={(text) => setSearchText(text)}
          placeholder="Search client..."
          styles="w-full"
        />
      </div>

      {!isGroupMode && (
        <ul className="border-border flex flex-col gap-1 border-b pb-2">
          <NewCustomerOverlay
            onSave={handleNewCustomer}
            trigger={
              <CustomerRow
                variant="action"
                label="Add new client"
                icon={<Icon icon={PlusSignIcon} styles="size-6" />}
              />
            }
          />
          <CustomerRow
            variant="action"
            label="Walk-In"
            icon={<Icon icon={WalkingIcon} styles="size-6" />}
            onSelect={() => onSave([])}
          />
        </ul>
      )}

      <div
        className="min-h-0 flex-1 overflow-y-auto pt-4 pb-24 dark:scheme-dark"
      >
        <ul className="flex flex-col gap-1">
          {filteredCustomers?.map((customer) => {
            const isSelected = selectedCustomers.some(
              (c) => c.customer_id === customer.customer_id
            );
            return (
              <li key={customer.customer_id}>
                <CustomerRow
                  customer={customer}
                  variant="customer"
                  isSelected={isSelected}
                  onSelect={handleCustomerClick}
                  isGroupMode={isGroupMode}
                />
              </li>
            );
          })}
        </ul>
      </div>

      {isGroupMode && (
        <div
          className="bg-bg border-border absolute right-0 bottom-0 left-0
            border-t px-4 py-2"
        >
          <Button
            buttonText="Save"
            onClick={() => onSave(selectedCustomers)}
            styles="w-full py-2 text-sm"
          />
        </div>
      )}
    </div>
  );
}

function CustomerRow({
  onSelect,
  customer,
  icon,
  label,
  variant = "customer",
  isSelected = false,
  isGroupMode,
  ...props
}) {
  const isAction = variant === "action";

  return (
    <button
      onClick={onSelect ? () => onSelect(customer) : undefined}
      className="hover:bg-hover flex w-full cursor-pointer items-center gap-3
        rounded-lg px-2 py-2 text-left"
      {...props}
    >
      {isAction ? (
        <div
          className="text-fg-primary bg-primary-subtle flex size-12 shrink-0
            items-center justify-center rounded-full"
        >
          {icon}
        </div>
      ) : (
        <Avatar
          styles="size-12! text-sm! shrink-0"
          img={customer?.avatar_url}
          initials={
            customer?.first_name && customer?.last_name
              ? `${customer.first_name[0]}${customer.last_name[0]}`
              : "?"
          }
        />
      )}

      <div className="flex min-w-0 flex-1 flex-col">
        <span className="text-fg truncate text-sm">
          {isAction ? label : `${customer?.first_name} ${customer?.last_name}`}
        </span>

        {!isAction && customer?.phone_number && (
          <span className="text-fg-muted text-xs">{customer.phone_number}</span>
        )}
      </div>

      {!isAction && isGroupMode && (
        <CheckBox checked={isSelected} readOnly styles="outline-none" />
      )}
    </button>
  );
}
