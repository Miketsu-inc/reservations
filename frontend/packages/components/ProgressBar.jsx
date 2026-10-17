import { Tick02Icon } from "@hugeicons/core-free-icons";
import { Fragment } from "react";
import Icon from "./Icon.jsx";

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
                  ? `border-primary flex-auto border-t-4 transition-all
                    duration-700 ease-in`
                  : "border-bg-inactive flex-auto border-t-4"
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
          ? `border-primary bg-primary relative flex h-8 w-8 items-center
            justify-center rounded-full border-[3px] p-4 transition-all
            duration-500 ease-in`
          : isActive
            ? `border-primary text-fg-primary relative flex h-8 w-8 items-center
              justify-center rounded-full border-[3px] p-4 transition-all
              duration-700 ease-in`
            : `text-fg-muted border-bg-inactive relative flex h-8 w-8
              items-center justify-center rounded-full border-[3px] p-4`
      }
    >
      {isCompleted ? (
        <div>
          <Icon icon={Tick02Icon} styles="text-white size-5" />
        </div>
      ) : (
        `${step}`
      )}
    </div>
  );
}
