import { createContext, useState } from "react";
import ToastElement from "./ToastElement";

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
        className="fixed top-4 z-10 flex w-full flex-col-reverse gap-4 px-6 transition-all
          duration-1000 sm:top-auto sm:right-4 sm:bottom-4 sm:w-auto sm:flex-col sm:px-2"
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
