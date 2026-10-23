import {
  ACCEPTED_IMAGE_TYPES,
  deleteImage,
  uploadImage,
  useToast,
  validateImageFile,
} from "@reservations/lib";
import { useRef, useState } from "react";
import ImageCropDialog from "./ImageCropDialog.jsx";

// Upload flow of a single square image: pick a file, crop it, upload it. The returned
// fileInput and cropDialog elements have to be rendered by the caller
export default function useImageUpload({
  url,
  cropShape = "round",
  cropTitle,
  onChange,
}) {
  const inputRef = useRef(null);
  const [file, setFile] = useState(null);
  const [previewUrl, setPreviewUrl] = useState(null);
  const [isSaving, setIsSaving] = useState(false);
  const { showToast } = useToast();

  function openFilePicker() {
    inputRef.current?.click();
  }

  function selectFile(selected) {
    const error = validateImageFile(selected);
    if (error) {
      showToast({ message: error, variant: "error" });
      return;
    }

    setFile(selected);
    setPreviewUrl(URL.createObjectURL(selected));
  }

  function closeCropDialog() {
    if (previewUrl) {
      URL.revokeObjectURL(previewUrl);
    }

    setFile(null);
    setPreviewUrl(null);
  }

  async function save(crop) {
    setIsSaving(true);

    try {
      await uploadImage(url, file, crop);
      await onChange?.();
      closeCropDialog();
    } catch (error) {
      showToast({ message: error.message, variant: "error" });
    } finally {
      setIsSaving(false);
    }
  }

  async function remove() {
    try {
      await deleteImage(url);
      await onChange?.();
    } catch (error) {
      showToast({ message: error.message, variant: "error" });
    }
  }

  const fileInput = (
    <input
      ref={inputRef}
      type="file"
      accept={ACCEPTED_IMAGE_TYPES.join(",")}
      className="hidden"
      onChange={(e) => {
        const selected = e.target.files?.[0];
        // allows selecting the same file again
        e.target.value = "";
        if (selected) selectFile(selected);
      }}
    />
  );

  const cropDialog = (
    <ImageCropDialog
      key={previewUrl}
      open={previewUrl !== null}
      onOpenChange={(open) => !open && !isSaving && closeCropDialog()}
      imageSrc={previewUrl}
      cropShape={cropShape}
      title={cropTitle}
      isSaving={isSaving}
      onSave={save}
    />
  );

  return {
    openFilePicker,
    selectFile,
    remove,
    isSaving,
    fileInput,
    cropDialog,
  };
}
