import { useState } from "react";

export default function MerchantLogo({
  styles = "",
  src,
  alt,
  fallback = null,
}) {
  // stores the url instead of a flag, so a new logo is tried again
  const [failedSrc, setFailedSrc] = useState(null);

  if (!src || src === failedSrc) {
    return fallback;
  }

  return (
    <img
      className={`${styles} rounded-xl object-cover`}
      src={src}
      alt={alt}
      onError={() => setFailedSrc(src)}
    />
  );
}
