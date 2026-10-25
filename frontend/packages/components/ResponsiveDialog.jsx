import { useWindowSize } from "@reservations/lib";
import { createContext, useContext } from "react";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerTrigger,
} from "./Drawer.jsx";
import { Modal, ModalClose, ModalContent, ModalTrigger } from "./Modal.jsx";

// Only the root reads the window size, the parts follow it. Otherwise a resize can re-render
// a part before the root and render a drawer part inside a modal root, which crashes
const IsDrawerContext = createContext(false);

export function ResponsiveDialog({ actionsRef, ...props }) {
  const { isWindowSmall } = useWindowSize();

  return (
    <IsDrawerContext.Provider value={isWindowSmall}>
      {isWindowSmall ? (
        <Drawer actionsRef={actionsRef} {...props} />
      ) : (
        <Modal actionsRef={actionsRef} {...props} />
      )}
    </IsDrawerContext.Provider>
  );
}

export function ResponsiveDialogTrigger({ asChild, ...props }) {
  if (useContext(IsDrawerContext)) {
    return <DrawerTrigger asChild={asChild} {...props} />;
  }

  return <ModalTrigger asChild={asChild} {...props} />;
}

export function ResponsiveDialogContent({ styles, ...props }) {
  if (useContext(IsDrawerContext)) {
    return <DrawerContent styles={styles} {...props} />;
  }

  return <ModalContent styles={styles} {...props} />;
}

export function ResponsiveDialogClose({ asChild, ...props }) {
  if (useContext(IsDrawerContext)) {
    return <DrawerClose asChild={asChild} {...props} />;
  }

  return <ModalClose asChild={asChild} {...props} />;
}
