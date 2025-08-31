import Button from "@components/Button";
import ServerError from "@components/ServerError";
import { invalidateLocalStorageAuth } from "@lib/lib";
import { createFileRoute } from "@tanstack/react-router";
import { useCallback, useState } from "react";
import SectionHeader from "../-components/SectionHeader";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_pages/profile"
)({
  component: ProfilePage,
});

function ProfilePage() {
  const navigate = Route.useNavigate();
  const [serverError, setServerError] = useState();

  const logOutOnAllDevices = useCallback(async () => {
    const response = await fetch("api/v1/auth/user/logout/all", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    });

    if (!response.ok) {
      const result = await response.json();
      setServerError(result.error.message);
    }

    invalidateLocalStorageAuth(401);
    navigate({
      from: Route.fullPath,
      to: "/",
    });
  }, [navigate]);

  return (
    <div className="flex flex-col gap-6">
      <ServerError error={serverError} />
      <div className="flex flex-col">
        <SectionHeader title="Change Password" styles="" />
      </div>

      <div className="flex flex-col gap-4">
        <SectionHeader styles="text-red-600 font-semibold" title="Log out" />
        <p className="text-text_color/70">
          Once you log out there is no goiong back! Please be certain.
        </p>
        <Button
          variant="danger"
          styles="py-2 px-2 w-min text-nowrap"
          buttonText="Log out on all devices"
          onClick={logOutOnAllDevices}
        />
      </div>
    </div>
  );
}
