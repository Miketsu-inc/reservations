import { useWindowSize } from "@reservations/lib";
import { Drawer, DrawerContent, Modal } from ".";

export default function ResponsiveDialog({
  styles,
  children,
  isOpen,
  onClose,
  disableFocusTrap,
  suspendCloseOnClickOutside,
}) {
  const { isWindowSmall } = useWindowSize();

  if (isWindowSmall) {
    return (
      <Drawer
        open={isOpen}
        onOpenChange={(open) => {
          if (!open) onClose();
        }}
      >
        <DrawerContent styles={styles}>{children}</DrawerContent>
      </Drawer>
    );
  }

  return (
    <Modal
      styles={styles}
      isOpen={isOpen}
      onClose={onClose}
      disableFocusTrap={disableFocusTrap}
      suspendCloseOnClickOutside={suspendCloseOnClickOutside}
    >
      {children}
    </Modal>
  );
}
