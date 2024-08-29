import { useEffect, useRef, useState } from "react";
import HamburgerMenuIcon from "../../assets/HamburgerMenuIcon";
import { useClickOutside, useWindowSize } from "../../lib/hooks";
import SidePanelProfile from "./SidePanelProfile";

export default function SidePanel({ children, profileImage, profileText }) {
  const windowSize = useWindowSize();
  const [isOpen, setIsOpen] = useState(windowSize !== "sm" ? true : false);
  const sidePanelRef = useRef();
  useClickOutside(sidePanelRef, closeSidePanelHandler);

  useEffect(() => {
    if (windowSize === "sm") {
      setIsOpen(false);
    } else {
      setIsOpen(true);
    }
  }, [windowSize, setIsOpen]);

  function sidePanelClickHandler() {
    setIsOpen(true);
  }

  function closeSidePanelHandler() {
    if (windowSize === "sm") {
      setIsOpen(false);
    }
  }

  return (
    <>
      <button
        aria-controls="sidepanel"
        type="button"
        className="hover:bg-hvr_gray text-text_color ms-3 mt-2 inline-flex items-center rounded-lg
          p-2 text-sm focus:outline-none focus:ring-2 focus:ring-gray-200
          dark:focus:ring-gray-600 sm:hidden"
        onClick={sidePanelClickHandler}
      >
        <span className="sr-only">Open sidepanel</span>
        <HamburgerMenuIcon styles={"h-6 w-6"} />
      </button>
      <aside
        ref={sidePanelRef}
        id="sidepanel"
        className={`${isOpen ? "sm:translate-x-0" : "-translate-x-full"} fixed left-0 top-0 z-40
          h-screen w-64 transition-transform`}
        aria-label="Sidepanel"
      >
        <div className="bg-layer_bg flex h-full flex-col overflow-y-auto px-3 py-4">
          <SidePanelProfile
            image={profileImage}
            text={profileText}
            closeSidePanel={closeSidePanelHandler}
            windowSize={windowSize}
          />
          <hr className="my-4"></hr>
          <div className="flex flex-1 flex-col space-y-2 font-medium">
            {children}
          </div>
        </div>
      </aside>
    </>
  );
}
