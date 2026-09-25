import { useWindowSize } from "@reservations/lib";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerTrigger,
} from "./Drawer.jsx";
import { Modal, ModalClose, ModalContent, ModalTrigger } from "./Modal.jsx";

export function ResponsiveDialog({ actionsRef, ...props }) {
  const { isWindowSmall } = useWindowSize();

  if (isWindowSmall) {
    return <Drawer actionsRef={actionsRef} {...props} />;
  }

  return <Modal actionsRef={actionsRef} {...props} />;
}

export function ResponsiveDialogTrigger({ asChild, ...props }) {
  const { isWindowSmall } = useWindowSize();

  if (isWindowSmall) {
    return <DrawerTrigger asChild={asChild} {...props} />;
  }

  return <ModalTrigger asChild={asChild} {...props} />;
}

export function ResponsiveDialogContent({ styles, ...props }) {
  const { isWindowSmall } = useWindowSize();

  if (isWindowSmall) {
    return <DrawerContent styles={styles} {...props} />;
  }

  return <ModalContent styles={styles} {...props} />;
}

export function ResponsiveDialogClose({ asChild, ...props }) {
  const { isWindowSmall } = useWindowSize();

  if (isWindowSmall) {
    return <DrawerClose asChild={asChild} {...props} />;
  }

  return <ModalClose asChild={asChild} {...props} />;
}
