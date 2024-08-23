import TickIcon from "../../assets/TickIcon";

export default function SubmissionCompleted({ text }) {
  return (
    <div className="flex flex-col items-center justify-center">
      <div className="my-4 mt-10 rounded-full border-4 border-green-600 p-6">
        <TickIcon height="60" width="60" styles="fill-green-600" />
      </div>
      <div className="mt-10 text-center text-xl font-semibold">{text}</div>
    </div>
  );
}
