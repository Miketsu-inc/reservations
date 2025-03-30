import { useClickOutside } from "@lib/hooks";
import { useEffect, useRef } from "react";

export default function Modal({
  styles,
  zindex = 40,
  suspendCloseOnClickOutside = false,
  isOpen,
  onClose,
  children,
}) {
  const modalRef = useRef();
  useClickOutside(modalRef, suspendCloseOnClickOutside ? () => {} : onClose);

  useEffect(() => {
    if (isOpen) {
      modalRef.current.focus();

      // TODO: Temporarily disable focustrap because with stacked modals
      // it steals the focus away from an input. Should be enabled
      // once stacked modals are eliminated.
      //
      // const focusOutHandler = (event) => {
      //   if (!modalRef.current?.contains(event.relatedTarget))
      //     modalRef.current?.focus();
      // };

      // modalRef.current.addEventListener("focusout", focusOutHandler);

      // return () => {
      //   document.removeEventListener("focusout", focusOutHandler);
      // };
    }
  }, [isOpen, onClose]);

  return (
    <>
      {isOpen && (
        <>
          <div
            aria-hidden="true"
            className={"fixed inset-0 bg-black/45"}
            style={{ zIndex: zindex }}
          ></div>
          <div
            className="fixed inset-0 flex w-full items-center justify-center p-4"
            style={{ zIndex: zindex }}
          >
            <div
              role="dialog"
              aria-modal="true"
              tabIndex={-1}
              className={`${styles} bg-layer_bg text-text_color w-full rounded-lg shadow-md
              shadow-gray-400 transition-all focus:outline-none md:w-fit dark:border
              dark:border-gray-600 dark:shadow-xs dark:shadow-gray-800`}
              ref={modalRef}
            >
              {children}
            </div>
          </div>
          {/* This is needed to trap focus and make tabbing loop */}
          <span
            aria-hidden="true"
            tabIndex={0}
            className="pointer-events-none fixed opacity-0 outline-none"
          ></span>
        </>
      )}
    </>
  );
}
