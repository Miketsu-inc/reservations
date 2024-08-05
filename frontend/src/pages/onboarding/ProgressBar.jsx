import ProgressBarStep from "./ProgressBarStep";

export default function ProgressBar({ page, submitted }) {
  return (
    <div className="mb-8 mt-6 flex items-center justify-center sm:mt-4">
      <ProgressBarStep
        step="1"
        stepName="Name"
        isActive={page === 0}
        isCompleted={page > 0}
      />
      <div
        className={
          page > 0
            ? "flex-auto border-t-2 border-green-700 transition-all"
            : "flex-auto border-t-2 border-gray-400"
        }
      ></div>
      <ProgressBarStep
        step="2"
        stepName="Email"
        isActive={page === 1}
        isCompleted={page > 1}
      />
      <div
        className={
          page > 1
            ? "flex-auto border-t-2 border-green-700 transition-all"
            : "flex-auto border-t-2 border-gray-400"
        }
      ></div>
      <ProgressBarStep
        step="3"
        stepName="Password"
        isActive={page === 2}
        isCompleted={submitted}
      />
    </div>
  );
}
