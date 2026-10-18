import {
  Camera01Icon,
  Delete02Icon,
  ImageUploadIcon,
} from "@hugeicons/core-free-icons";
import {
  ACCEPTED_IMAGE_TYPES,
  deleteImage,
  uploadImage,
  useToast,
  validateImageFile,
} from "@reservations/lib";
import { useRef, useState } from "react";
import Avatar from "./Avatar.jsx";
import Icon from "./Icon.jsx";
import ImageCropDialog from "./ImageCropDialog.jsx";
import { Popover, PopoverContent, PopoverTrigger } from "./Popover.jsx";

// Avatar (or logo with shape="square") that can be changed by uploading a new image
export default function AvatarUpload({
  img,
  initials,
  alt,
  styles = "",
  shape = "round",
  uploadUrl,
  deleteUrl,
  canRemove,
  onChange,
}) {
  const inputRef = useRef(null);
  const [file, setFile] = useState(null);
  const [previewUrl, setPreviewUrl] = useState(null);
  const [isSaving, setIsSaving] = useState(false);
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const { showToast } = useToast();

  const isSquare = shape === "square";

  function openFilePicker() {
    setIsMenuOpen(false);
    inputRef.current?.click();
  }

  function selectFile(e) {
    const selected = e.target.files?.[0];
    // allows selecting the same file again
    e.target.value = "";

    if (!selected) {
      return;
    }

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

  async function saveHandler(crop) {
    setIsSaving(true);

    try {
      await uploadImage(uploadUrl, file, crop);
      await onChange?.();
      closeCropDialog();
    } catch (error) {
      showToast({ message: error.message, variant: "error" });
    } finally {
      setIsSaving(false);
    }
  }

  async function removeHandler() {
    setIsMenuOpen(false);

    try {
      await deleteImage(deleteUrl);
      await onChange?.();
    } catch (error) {
      showToast({ message: error.message, variant: "error" });
    }
  }

  const badge = (
    // the hover color is translucent, so it is layered on top of the opaque background
    <span
      className="bg-bg-raised border-border shadow-card absolute -right-1
        -bottom-1 cursor-pointer overflow-hidden rounded-full border"
    >
      <span className="hover:bg-hover flex items-center justify-center p-1.5">
        <Icon icon={Camera01Icon} styles="size-4 text-fg" />
      </span>
    </span>
  );

  return (
    <div className="relative flex w-fit">
      <Avatar
        styles={`${styles} ${isSquare ? "rounded-xl!" : ""}`}
        img={img}
        initials={initials}
        alt={alt}
      />
      {canRemove ? (
        <Popover open={isMenuOpen} onOpenChange={setIsMenuOpen}>
          <PopoverTrigger aria-label="Change picture">{badge}</PopoverTrigger>
          <PopoverContent align="start">
            <div
              className="*:hover:bg-hover space-y-1 text-sm *:flex *:w-full
                *:cursor-pointer *:flex-row *:items-center *:gap-2 *:rounded-lg
                *:p-2"
            >
              <button type="button" onClick={openFilePicker}>
                <Icon icon={ImageUploadIcon} styles="size-4.5 text-fg-muted" />
                Upload new picture
              </button>
              <button
                type="button"
                className="text-fg-danger"
                onClick={removeHandler}
              >
                <Icon icon={Delete02Icon} styles="size-4.5" />
                Remove picture
              </button>
            </div>
          </PopoverContent>
        </Popover>
      ) : (
        <button
          type="button"
          aria-label="Upload picture"
          onClick={openFilePicker}
        >
          {badge}
        </button>
      )}
      <input
        ref={inputRef}
        type="file"
        accept={ACCEPTED_IMAGE_TYPES.join(",")}
        className="hidden"
        onChange={selectFile}
      />
      <ImageCropDialog
        key={previewUrl}
        open={previewUrl !== null}
        onOpenChange={(open) => !open && !isSaving && closeCropDialog()}
        imageSrc={previewUrl}
        cropShape={isSquare ? "rect" : "round"}
        title={isSquare ? "Crop logo" : "Crop picture"}
        isSaving={isSaving}
        onSave={saveHandler}
      />
    </div>
  );
}
