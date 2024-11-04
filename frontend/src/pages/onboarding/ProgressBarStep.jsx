import TickIcon from "../../assets/icons/TickIcon";

export default function ProgressBarStep({ step, isActive, isCompleted }) {
  return (
    <div
      className={
        isCompleted
          ? `relative flex h-8 w-8 items-center justify-center rounded-full border-[3px]
            border-green-700 bg-green-700 p-4 transition-all duration-500 ease-in`
          : isActive
            ? `relative flex h-8 w-8 items-center justify-center rounded-full border-[3px]
              border-primary/70 p-4 transition-all duration-700 ease-in`
            : `relative flex h-8 w-8 items-center justify-center rounded-full border-[3px]
              border-gray-600 p-4`
      }
    >
      {isCompleted ? (
        <div>
          <TickIcon styles="fill-white" />
        </div>
      ) : (
        `${step}`
      )}
    </div>
  );
}
