import { MessageIcon } from "@reservations/assets";
import { useEffect, useRef, useState } from "react";

export default function ExpandableNote({ text }) {
  const [showFullNote, setShowFullNote] = useState(false);
  const [isNoteOverflowing, setIsNoteOverflowing] = useState(false);
  const NoteRef = useRef(null);

  useEffect(() => {
    const checkOverflow = () => {
      const element = NoteRef.current;
      if (!element) return;

      const isOverflowing = element.scrollWidth > element.clientWidth;
      setIsNoteOverflowing(isOverflowing);
    };

    checkOverflow();
    window.addEventListener("resize", checkOverflow);
    return () => {
      window.removeEventListener("resize", checkOverflow);
    };
  }, []);

  return (
    <>
      {text && (
        <div
          className="text-text_color/70 flex w-full items-start justify-start
            gap-3 pr-2 sm:pr-4"
        >
          <MessageIcon
            styles="size-4 shrink-0 mt-1 stroke-text_color/50
              fill-text_color/50"
          />

          {showFullNote ? (
            <div className="min-w-0 flex-1 text-sm">
              <span>Note: {text}</span>
              <button
                className="text-secondary ml-2 cursor-pointer text-xs"
                onClick={() => setShowFullNote(false)}
              >
                Show less
              </button>
            </div>
          ) : (
            <div className="flex min-w-0 items-baseline gap-1">
              <span
                ref={NoteRef}
                className="truncate overflow-hidden text-sm text-ellipsis"
              >
                Note: {text}
              </span>
              {isNoteOverflowing && (
                <button
                  className="text-secondary cursor-pointer text-xs text-nowrap"
                  onClick={() => setShowFullNote(true)}
                >
                  Show more
                </button>
              )}
            </div>
          )}
        </div>
      )}
    </>
  );
}
