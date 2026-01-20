import { InfoIcon, TickIcon, WarningIcon, XIcon } from "@reservations/assets";
import { createContext, useCallback, useEffect, useState } from "react";

export const ToastContext = createContext();

export default function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([]);

  const showToast = ({ message, variant, duration }) => {
    const id = Date.now();
    setToasts((prev) => [...prev, { id, message, variant, duration }]);
  };

  function removeToast(id) {
    setToasts((prev) => prev.filter((toast) => toast.id !== id));
  }

  return (
    // sharing the showToast function globally
    <ToastContext.Provider value={{ showToast }}>
      {children}
      <div
        className="fixed top-4 z-50 flex w-full flex-col-reverse gap-4 px-6
          transition-all duration-1000 sm:top-auto sm:right-4 sm:bottom-4
          sm:w-auto sm:flex-col sm:px-2"
      >
        {toasts.map((toast) => (
          <ToastElement
            key={toast.id}
            variant={toast.variant}
            message={toast.message}
            duration={toast.duration}
            onClose={() => removeToast(toast.id)}
          />
        ))}
      </div>
    </ToastContext.Provider>
  );
}

const icons = {
  success: (
    <div
      className="w-min rounded-full border border-green-700
        dark:border-green-600"
    >
      <TickIcon styles="size-6 dark:fill-green-600 fill-green-700" />
    </div>
  ),
  error: (
    <div
      className="w-min rounded-full border border-red-700 dark:border-red-600"
    >
      <XIcon styles="size-6 dark:fill-red-600 fill-red-700" />
    </div>
  ),
  warning: (
    <div className="w-min rounded-full">
      <WarningIcon styles="size-6 fill-yellow-600" />
    </div>
  ),
  info: <InfoIcon styles="size-7 text-blue-600 stroke-blue-600" />,
};

const typeStyles = {
  success: "border dark:border-green-500 border-green-700 bg-green-50",
  error: "border dark:border-red-500 border-red-700 bg-red-50",
  warning: "border border-yellow-600 bg-yellow-50",
  info: "border border-blue-500 bg-blue-50",
};

function ToastElement({ variant, message, onClose, duration = 5000 }) {
  const [fadingOut, setFadingOut] = useState(false);

  const startFadeOut = useCallback(() => {
    setFadingOut(true);
    setTimeout(onClose, 600); // Call onClose after the animation completes (600ms)
  }, [onClose]);

  useEffect(() => {
    const timer = setTimeout(startFadeOut, duration - 600); // Start fading out 600ms before the toast is removed
    return () => clearTimeout(timer);
  }, [duration, startFadeOut]);

  return (
    <div
      className={`${typeStyles[variant]} flex w-full items-center
        justify-between rounded-md p-4 shadow-lg transition-all sm:max-w-md
        dark:bg-gray-900 dark:shadow-none
        ${fadingOut ? "toast-exit" : window.innerWidth < 640 ? "toast-enter-top" : "toast-enter-bottom"}`}
    >
      <div className="mr-6 flex items-center gap-3">
        <div>{icons[variant]}</div>
        <span className="grow text-sm text-gray-800 dark:text-gray-400">
          {message}
        </span>
      </div>
      <div
        className="group dark:hover:bg-hvr_gray rounded-lg hover:bg-gray-200/25"
      >
        <XIcon
          onClick={onClose}
          styles="size-6 dark:fill-gray-400 m-1 fill-gray-500
            group-hover:fill-current cursor-pointer"
        />
      </div>
    </div>
  );
}
