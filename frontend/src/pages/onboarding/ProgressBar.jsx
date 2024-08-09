import ProgressBarStep from "./ProgressBarStep";

export default function ProgressBar({ step }) {
  return (
    <div className="mb-8 mt-6 flex items-center justify-center sm:mt-4">
      <ProgressBarStep
        step="1"
        stepName="Name"
        isActive={step === 0}
        isCompleted={step > 0}
      />
      <div
        className={
          step > 0
            ? "flex-auto border-t-2 border-green-700 transition-all"
            : "flex-auto border-t-2 border-gray-400"
        }
      ></div>
      <ProgressBarStep
        step="2"
        stepName="Email"
        isActive={step === 1}
        isCompleted={step > 1}
      />
      <div
        className={
          step > 1
            ? "flex-auto border-t-2 border-green-700 transition-all"
            : "flex-auto border-t-2 border-gray-400"
        }
      ></div>
      <ProgressBarStep
        step="3"
        stepName="Password"
        isActive={step === 2}
        isCompleted={step > 2}
      />
    </div>
  );
}
