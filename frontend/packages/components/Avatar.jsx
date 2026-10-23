import { useState } from "react";

// Falls back to the initials, or to the fallback when given (null renders nothing), if the image fails to load
export default function Avatar({
  styles = "",
  img,
  initials,
  alt = "",
  fallback,
}) {
  // stores the url instead of a flag, so a new image is tried again
  const [failedImg, setFailedImg] = useState(null);

  if (img && img !== failedImg) {
    return (
      <img
        className={`${styles} size-16 rounded-full object-cover`}
        src={img}
        alt={alt}
        onError={() => setFailedImg(img)}
      />
    );
  }

  if (fallback !== undefined) {
    return fallback;
  }

  return (
    <div
      className={`${styles} from-fg-primary to-primary bg-primary
        text-fg-on-primary flex size-16 items-center justify-center rounded-full
        text-lg font-semibold dark:bg-linear-to-br`}
    >
      {initials?.toUpperCase()}
    </div>
  );
}
