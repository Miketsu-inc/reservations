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
      className={`${styles} w-full rounded-lg bg-layer_bg text-text_color shadow-md
        shadow-layer_bg transition-all backdrop:bg-black backdrop:bg-opacity-35 md:w-fit`}
      ref={modalRef}
    >
      {children}
    </dialog>
  );
}
