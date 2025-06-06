import TickIcon from "@icons/TickIcon";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/m/$merchantName/cancel/$appointmentId/completed"
)({
  component: CancelCompleted,
});

function CancelCompleted() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center">
      <div className="text-center">
        <div
          className="mx-auto mb-6 flex h-24 w-24 items-center justify-center rounded-full
            bg-green-500"
        >
          <TickIcon className="h-12 w-12" />
        </div>
        <h1 className="text-text_color mb-4 text-3xl font-bold">All Done!</h1>
        <p className="mb-2 text-lg text-gray-600 dark:text-gray-300">
          Your appointment has been cancelled
        </p>
      </div>
    </div>
  );
}
