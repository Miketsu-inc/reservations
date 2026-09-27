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
import {
  invalidateLocalStorageAuth,
  useTheme,
  useToast,
  useWindowSize,
} from "@reservations/lib";
import { createFileRoute } from "@tanstack/react-router";
import ChangePasswordModal from "./-components/ChangePasswordModal";

export const Route = createFileRoute("/_authenticated/_navigation/settings")({
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = Route.useNavigate();
  const { isWindowSmall } = useWindowSize();
  const { showToast } = useToast();

  const { isDarkTheme, switchTheme } = useTheme();

  async function logoutHandler(allDevices = false) {
    try {
      const response = await fetch(
        allDevices ? "/api/v1/auth/logout/all" : "/api/v1/auth/logout",
        {
          method: "POST",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
        }
      );

      if (!response.ok) {
        const result = await response.json();
        showToast({ message: result.error.message, variant: "error" });
        return;
      }

      invalidateLocalStorageAuth(401);
      navigate({
        from: Route.fullPath,
        to: "/",
      });
    } catch (error) {
      showToast({ message: error.message, variant: "error" });
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
      <div className="flex w-full max-w-md flex-col justify-center pb-8">
        <div className="pt-4 pb-8">
          <h1 className="text-2xl font-semibold tracking-tight">Settings</h1>
          <p className="text-text_color/65 mt-1 text-sm">
            Manage your account, appearance, and active sessions.
          </p>
        </div>
        <div className="flex flex-col gap-4">
          <Card>
            <p className="text-lg font-medium">Password</p>
            <p className="text-text_color/65 mt-1 text-sm">
              Update the password used to sign in to your account.
            </p>
            <ChangePasswordModal
              trigger={
                <Button
                  styles="mt-4 px-4 py-2"
                  variant="tertiary"
                  buttonText="Update password"
                />
              }
            />
          </Card>
          {isWindowSmall && (
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
          )}
          <Card>
            <p className="text-lg font-medium">Sessions</p>
            <div className="mt-2 flex flex-col divide-y divide-gray-200 dark:divide-gray-700">
              <div
                className="flex flex-col items-start justify-between gap-3 py-4
                  sm:flex-row sm:items-center"
              >
                <div>
                  <p className="font-medium">Sign out</p>
                  <p className="text-text_color/65 mt-1 text-sm">
                    End the session on this device.
                  </p>
                </div>
                <Button
                  styles="shrink-0 px-4 py-2"
                  variant="tertiary"
                  buttonText="Sign out"
                  onClick={() => logoutHandler(false)}
                >
                  <Icon icon={Logout05Icon} styles="mr-2 size-5" />
                </Button>
              </div>
              <div
                className="flex flex-col items-start justify-between gap-3 pt-4
                  sm:flex-row sm:items-center"
              >
                <div>
                  <p className="font-medium">Sign out everywhere</p>
                  <p className="text-text_color/65 mt-1 text-sm">
                    End every active session, including this one.
                  </p>
                </div>
                <Button
                  styles="shrink-0 px-4 py-2"
                  variant="danger"
                  buttonText="Sign out everywhere"
                  onClick={() => logoutHandler(true)}
                />
              </div>
            </div>
          </Card>
          <Card styles="border-red-300 dark:border-red-900">
            <p className="text-lg font-medium text-red-600 dark:text-red-400">
              Delete account
            </p>
            <p className="text-text_color/65 mt-1 mb-4 text-sm">
              Permanently delete your personal account and its data.
            </p>
            <DeleteModal
              onDelete={deleteHandler}
              itemName="your user account"
              trigger={
                <Button
                  styles="px-4 py-2"
                  variant="danger"
                  buttonText="Delete account"
                />
              }
            />
          </Card>
        </div>
      </div>
    </div>
  );
}
