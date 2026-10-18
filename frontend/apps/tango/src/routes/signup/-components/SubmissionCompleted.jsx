import { Tick02Icon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";

export default function SubmissionCompleted({ text }) {
  return (
    <div className="flex flex-col items-center gap-4 py-6">
      <div className="bg-success-subtle rounded-full p-3">
        <Icon icon={Tick02Icon} styles="size-8 text-fg-success" />
      </div>
      <p className="text-fg text-center text-lg font-semibold">{text}</p>
    </div>
  );
}
