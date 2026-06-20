import { ArrowLeft01Icon } from "@hugeicons/core-free-icons";
import {
  Button,
  Card,
  Icon,
  Input,
  Loading,
  ServerError,
} from "@reservations/components";
import { meQueryOptions, useToast, useWindowSize } from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link, useRouteContext } from "@tanstack/react-router";
import { useState } from "react";

export const Route = createFileRoute(
  "/_authenticated/_navigation/profile/edit"
)({
  loader: async ({ context: { queryClient } }) => {
    await queryClient.ensureQueryData(meQueryOptions());
  },
  component: RouteComponent,
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function RouteComponent() {
  const { data: user, isLoading } = useQuery(meQueryOptions());
  const { showToast } = useToast();
  const navigate = Route.useNavigate();
  const { queryClient } = useRouteContext({ from: Route.id });
  const windowSize = useWindowSize();
  const isWindowSmall = windowSize === "sm" || windowSize === "md";

  async function invalidateMeQuery() {
    await queryClient.invalidateQueries({
      queryKey: ["me"],
    });
  }

  async function submitHandler(e) {
    e.preventDefault();

    const response = await fetch("/api/v1/users", {
      method: "PUT",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify(userData),
    });

    if (!response.ok) {
      const result = await response.json();
      showToast({ message: result.error.message, variant: "error" });
    } else {
      invalidateMeQuery();
      navigate({
        from: Route.fullPath,
        to: "/profile",
      });
    }
  }

  const [userData, setUserData] = useState(user);

  function updateUserData(data) {
    setUserData((prev) => ({ ...prev, ...data }));
  }

  if (isLoading) {
    return <Loading />;
  }

  return (
    <div className="flex justify-center">
      <div className="flex w-full flex-col justify-center md:max-w-lg">
        <div className="flex flex-row items-center gap-4 pt-4 pb-8">
          {!isWindowSmall && (
            <Link from={Route.fullPath} to="..">
              <Icon icon={ArrowLeft01Icon} styles="size-7 text-text_color" />
            </Link>
          )}
          <p className="text-2xl">Edit profile</p>
        </div>
        <Card styles="p-8">
          <form className="flex flex-col gap-4" onSubmit={submitHandler}>
            <div
              className="flex flex-col gap-6 md:flex-row md:items-center
                md:gap-4"
            >
              <Input
                name="Firstname"
                styles="p-2"
                labelText="First name"
                value={userData.first_name}
                inputData={(data) => updateUserData({ first_name: data.value })}
              />
              <Input
                name="Lastname"
                styles="p-2"
                labelText="Last name"
                value={userData.last_name}
                inputData={(data) => updateUserData({ last_name: data.value })}
              />
            </div>
            <Input
              id="PhoneNumber"
              name="Phonenumber"
              type="tel"
              labelText="Phone number"
              placeholder="+36 20 678 2012"
              value={userData.phone_number}
              inputData={(data) => updateUserData({ phone_number: data.value })}
            />
            <Input
              name="Email"
              styles="p-2"
              labelText="Email"
              value={userData.email}
              inputData={(data) => updateUserData({ email: data.value })}
            />
            <Button styles="w-full py-2 mt-2" buttonText="Save" />
          </form>
        </Card>
      </div>
    </div>
  );
}
