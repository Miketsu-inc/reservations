import { TickIcon } from "@reservations/assets";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/m/$merchantName/booking/completed")({
  component: BookingCompleted,
});

function BookingCompleted() {
  return (
    <div
      className="flex min-h-screen flex-col items-center justify-center gap-10"
    >
      <div className="rounded-full border-4 border-green-600 p-4">
        <TickIcon styles="h-16 w-16 fill-green-600" />
      </div>
      <p className="text-text_color text-xl">Successful booking</p>
    </div>
  );
}
