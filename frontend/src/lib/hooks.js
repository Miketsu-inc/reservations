import { useState } from "react";

export function useMultiStepForm(steps) {
  const [stepIndex, setStepIndex] = useState(0);

  function next() {
    setStepIndex((i) => {
      if (i >= steps.length - 1) return i;
      return i + 1;
    });
  }

  return {
    stepIndex,
    step: steps[stepIndex],
    nextStep: next,
  };
}
