import { Tick02Icon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/m/$merchantName/cancel/$bookingId/completed"
)({
  component: CancelCompleted,
});

function CancelCompleted() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center">
      <div className="text-center">
        <div
          className="bg-success mx-auto mb-6 flex h-24 w-24 items-center
            justify-center rounded-full"
        >
          <Icon icon={Tick02Icon} className="size-12" />
        </div>
        <h1 className="text-fg mb-4 text-3xl font-bold">All Done!</h1>
        <p className="text-fg-muted mb-2 text-lg">
          Your booking has been cancelled
        </p>
      </div>
    </div>
  );
}
