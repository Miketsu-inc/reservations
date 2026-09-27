import { Delete02Icon, FileImageIcon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";
import { useEffect, useRef, useState } from "react";

export default function ImageUploader({
  onImageUpload,
  text,
  styles,
  imageStyles,
}) {
  const [preview, setPreview] = useState(null);
  const fileInputRef = useRef(null);

  useEffect(() => {
    return () => {
      if (preview) URL.revokeObjectURL(preview);
    };
  }, [preview]);

  function handleFile(file) {
    const allowedTypes = ["image/jpeg", "image/png", "image/svg+xml"];
    const maxSize = 5 * 1024 * 1024;

    if (!allowedTypes.includes(file.type)) {
      alert("Only JPG, PNG, and SVG files are allowed");
      return;
    }

    if (file.size > maxSize) {
      alert("File size should not exceed 5MB");
      return;
    }

    setPreview(URL.createObjectURL(file));
    onImageUpload?.(file);
  }

  function clearImage(event) {
    event.preventDefault();
    setPreview(null);
    if (fileInputRef.current) fileInputRef.current.value = "";
  }

  return (
    <label
      onClick={(event) => preview && event.preventDefault()}
      onDrop={(event) => {
        event.preventDefault();
        const file = event.dataTransfer.files[0];
        if (file) handleFile(file);
      }}
      onDragOver={(event) => event.preventDefault()}
      className={`${styles} group bg-hvr_gray relative flex h-64 w-full
        cursor-pointer flex-col items-center justify-center border-2
        border-dashed border-gray-400 transition-colors hover:border-gray-500
        dark:border-gray-600 dark:hover:border-gray-400`}
    >
      {preview ? (
        <>
          <img
            src={preview}
            alt={`${text} preview`}
            className={`${imageStyles} size-full object-contain`}
          />
          <button
            type="button"
            aria-label={`Remove ${text.toLowerCase()}`}
            onClick={clearImage}
            className="absolute top-2 right-2 rounded-full bg-gray-600 p-1
              text-white"
          >
            <Icon icon={Delete02Icon} styles="size-5" />
          </button>
        </>
      ) : (
        <div className="flex flex-col items-center justify-center px-4 text-center">
          <Icon icon={FileImageIcon} styles="text-gray-500 dark:text-gray-400" />
          <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
            <span className="font-semibold">Click to upload</span> or drag and
            drop
          </p>
          <p className="text-xs text-gray-500 dark:text-gray-400">
            JPG, PNG, SVG (max. 5 MB)
          </p>
          <span className="mt-5 text-sm text-gray-500 dark:text-gray-400">
            {text}
          </span>
        </div>
      )}
      <input
        ref={fileInputRef}
        type="file"
        accept=".jpg,.jpeg,.png,.svg"
        onChange={(event) => {
          const file = event.target.files[0];
          if (file) handleFile(file);
        }}
        className="hidden"
      />
    </label>
  );
}
