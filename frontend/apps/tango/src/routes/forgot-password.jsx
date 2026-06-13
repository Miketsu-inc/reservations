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
    <div className="flex h-screen items-center justify-center px-4">
      <div className="flex w-full max-w-xl flex-col">
        <div className="pt-4 pb-12">
          <p className="text-2xl">Forgot password</p>
          <p className="text-text_color/60">
            You will get an email with a link to reset your password
          </p>
        </div>
        <form onSubmit={submitHandler} className="flex h-full flex-col gap-6">
          <Input
            styles="p-2"
            name="email"
            labelText="Email"
            type="email"
            required={true}
            inputData={(data) => setEmail(data.value)}
            value={email}
          />
          <Button styles="p-2" type="submit" buttonText="Send email" />
        </form>
      </div>
    </div>
  );
}
