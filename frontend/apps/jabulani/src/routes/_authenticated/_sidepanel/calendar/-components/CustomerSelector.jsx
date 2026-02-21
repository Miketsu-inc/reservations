import { PlusIcon } from "@reservations/assets"; // Make sure you have a CheckIcon
import {
  Avatar,
  Button,
  CheckBox,
  SearchInput,
} from "@reservations/components";
import { useState } from "react";

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

    const isSelected = selectedCustomers.some((c) => c.id === customer.id);

    if (isSelected) {
      setSelectedCustomers((prev) => prev.filter((c) => c.id !== customer.id));
    } else {
      setSelectedCustomers((prev) => [...prev, customer]);
    }
  };

  return (
    <div className="relative flex h-full flex-col">
      <div className={`flex flex-col gap-5 px-4 pt-6 pb-2 ${styles}`}>
        <div className="flex items-center justify-between">
          <h2 className="text-text_color text-xl font-semibold">
            {isGroupMode ? "Select clients" : "Select a client"}
          </h2>
        </div>

        <SearchInput
          searchText={searchText}
          onChange={(text) => setSearchText(text)}
          placeholder="Search client..."
          styles="w-full p-2"
        />
      </div>

      <ul className="border-border_color flex flex-col gap-3 border-b px-3 pb-2">
        <CustomerRow
          variant="action"
          label="Add new client"
          icon={<PlusIcon styles="size-6" />}
          onClick={() => console.log("Create new client")}
        />
        {!isGroupMode && (
          <CustomerRow
            variant="action"
            label="Walk-In"
            icon={
              <svg
                className="size-6"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M13.5 4.5L11 13l-4-2-2 5m8-3l4 4m0-11a2 2 0 11-4 0 2 2 0 014 0z"
                />
              </svg>
            }
            onClick={() => onSave([])}
          />
        )}
      </ul>

      <div
        className="min-h-0 flex-1 overflow-y-auto px-3 pt-4 pb-24
          dark:scheme-dark"
      >
        <ul className="flex flex-col gap-3">
          {filteredCustomers?.map((customer) => {
            const isSelected = selectedCustomers.some(
              (c) => c.id === customer.id
            );
            return (
              <li key={customer.id}>
                <CustomerRow
                  customer={customer}
                  variant="customer"
                  isSelected={isSelected}
                  onClick={handleCustomerClick}
                  isGroupMode={isGroupMode}
                />
              </li>
            );
          })}
        </ul>
      </div>

      {isGroupMode && (
        <div
          className="bg-layer_bg border-border_color absolute right-0 bottom-0
            left-0 border-t p-4"
        >
          <Button
            buttonText="Save"
            onClick={() => onSave(selectedCustomers)}
            styles="w-full p-2"
          />
        </div>
      )}
    </div>
  );
}

function CustomerRow({
  onClick,
  customer,
  icon,
  label,
  variant = "customer",
  isSelected = false,
  isGroupMode,
}) {
  const isAction = variant === "action";

  return (
    <button
      onClick={() => onClick(customer)}
      className="flex w-full items-center gap-4 rounded-xl px-3 py-2 text-left
        transition-all hover:bg-gray-200/40 dark:hover:bg-gray-700/10"
    >
      {isAction ? (
        <div
          className="text-primary bg-primary/20 flex size-14 shrink-0
            items-center justify-center rounded-full"
        >
          {icon}
        </div>
      ) : (
        <Avatar
          styles="size-14! text-[16px]! shrink-0 rounded-full!"
          img={customer?.avatar_url}
          initials={
            customer?.first_name && customer?.last_name
              ? `${customer.first_name[0]}${customer.last_name[0]}`
              : "?"
          }
        />
      )}

      <div className="flex flex-1 flex-col">
        <span className="text-text_color font-semibold">
          {isAction ? label : `${customer?.first_name} ${customer?.last_name}`}
        </span>

        {!isAction && customer?.phone_number && (
          <span className="text-sm text-gray-500">{customer.phone_number}</span>
        )}
      </div>

      {!isAction && isGroupMode && (
        <CheckBox checked={isSelected} readOnly styles="outline-none" />
      )}
    </button>
  );
}
