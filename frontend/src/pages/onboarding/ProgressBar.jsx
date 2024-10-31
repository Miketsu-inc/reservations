import { Fragment } from "react";
import ProgressBarStep from "./ProgressBarStep";

export default function ProgressBar({ currentStep, stepCount, isSubmitDone }) {
  return (
    <div className="mb-8 mt-6 flex items-center justify-center sm:mt-4">
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
                  : "flex-auto border-t-4 border-gray-600"
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
