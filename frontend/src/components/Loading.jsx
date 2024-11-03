import LoaderIcon from "../assets/icons/LoaderIcon";

export default function Loading() {
  return (
    <div className="flex flex-col items-center gap-4 p-3">
      <LoaderIcon styles="w-12 h-12 text-primary" />
      <p className="text-text_color">Loading...</p>
    </div>
  );
}
