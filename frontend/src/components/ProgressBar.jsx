import TickIcon from "@icons/TickIcon";
import { Fragment } from "react";

export default function ProgressBar({ currentStep, stepCount, isSubmitDone }) {
  return (
    <div className="mt-6 mb-8 flex items-center justify-center sm:mt-4">
      {[...Array(stepCount)].map((_, i) => (
        <Fragment key={i}>
          <ProgressBarStep
            step={i + 1}
            isActive={currentStep === i}
            isCompleted={isSubmitDone ? true : currentStep > i}
          />
          {i !== stepCount - 1 ? (
            <div
              className={
                currentStep > i
                  ? "flex-auto border-t-4 border-green-700 transition-all duration-700 ease-in"
                  : "flex-auto border-t-4 border-gray-400 dark:border-gray-600"
              }
            ></div>
          ) : (
            <></>
          )}
        </Fragment>
      ))}
    </div>
  );
}

function ProgressBarStep({ step, isActive, isCompleted }) {
  return (
    <div
      className={
        isCompleted
          ? `relative flex h-8 w-8 items-center justify-center rounded-full border-[3px]
            border-green-700 bg-green-700 p-4 transition-all duration-500 ease-in`
          : isActive
            ? `border-primary/70 relative flex h-8 w-8 items-center justify-center rounded-full
              border-[3px] p-4 transition-all duration-700 ease-in`
            : `relative flex h-8 w-8 items-center justify-center rounded-full border-[3px]
              border-gray-400 p-4 dark:border-gray-600`
      }
    >
      {isCompleted ? (
        <div>
          <TickIcon styles="fill-white h-5 w-5" />
        </div>
      ) : (
        `${step}`
      )}
    </div>
  );
}
