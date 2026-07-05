import { ToastContext } from "@reservations/components";
import {
  useCallback,
  useContext,
  useEffect,
  useState,
  useSyncExternalStore,
} from "react";
import "./autofill/detect-autofill";
import { SCREEN_LG } from "./constants";
import { getBreakPoint } from "./utils";

export function useMultiStepForm(steps) {
  const [stepIndex, setStepIndex] = useState(0);

  function next() {
    setStepIndex((i) => {
      if (i >= steps.length - 1) return i;
      return i + 1;
    });
  }

  function previous() {
    setStepIndex((i) => {
      if (i <= 0) return i;
      return i - 1;
    });
  }

  return {
    stepIndex,
    step: steps[stepIndex],
    nextStep: next,
    stepCount: steps.length,
    previousStep: previous,
  };
}

function getWindowWidthSnapshot() {
  return window.innerWidth;
}

function subscribeWindowWidth(callback) {
  window.addEventListener("resize", callback);
  return () => window.removeEventListener("resize", callback);
}

export function useWindowSize() {
  const width = useSyncExternalStore(
    subscribeWindowWidth,
    getWindowWidthSnapshot
  );

  return {
    width,
    windowSize: getBreakPoint(width),
    isWindowSmall: width < SCREEN_LG,
  };
}

export function useClickOutside(ref, callback) {
  useEffect(() => {
    function clickOutsideHandler(e) {
      if (ref.current) {
        if (!ref.current.contains(e.target) || e.target === ref.current) {
          callback();
        }
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

export function useToast() {
  return useContext(ToastContext);
}

export function useTheme() {
  const [isDarkTheme, setIsDarkTheme] = useState(
    document.documentElement.classList.contains("dark")
  );

  const switchTheme = useCallback(() => {
    window.toggleTheme();

    setIsDarkTheme(document.documentElement.classList.contains("dark"));
  }, []);

  useEffect(() => {
    function handleSystemThemeChange() {
      if (!localStorage.getItem("theme")) {
        setIsDarkTheme(document.documentElement.classList.contains("dark"));
      }
    }

    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    mediaQuery.addEventListener("change", handleSystemThemeChange);

    return () => {
      mediaQuery.removeEventListener("change", handleSystemThemeChange);
    };
  }, []);

  return {
    switchTheme,
    isDarkTheme,
  };
}

export function useActiveSection(sectionIds) {
  const [activeSection, setActiveSection] = useState(sectionIds[0]);

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setActiveSection(entry.target.id);
          }
        });
      },
      { rootMargin: "-75% 0px -25% 0px" }
    );

    sectionIds.forEach((id) => {
      const element = document.getElementById(id);
      if (element) observer.observe(element);
    });

    return () => observer.disconnect();
  }, [sectionIds]);

  return activeSection;
}
