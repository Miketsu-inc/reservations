import { LoaderIcon } from "@reservations/assets";

export default function Loading() {
  return (
    <div className="flex flex-col items-center gap-4 p-3">
      <LoaderIcon styles="w-12 h-12 text-fg-primary" />
      <p className="text-fg">Loading...</p>
    </div>
  );
}
