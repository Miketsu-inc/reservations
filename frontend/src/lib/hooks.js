import { useEffect, useState } from "react";
import { getBreakPoint } from "./utils";

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

export function useWindowSize() {
  const isWindowClient = typeof window === "object";
  const [windowSize, setWindowSize] = useState(
    isWindowClient ? window.innerWidth : undefined
  );

  useEffect(() => {
    if (isWindowClient) {
      function setSize() {
        setWindowSize(window.innerWidth);
      }

      window.addEventListener("resize", setSize);

      return () => window.removeEventListener("resize", setSize);
    }
  }, [isWindowClient, setWindowSize]);

  return getBreakPoint(windowSize);
}

export function useClickOutside(ref, callback) {
  useEffect(() => {
    function clickOutsideHandler(e) {
      if (ref.current && !ref.current.contains(e.target)) {
        callback();
      }
    }

    document.addEventListener("mousedown", clickOutsideHandler);
    return () => document.removeEventListener("mousedown", clickOutsideHandler);
  });
}
