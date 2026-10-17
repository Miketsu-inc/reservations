import {
  Alert02Icon,
  Cancel01Icon,
  InformationCircleIcon,
  Tick02Icon,
} from "@hugeicons/core-free-icons";
import {
  createContext,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { createPortal } from "react-dom";
import Icon from "./Icon.jsx";

export const ToastContext = createContext();

export default function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([]);
  const nextId = useRef(0);

  const showToast = useCallback(({ message, variant = "info", duration }) => {
    const id = nextId.current++;
    setToasts((prev) => [...prev, { id, message, variant, duration }]);
  }, []);

  const removeToast = useCallback((id) => {
    setToasts((prev) => prev.filter((toast) => toast.id !== id));
  }, []);

  const contextValue = useMemo(() => ({ showToast }), [showToast]);

  return (
    // sharing the showToast function globally
    <ToastContext.Provider value={contextValue}>
      {children}
      {createPortal(
        <div
          className="fixed top-4 z-70 flex w-full flex-col-reverse gap-4 px-6
            transition-all duration-1000 sm:top-auto sm:right-4 sm:bottom-4
            sm:w-auto sm:flex-col sm:px-2"
        >
          {toasts.map((toast) => (
            <ToastElement
              key={toast.id}
              variant={toast.variant}
              message={toast.message}
              duration={toast.duration}
              onClose={removeToast}
              toastId={toast.id}
            />
          ))}
        </div>,
        document.body
      )}
    </ToastContext.Provider>
  );
}

const icons = {
  success: (
    <div className="bg-success-subtle w-min rounded-full p-0.5">
      <Icon icon={Tick02Icon} styles="size-5 text-fg-success" />
    </div>
  ),
  error: (
    <div className="bg-danger-subtle w-min rounded-full p-0.5">
      <Icon icon={Cancel01Icon} styles="size-5 text-fg-danger" />
    </div>
  ),
  warning: (
    <div className="w-min rounded-full">
      <Icon icon={Alert02Icon} styles="size-6 text-fg-warning" />
    </div>
  ),
  info: <Icon icon={InformationCircleIcon} styles="size-7 text-fg-info" />,
};

function ToastElement({ variant, message, onClose, toastId, duration = 5000 }) {
  const [fadingOut, setFadingOut] = useState(false);

  useEffect(() => {
    let closeTimer;
    const fadeTimer = setTimeout(() => {
      setFadingOut(true);
      closeTimer = setTimeout(() => onClose(toastId), 600);
    }, duration - 600);

    return () => {
      clearTimeout(fadeTimer);
      clearTimeout(closeTimer);
    };
  }, [duration, onClose, toastId]);

  return (
    <div
      role={variant === "error" ? "alert" : "status"}
      className={`bg-bg shadow-overlay flex w-full items-center justify-between
        rounded-xl border p-3 transition-all sm:max-w-md
        ${fadingOut ? "toast-exit" : window.innerWidth < 640 ? "toast-enter-top" : "toast-enter-bottom"}`}
    >
      <div className="mr-6 flex items-center gap-3">
        <div>{icons[variant]}</div>
        <span className="text-fg grow text-sm">{message}</span>
      </div>
      <button
        type="button"
        aria-label="Dismiss notification"
        onClick={() => onClose(toastId)}
        className="group hover:bg-hover rounded-lg"
      >
        <Icon
          icon={Cancel01Icon}
          styles="size-5 m-1 text-fg-muted group-hover:text-fg cursor-pointer"
        />
      </button>
    </div>
  );
}
