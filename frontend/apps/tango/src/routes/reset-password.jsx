import { Button, Input } from "@reservations/components";
import { useToast } from "@reservations/lib";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";

export const Route = createFileRoute("/reset-password")({
  component: RouteComponent,
});

function RouteComponent() {
  const { token } = Route.useSearch();
  const navigate = Route.useNavigate();
  const [passwords, setPasswords] = useState({
    password: "",
    confirmPassword: "",
  });
  const { showToast } = useToast();

  function updatePasswords(data) {
    setPasswords((prev) => ({ ...prev, ...data }));
  }

  const doPasswordsMatch =
    passwords.password != "" &&
    passwords.password === passwords.confirmPassword;

  async function submitHandler(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    const response = await fetch("/api/v1/auth/reset-password", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify({
        token: token,
        password: passwords.password,
      }),
    });

    if (!response.ok) {
      const result = await response.json();
      showToast({ message: result.error.message, variant: "error" });
    } else {
      showToast({
        message: "Password successfully updated",
        variant: "success",
      });
      navigate({
        from: Route.fullPath,
        to: "/login",
      });
    }
  }

  if (!token) {
    return (
      <div className="flex min-h-screen items-center justify-center px-4 py-10">
        <div
          className="sm:bg-bg-raised sm:border-border sm:shadow-card flex w-full
            max-w-xl flex-col sm:rounded-2xl sm:border sm:p-8"
        >
          <div>
            <h1 className="text-xl font-semibold">
              Invalid token query parameter
            </h1>
            <Link
              from={Route.fullPath}
              to="/login"
              className="text-fg-primary mt-2 inline-block text-sm font-medium
                hover:underline"
            >
              Back to login
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center px-4 py-10">
      <div
        className="sm:bg-bg-raised sm:border-border sm:shadow-card flex w-full
          max-w-xl flex-col sm:rounded-2xl sm:border sm:p-8"
      >
        <div className="pb-6">
          <h1 className="text-2xl font-semibold">Reset password</h1>
          <p className="text-fg-muted mt-1 text-sm">Enter your new password</p>
        </div>
        <form onSubmit={submitHandler} className="flex h-full flex-col gap-5">
          <Input
            name="NewPassword"
            type="password"
            labelText="New password"
            inputData={(data) => updatePasswords({ password: data.value })}
            value={passwords.password}
          />
          <Input
            name="ConfirmNewPassword"
            type="password"
            labelText="Confirm new password"
            inputData={(data) =>
              updatePasswords({ confirmPassword: data.value })
            }
            value={passwords.confirmPassword}
          />
          <Button
            styles="py-2.5"
            type="submit"
            buttonText="Reset password"
            disabled={!doPasswordsMatch}
          />
        </form>
      </div>
    </div>
  );
}
