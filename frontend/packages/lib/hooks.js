import { ToastContext } from "@reservations/components/ToastProvider.jsx";
import {
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import "./autofill/detect-autofill";
import { SCREEN_LG } from "./constants";
import { invalidateLocalStorageAuth } from "./lib";
import {
  ACCEPTED_IMAGE_TYPES,
  preloadImages,
  uploadImage,
  validateImageFile,
} from "./media";
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

export function useClickOutside(ref, callback, enabled = true) {
  const callbackRef = useRef(callback);

  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

  useEffect(() => {
    if (!enabled) return undefined;

    function clickOutsideHandler(e) {
      if (ref.current && !ref.current.contains(e.target)) {
        callbackRef.current?.(e);
      }
    }

    document.addEventListener("mousedown", clickOutsideHandler);
    return () => document.removeEventListener("mousedown", clickOutsideHandler);
  }, [enabled, ref]);
}

export function useAutofill(ref, callback) {
  const callbackRef = useRef(callback);

  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

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

      callbackRef.current?.(event);
    }

    input.addEventListener("onautocomplete", onAutofill);
    return () => input.removeEventListener("onautocomplete", onAutofill);
  }, [ref]);
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
    if (sectionIds.length === 0) return undefined;

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

  return sectionIds.includes(activeSection) ? activeSection : sectionIds[0];
}

export function useCroppedImageUpload({ img, url, label, onChange }) {
  const inputRef = useRef(null);
  const [isDragging, setIsDragging] = useState(false);
  const [cropping, setCropping] = useState(null);
  const [pending, setPending] = useState(null);
  const [isSaving, setIsSaving] = useState(false);
  const { showToast } = useToast();

  if (pending && img !== pending.replacedImg) {
    setPending(null);
  }

  useEffect(() => {
    if (!pending) return;

    return () => URL.revokeObjectURL(pending.url);
  }, [pending]);

  function openFilePicker() {
    inputRef.current?.click();
  }

  function selectFile(selected) {
    const error = validateImageFile(selected);
    if (error) {
      showToast({ message: error, variant: "error" });
      return;
    }

    setCropping({ file: selected, url: URL.createObjectURL(selected) });
  }

  function cancelCrop() {
    URL.revokeObjectURL(cropping.url);
    setCropping(null);
  }

  async function save(crop, area) {
    setCropping(null);
    setPending({ url: cropping.url, area, replacedImg: img });
    setIsSaving(true);

    try {
      const image = await uploadImage(url, cropping.file, crop);
      if (image) {
        await preloadImages(image.url, image.thumbnail_url);
      } else {
        showToast({
          message: `Your ${label} is still processing, it will show up shortly`,
          variant: "info",
        });
      }
      await onChange?.();
    } catch (error) {
      showToast({ message: error.message, variant: "error" });
      setPending(null);
    } finally {
      setIsSaving(false);
    }
  }

  async function remove() {
    try {
      const response = await fetch(url, {
        method: "DELETE",
        headers: {
          Accept: "application/json",
        },
      });

      if (!response.ok) {
        const result = await response.json();
        invalidateLocalStorageAuth(response.status);
        showToast({ message: result.error.message, variant: "error" });
      } else {
        await onChange?.();
      }
    } catch (error) {
      showToast({ message: error.message, variant: "error" });
    }
  }

  return {
    isDragging,
    isSaving,
    pendingStyle: pending ? cropPreviewStyle(pending.url, pending.area) : null,
    openFilePicker,
    remove,
    inputProps: {
      ref: inputRef,
      type: "file",
      accept: ACCEPTED_IMAGE_TYPES.join(","),
      className: "hidden",
      onChange: (e) => {
        const selected = e.target.files?.[0];
        // allows selecting the same file again
        e.target.value = "";
        if (selected) selectFile(selected);
      },
    },
    dropProps: {
      onDragOver: (e) => {
        e.preventDefault();
        setIsDragging(!isSaving);
      },
      onDragLeave: (e) => {
        if (!e.currentTarget.contains(e.relatedTarget)) {
          setIsDragging(false);
        }
      },
      onDrop: (e) => {
        e.preventDefault();
        setIsDragging(false);
        const file = e.dataTransfer.files[0];
        if (file && !isSaving) selectFile(file);
      },
    },
    cropDialogProps: {
      open: cropping !== null,
      onOpenChange: (open) => !open && cancelCrop(),
      imageSrc: cropping?.url,
      title: `Crop ${label}`,
      onSave: save,
    },
  };
}

function cropPreviewStyle(url, area) {
  const position = (offset, size) =>
    size < 100 ? (offset / (100 - size)) * 100 : 0;

  return {
    backgroundImage: `url(${url})`,
    backgroundSize: `${10000 / area.width}% ${10000 / area.height}%`,
    backgroundPosition: `${position(area.x, area.width)}% ${position(area.y, area.height)}%`,
  };
}
