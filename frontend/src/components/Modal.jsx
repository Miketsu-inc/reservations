import { useClickOutside } from "@lib/hooks";
import { useEffect, useRef } from "react";

export default function Modal({ styles, isOpen, onClose, children }) {
  const modalRef = useRef();
  useClickOutside(modalRef, onClose);

  useEffect(() => {
    isOpen ? modalRef.current.showModal() : modalRef.current.close();
  }, [isOpen]);

  return (
    <dialog
      className={`${styles} bg-layer_bg text-text_color w-full place-self-center rounded-lg
        shadow-md shadow-gray-400 transition-all backdrop:bg-black/45 md:w-fit
        dark:border dark:border-gray-600 dark:shadow-xs dark:shadow-gray-800`}
      ref={modalRef}
    >
      {children}
    </dialog>
  );
}
