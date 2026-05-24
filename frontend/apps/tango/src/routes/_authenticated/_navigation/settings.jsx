import {
  Logout05Icon,
  Moon02Icon,
  Sun03Icon,
} from "@hugeicons/core-free-icons";
import {
  Button,
  Card,
  DeleteModal,
  Icon,
  Switch,
} from "@reservations/components";
import { useTheme, useToast, useWindowSize } from "@reservations/lib";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import ChangePasswordModal from "./-components/ChangePasswordModal";

export const Route = createFileRoute("/_authenticated/_navigation/settings")({
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = Route.useNavigate();
  const windowSize = useWindowSize();
  const { showToast } = useToast();
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState();
  const [isChangePasswordModalOpen, setIsChangePasswordModalOpen] = useState();

  const { isDarkTheme, switchTheme } = useTheme();

  const isWindowSmall = windowSize === "sm" || windowSize === "md";

  async function logoutHandler() {
    const response = await fetch("/api/v1/auth/logout", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    });

    if (response.ok) {
      navigate({
        from: Route.fullPath,
        to: "/",
      });
    }
  }

  async function deleteHandler() {
    const response = await fetch("/api/v1/users", {
      method: "DELETE",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    });

    if (!response.ok) {
      const result = await response.json();
      showToast({ message: result.error.message, variant: "error" });
    } else {
      navigate({
        from: Route.fullPath,
        to: "/",
      });
    }
  }

  return (
    <div className="flex justify-center">
      <div className="flex w-full max-w-sm flex-col justify-center">
        <p className="pt-4 pb-8 text-2xl">Settings</p>
        <div className="flex flex-col gap-4">
          <Card>
            <p className="mb-4 text-lg">Change password</p>
            <Button
              styles="px-4 py-2"
              variant="danger"
              buttonText="Update password"
              onClick={() => setIsChangePasswordModalOpen(true)}
            />
          </Card>
          <Card>
            <p className="mb-4 text-lg">Delete account</p>
            <Button
              styles="px-4 py-2"
              variant="danger"
              buttonText="Delete account"
              onClick={() => setIsDeleteModalOpen(true)}
            />
          </Card>
          {isWindowSmall && (
            <>
              <Card>
                <div
                  className="flex w-full flex-row items-center justify-between
                    p-2"
                >
                  <div className="flex flex-row items-center gap-4">
                    <Icon
                      icon={Moon02Icon}
                      altIcon={Sun03Icon}
                      showAlt={isDarkTheme}
                      styles="size-6"
                    />
                    <p>Use dark theme</p>
                  </div>
                  <Switch
                    size="large"
                    defaultValue={isDarkTheme}
                    onSwitch={switchTheme}
                  />
                </div>
              </Card>
              <Card>
                <button
                  type="button"
                  className="flex w-full flex-row items-center gap-4 p-2"
                  onClick={logoutHandler}
                >
                  <Icon icon={Logout05Icon} styles="size-6" />
                  <p>Sign out</p>
                </button>
              </Card>
            </>
          )}
        </div>
        <DeleteModal
          isOpen={isDeleteModalOpen}
          onClose={() => setIsDeleteModalOpen(false)}
          onDelete={deleteHandler}
          itemName="your user account"
        />
        <ChangePasswordModal
          isOpen={isChangePasswordModalOpen}
          onClose={() => setIsChangePasswordModalOpen(false)}
        />
      </div>
    </div>
  );
}
