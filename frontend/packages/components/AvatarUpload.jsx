import {
  Camera01Icon,
  Delete02Icon,
  ImageAdd01Icon,
  ImageUploadIcon,
} from "@hugeicons/core-free-icons";
import { LoaderIcon } from "@reservations/assets";
import {
  ACCEPTED_IMAGE_TYPES,
  invalidateLocalStorageAuth,
  uploadImage,
  useToast,
  validateImageFile,
} from "@reservations/lib";
import { useRef, useState } from "react";
import Avatar from "./Avatar.jsx";
import DeleteModal from "./DeleteModal.jsx";
import Icon from "./Icon.jsx";
import ImageCropDialog from "./ImageCropDialog.jsx";
import { Popover, PopoverContent, PopoverTrigger } from "./Popover.jsx";

// Used for every single square image (user and team avatars, merchant logo), the shape only changes the look
export default function AvatarUpload({
  img,
  initials,
  alt,
  styles = "",
  shape = "round",
  label = "picture",
  url,
  canRemove = !!img,
  onChange,
}) {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [isDragging, setIsDragging] = useState(false);
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
      const response = await fetch(url, {
        method: "DELETE",
        headers: {
          Accept: "application/json",
        },
      });

      if (!response.ok) {
        const result = await response.json();
        invalidateLocalStorageAuth(response.status);
        showToast({ message: result.error.message, variant: "error" });
      } else {
        await onChange?.();
      }
    } catch (error) {
      showToast({ message: error.message, variant: "error" });
    }
  }

  const rounding = shape === "square" ? "rounded-2xl!" : "rounded-full!";

  return (
    <div
      className="relative flex w-fit shrink-0"
      onDragOver={(e) => {
        e.preventDefault();
        setIsDragging(true);
      }}
      onDragLeave={() => setIsDragging(false)}
      onDrop={(e) => {
        e.preventDefault();
        setIsDragging(false);
        const file = e.dataTransfer.files[0];
        if (file && !isSaving) selectFile(file);
      }}
    >
      <Avatar
        styles={`${styles} ${rounding}`}
        img={img}
        initials={initials}
        alt={alt}
        fallback={
          initials ? undefined : (
            <div
              className={`${styles} ${rounding} border-border-strong
                bg-bg-sunken text-fg-muted flex size-16 items-center
                justify-center border border-dashed`}
            >
              <Icon icon={ImageAdd01Icon} styles="size-6" />
            </div>
          )
        }
      />
      {(isSaving || isDragging) && (
        <span
          className={`${rounding} bg-bg-backdrop ring-primary absolute inset-0
          flex items-center justify-center ${isDragging ? "ring-2" : ""}`}
        >
          {isSaving && <LoaderIcon styles="size-6 text-fg-on-primary" />}
        </span>
      )}
      {canRemove ? (
        <Popover open={isMenuOpen} onOpenChange={setIsMenuOpen}>
          <PopoverTrigger aria-label={`Change ${label}`}>
            <CameraBadge />
          </PopoverTrigger>
          <PopoverContent align="start">
            <div
              className="*:hover:bg-hover space-y-1 text-sm *:flex *:w-full
                *:cursor-pointer *:flex-row *:items-center *:gap-2 *:rounded-lg
                *:p-2"
            >
              <button
                type="button"
                onClick={() => {
                  setIsMenuOpen(false);
                  openFilePicker();
                }}
              >
                <Icon icon={ImageUploadIcon} styles="size-4.5 text-fg-muted" />
                Upload new {label}
              </button>
              <button
                type="button"
                className="text-fg-danger"
                onClick={() => {
                  setIsMenuOpen(false);
                  setIsDeleteOpen(true);
                }}
              >
                <Icon icon={Delete02Icon} styles="size-4.5" />
                Remove {label}
              </button>
            </div>
          </PopoverContent>
        </Popover>
      ) : (
        <button
          type="button"
          aria-label={`Upload ${label}`}
          onClick={openFilePicker}
        >
          <CameraBadge />
        </button>
      )}
      <DeleteModal
        open={isDeleteOpen}
        onOpenChange={setIsDeleteOpen}
        itemName={`the ${label}`}
        onDelete={remove}
      />
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
      <ImageCropDialog
        key={previewUrl}
        open={previewUrl !== null}
        onOpenChange={(open) => !open && !isSaving && closeCropDialog()}
        imageSrc={previewUrl}
        cropShape={shape === "square" ? "rect" : "round"}
        title={`Crop ${label}`}
        isSaving={isSaving}
        onSave={save}
      />
    </div>
  );
}

function CameraBadge() {
  return (
    <span
      className="bg-bg-raised border-border shadow-card absolute -right-1
        -bottom-1 cursor-pointer overflow-hidden rounded-full border"
    >
      <span className="hover:bg-hover flex items-center justify-center p-1.5">
        <Icon icon={Camera01Icon} styles="size-4 text-fg" />
      </span>
    </span>
  );
}
