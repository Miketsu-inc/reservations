import { Tick02Icon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/m/$merchantName/book/completed")({
  component: BookingCompleted,
});

function BookingCompleted() {
  return (
    <div
      className="flex min-h-screen flex-col items-center justify-center gap-10"
    >
      <div className="border-success rounded-full border-4 p-4">
        <Icon icon={Tick02Icon} styles="size-16 text-fg-success" />
      </div>
      <p className="text-fg text-xl">Successful booking</p>
    </div>
  );
}
