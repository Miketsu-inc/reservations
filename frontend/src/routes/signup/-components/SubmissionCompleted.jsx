import TickIcon from "@icons/TickIcon";

export default function SubmissionCompleted({ text }) {
  return (
    <div className="flex flex-col items-center gap-10">
      <div className="my-4 mt-10 rounded-full border-4 border-green-600 p-4">
        <TickIcon styles="w-16 w-16 fill-green-600" />
      </div>
      <p className="text-xl text-text_color">{text}</p>
    </div>
  );
}
