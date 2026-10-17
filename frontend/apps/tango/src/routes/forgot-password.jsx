import { Button, Input } from "@reservations/components";
import { useToast } from "@reservations/lib";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";

export const Route = createFileRoute("/forgot-password")({
  component: RouteComponent,
});

function RouteComponent() {
  const [email, setEmail] = useState("");
  const { showToast } = useToast();

  async function submitHandler(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    const response = await fetch("/api/v1/auth/forgot-password", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify({
        email: email,
      }),
    });

    if (!response.ok) {
      const result = await response.json();
      showToast({ message: result.error.message, variant: "error" });
    } else {
      showToast({
        message: "Email has been sent",
        variant: "success",
      });
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center px-4 py-10">
      <div
        className="sm:bg-bg-raised sm:border-border sm:shadow-card flex w-full
          max-w-xl flex-col sm:rounded-2xl sm:border sm:p-8"
      >
        <div className="pb-10">
          <h1 className="text-2xl font-semibold tracking-tight">
            Forgot password
          </h1>
          <p className="text-fg-muted mt-1 text-sm">
            You will get an email with a link to reset your password
          </p>
        </div>
        <form onSubmit={submitHandler} className="flex h-full flex-col gap-5">
          <Input
            name="email"
            labelText="Email"
            type="email"
            required={true}
            inputData={(data) => setEmail(data.value)}
            value={email}
          />
          <Button styles="py-2.5" type="submit" buttonText="Send email" />
        </form>
      </div>
    </div>
  );
}
