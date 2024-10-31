import { useEffect, useState } from "react";
import "./autofill/detect-autofill";
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
    stepCount: steps.length,
  };
}

export function useWindowSize() {
  const isWindowClient = typeof window === "object";
  const [windowSize, setWindowSize] = useState(
    isWindowClient ? window.innerWidth : undefined
  );

  function setSize() {
    setWindowSize(window.innerWidth);
  }

  useEffect(() => {
    if (isWindowClient) {
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

export function useAutofill(ref, callback) {
  useEffect(() => {
    const input = ref.current;
    if (!input) return;

    function onAutofill() {
      const event = new CustomEvent("autofillEvent", {
        detail: { isAutofillEvent: true },
      });
      Object.defineProperty(event, "target", {
        writable: false,
        value: {
          value: input.value,
        },
      });

      callback(event);
    }

    input.addEventListener("onautocomplete", onAutofill);
    return () => input.removeEventListener("onautocomplete", onAutofill);
  }, [ref, callback]);
}
