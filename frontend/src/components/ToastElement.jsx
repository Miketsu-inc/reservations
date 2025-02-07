import InfoIcon from "@icons/InfoIcon";
import TickIcon from "@icons/TickIcon";
import WarningIcon from "@icons/WarningIcon";
import XIcon from "@icons/XIcon";
import { useCallback, useEffect, useState } from "react";

const icons = {
  success: (
    <div className="w-min rounded-full border border-green-700 dark:border-green-600">
      <TickIcon styles="h-6 w-6 dark:fill-green-600 fill-green-700" />
    </div>
  ),
  error: (
    <div className="w-min rounded-full border border-red-700 dark:border-red-600">
      <XIcon styles="h-6 w-6 dark:fill-red-600 fill-red-700" />
    </div>
  ),
  warning: (
    <div className="w-min rounded-full">
      <WarningIcon styles="h-6 w-6 fill-yellow-600" />
    </div>
  ),
  info: <InfoIcon styles="h-7 w-7 text-blue-600 stroke-blue-600" />,
};

const typeStyles = {
  success: "border dark:border-green-500 border-green-700 bg-green-50",
  error: "border dark:border-red-500 border-red-700 bg-red-50",
  warning: "border border-yellow-600 bg-yellow-50",
  info: "border border-blue-500 bg-blue-50",
};

export default function ToastElement({
  variant,
  message,
  onClose,
  duration = 5000,
}) {
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
      className={`${typeStyles[variant]} flex w-full items-center justify-between rounded-md p-4
        shadow-lg transition-all sm:max-w-md dark:bg-gray-900 dark:shadow-none
        ${fadingOut ? "toast-exit" : window.innerWidth < 640 ? "toast-enter-top" : "toast-enter-bottom"}`}
    >
      <div className="mr-6 flex items-center gap-3">
        <div>{icons[variant]}</div>
        <span className="flex-grow text-sm text-gray-800 dark:text-gray-400">
          {message}
        </span>
      </div>
      <div className="group dark:hover:bg-hvr_gray rounded-lg hover:bg-gray-200/25">
        <XIcon
          onClick={onClose}
          styles="h-6 w-6 dark:fill-gray-400 m-1 fill-gray-500 group-hover:fill-current
            cursor-pointer"
        />
      </div>
    </div>
  );
}
