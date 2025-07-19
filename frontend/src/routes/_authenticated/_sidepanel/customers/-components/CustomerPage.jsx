import Button from "@components/Button";
import Card from "@components/Card";
import Input from "@components/Input";
import Select from "@components/Select";
import { Block, Link, useRouter } from "@tanstack/react-router";
import { useEffect, useMemo, useState } from "react";

const Customerfields = [
  "first_name",
  "last_name",
  "email",
  "phone_number",
  "birthday_day",
  "birthday_month",
  "note",
];

export default function CustomerPage({ customer, onSave }) {
  const originalData = useMemo(() => {
    const date = customer?.birthday ? new Date(customer.birthday) : null;

    return {
      id: customer?.id,
      first_name: customer?.first_name || "",
      last_name: customer?.last_name || "",
      email: customer?.email || "",
      phone_number: customer?.phone_number || "",
      birthday_day: date ? date.getUTCDate() : "",
      birthday_month: date ? date.getUTCMonth() + 1 : "",
      note: customer?.note || "",
    };
  }, [customer]);
  const router = useRouter();
  const [customerData, setCustomerData] = useState(originalData);
  const [lastSavedData, setLastSavedData] = useState();

  useEffect(() => {
    setCustomerData(originalData);
    setLastSavedData(originalData);
  }, [originalData]);

  function updateCustomerData(data) {
    setCustomerData((prev) => ({ ...prev, ...data }));
  }

  function submitHandler(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    const changes = {};

    Customerfields.forEach((field) => {
      if (customerData[field] !== lastSavedData[field]) {
        changes[field] = customerData[field];
      }
    });

    if (changes?.birthday_day || changes?.birthday_month) {
      const day = String(
        changes.birthday_day || customerData.birthday_day
      ).padStart(2, "0");
      const month = String(
        changes.birthday_month || customerData.birthday_month
      ).padStart(2, "0");

      changes.birthday =
        (changes.birthday_day || customerData.birthday_day) &&
        (changes.birthday_month || customerData.birthday_month)
          ? new Date(`2000-${month}-${day}T00:00:00Z`).toISOString()
          : null;
    }
    if (customerData.id) {
      changes.id = customerData.id;
    }

    setLastSavedData(customerData);
    onSave(changes);
  }

  return (
    <Block
      shouldBlockFn={() => {
        if (JSON.stringify(customerData) === JSON.stringify(lastSavedData)) {
          return false;
        }
        const canLeave = confirm(
          "You have unsaved changes, are you sure you want to leave?"
        );
        return !canLeave;
      }}
    >
      <form
        onSubmit={submitHandler}
        className="flex justify-center pt-4 md:p-0"
      >
        <div className="flex w-full flex-col gap-4 px-3 sm:px-0 lg:w-2/3 2xl:w-1/2">
          <Card>
            <p className="text-xl">
              {customer ? "Edit Your Customer" : "Add New Customer"}
            </p>
          </Card>
          <Card styles="flex flex-col gap-4">
            <div className="flex flex-col gap-4 md:flex-row">
              <Input
                styles="p-2"
                id="FirstName"
                name="FirstName"
                type="text"
                labelText="First Name"
                placeholder="Travis"
                value={customerData.first_name}
                inputData={(data) =>
                  updateCustomerData({ first_name: data.value })
                }
              />
              <Input
                styles="p-2"
                id="LastName"
                name="LastName"
                type="text"
                labelText="Last Name"
                placeholder="Scott"
                value={customerData.last_name}
                inputData={(data) =>
                  updateCustomerData({ last_name: data.value })
                }
              />
            </div>
            <Input
              styles="p-2"
              id="Email"
              name="Email"
              type="email"
              labelText="Email"
              placeholder="example@gmail.com"
              required={false}
              value={customerData.email}
              inputData={(data) => updateCustomerData({ email: data.value })}
              disabled={customer ? !customer.is_dummy : false}
            />
            <Input
              styles="p-2"
              id="PhoneNumber"
              name="PhoneNumber"
              type="text"
              labelText="Phone Number"
              placeholder="+36 20 678 2012"
              required={false}
              value={customerData.phone_number}
              inputData={(data) =>
                updateCustomerData({ phone_number: data.value })
              }
            />
            <Input
              styles="p-2"
              id="Birthday"
              name="Birthday"
              type="number"
              min={1}
              max={31}
              labelText="Birthday"
              required={false}
              placeholder="e.g. 4"
              value={customerData.birthday_day}
              inputData={(data) =>
                updateCustomerData({ birthday_day: data.value })
              }
              childrenSide="left"
            >
              <Select
                styles="w-full rounded-r-none"
                value={customerData.birthday_month}
                options={[
                  { value: 1, label: "January" },
                  { value: 2, label: "February" },
                  { value: 3, label: "March" },
                  { value: 4, label: "April" },
                  { value: 5, label: "May" },
                  { value: 6, label: "June" },
                  { value: 7, label: "July" },
                  { value: 8, label: "August" },
                  { value: 9, label: "September" },
                  { value: 10, label: "October" },
                  { value: 11, label: "November" },
                  { value: 12, label: "December" },
                ]}
                onSelect={(option) =>
                  updateCustomerData({ birthday_month: option.value })
                }
                placeholder="Choose a month"
              />
            </Input>

            <div className="flex flex-col gap-1">
              <label htmlFor="note">Note</label>
              <textarea
                id="note"
                name="note"
                placeholder="About this cutomer..."
                className="bg-layer_bg focus:ring-primary/30 focus:border-primary border-input_border_color max-h-20 min-h-20 w-full rounded-lg border p-2 text-sm outline-hidden transition-[border-color,box-shadow] focus:ring-4 md:max-h-32 md:min-h-32"
                value={customerData.note}
                onChange={(e) => updateCustomerData({ note: e.target.value })}
              />
            </div>
            <div className="flex w-full justify-end gap-2">
              <Link from={router.fullPath}>
                <Button
                  variant="tertiary"
                  styles="sm:py-2 sm:px-4 w-fit p-2"
                  buttonText="Cancel"
                  type="button"
                  onClick={() => router.history.back()}
                />
              </Link>
              <Button
                styles="py-2 px-4"
                variant="primary"
                type="submit"
                buttonText={customer ? "Save" : "Create Customer"}
              />
            </div>
          </Card>
        </div>
      </form>
    </Block>
  );
}
