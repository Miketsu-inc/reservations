import { Tick02Icon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";

export default function SubmissionCompleted({ text }) {
  return (
    <div className="flex flex-col items-center gap-10">
      <div className="border-success my-4 mt-10 rounded-full border-4 p-4">
        <Icon icon={Tick02Icon} styles="w-16 text-fg-success" />
      </div>
      <p className="text-fg text-xl">{text}</p>
    </div>
  );
}
