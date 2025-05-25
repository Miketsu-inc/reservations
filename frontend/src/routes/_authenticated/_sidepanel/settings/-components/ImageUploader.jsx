import TrashBinIcon from "@icons/TrashBinIcon";
import UploadIcon from "@icons/UploadIcon";
import { useRef, useState } from "react";

export default function ImageUploader({
  onImageUpload,
  text,
  styles,
  imageStyles,
}) {
  const [preview, setPreview] = useState(null);
  // needed to properly reset  the file input field
  const fileInputRef = useRef(null);

  const validateImage = (file) => {
    const allowedTypes = ["image/jpeg", "image/png", "image/svg+xml"];
    const maxSize = 5 * 1024 * 1024; // 5MB

    if (!allowedTypes.includes(file.type)) {
      alert("Only JPG, PNG, and SVG files are allowed");
      return false;
    }

    if (file.size > maxSize) {
      alert("File size should not exceed 5MB");
      return false;
    }

    return true;
  };

  function handleFileChange(e) {
    const file = e.target.files[0];
    if (file && validateImage(file)) {
      const objectUrl = URL.createObjectURL(file);
      setPreview(objectUrl);
      onImageUpload?.(file);
    }
  }

  function handleDrop(e) {
    e.preventDefault();
    const file = e.dataTransfer.files[0];
    if (file && validateImage(file)) {
      const objectUrl = URL.createObjectURL(file);
      setPreview(objectUrl);
      onImageUpload?.(file);
    }
  }

  function clearImage(e) {
    e.preventDefault();
    setPreview(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  }

  return (
    <label
      onClick={(e) => {
        if (preview) {
          e.preventDefault(); // Prevent clicking on label and triggering the input thus setting the prewiev to null
        }
      }}
      onDrop={handleDrop}
      onDragOver={(e) => {
        e.preventDefault();
      }}
      className={`${styles} group bg-hvr_gray relative flex h-64 w-full cursor-pointer flex-col
        items-center justify-center border-2 border-dashed border-gray-400
        transition-all duration-300 ease-in-out hover:border-gray-500
        dark:border-gray-600 dark:hover:border-gray-400`}
    >
      {preview ? (
        <>
          <img
            src={preview}
            alt="Preview"
            className={`${imageStyles} size-full object-contain`}
          />
          <button
            onClick={clearImage}
            className="absolute top-2 right-2 rounded-full bg-gray-400 p-1 text-white transition-colors
              sm:hidden sm:group-hover:block dark:bg-gray-600"
          >
            <TrashBinIcon styles="size-6" />
          </button>
        </>
      ) : (
        <div className="flex flex-col items-center justify-center text-center">
          <UploadIcon styles="text-gray-500 dark:text-gray-400" />
          <p className="mb-2 text-sm text-gray-500 dark:text-gray-400">
            <span className="font-semibold">Click to upload</span> or drag and
            drop
          </p>
          <p className="text-xs text-gray-500 dark:text-gray-400">
            JPG, PNG, SVG (Max 5MB)
          </p>
          <span className="mt-6 text-sm text-gray-500 dark:text-gray-400">
            {text}
          </span>
        </div>
      )}
      <input
        ref={fileInputRef}
        type="file"
        accept=".jpg,.jpeg,.png,.svg"
        onChange={handleFileChange}
        className="hidden"
      />
    </label>
  );
}
