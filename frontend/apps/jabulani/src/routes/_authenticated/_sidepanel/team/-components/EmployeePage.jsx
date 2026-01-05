import { Button, Card, Input, Select, Switch } from "@reservations/components";
import { Link, useRouter } from "@tanstack/react-router";
import { useMemo, useState } from "react";

const employeeFields = [
  "first_name",
  "last_name",
  "email",
  "phone_number",
  "role",
  "is_active",
];

export default function EmployeePage({ employee, onSave }) {
  const originalData = useMemo(() => {
    return {
      id: employee?.id,
      first_name: employee?.first_name || "",
      last_name: employee?.last_name || "",
      email: employee?.email || "",
      phone_number: employee?.phone_number || "",
      role: employee?.role || "",
      is_active: employee?.is_active ?? true,
    };
  }, [employee]);
  const [employeeData, setEmployeeData] = useState(originalData);

  console.log(employeeData);

  const router = useRouter();

  function updateEmployeeData(data) {
    setEmployeeData((prev) => ({ ...prev, ...data }));
  }

  function submitHandler(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    const changes = {};

    if (employeeData.id) {
      changes.id = employeeData.id;
    }

    console.log("employeeData: ", employeeData);

    employeeFields.forEach((field) => {
      if (employeeData[field] !== originalData[field]) {
        changes[field] = employeeData[field];
      }
    });

    console.log("changes before: ", changes);

    if (!changes["is_active"]) {
      changes["is_active"] = employeeData["is_active"];
    }

    console.log("changes after: ", changes);

    onSave(changes);
  }

  return (
    <form onSubmit={submitHandler} className="flex justify-center pt-4 md:p-0">
      <div
        className="flex w-full flex-col gap-4 px-3 sm:px-0 lg:w-2/3 2xl:w-1/2"
      >
        <Card>
          <p className="text-xl">
            {employee ? "Edit team member" : "Add new team member"}
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
              value={employeeData.first_name}
              inputData={(data) =>
                updateEmployeeData({ first_name: data.value })
              }
            />
            <Input
              styles="p-2"
              id="LastName"
              name="LastName"
              type="text"
              labelText="Last Name"
              placeholder="Scott"
              value={employeeData.last_name}
              inputData={(data) =>
                updateEmployeeData({ last_name: data.value })
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
            value={employeeData.email}
            inputData={(data) => updateEmployeeData({ email: data.value })}
          />
          <Input
            styles="p-2"
            id="PhoneNumber"
            name="PhoneNumber"
            type="text"
            labelText="Phone Number"
            placeholder="+36 20 678 2012"
            required={false}
            value={employeeData.phone_number}
            inputData={(data) =>
              updateEmployeeData({ phone_number: data.value })
            }
          />
          <Select
            required={true}
            labelText="Role"
            styles="w-full"
            value={employeeData.role}
            options={[
              { value: "staff", label: "Staff" },
              { value: "admin", label: "Admin" },
            ]}
            allOptions={[
              { value: "staff", label: "Staff" },
              { value: "admin", label: "Admin" },
              { value: "owner", label: "Owner" },
            ]}
            onSelect={(option) => updateEmployeeData({ role: option.value })}
            placeholder="Pick a role"
          />
          <div className="flex flex-row items-center gap-4">
            <p>Is active</p>
            <Switch
              size="large"
              defaultValue={employeeData.is_active}
              onSwitch={() =>
                updateEmployeeData({ is_active: !employeeData.is_active })
              }
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
              buttonText={employee ? "Save" : "Create team member"}
            />
          </div>
        </Card>
      </div>
    </form>
  );
}
