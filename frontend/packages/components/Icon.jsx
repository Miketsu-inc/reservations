import { HugeiconsIcon } from "@hugeicons/react";

export default function Icon({ styles, icon, altIcon, showAlt, ...props }) {
  return (
    <HugeiconsIcon
      className={`${styles}`}
      icon={icon}
      altIcon={altIcon}
      showAlt={showAlt}
      {...props}
    />
  );
}
