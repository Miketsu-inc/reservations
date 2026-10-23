import {
  Camera01Icon,
  Delete02Icon,
  ImageAdd01Icon,
  ImageUploadIcon,
} from "@hugeicons/core-free-icons";
import { LoaderIcon } from "@reservations/assets";
import { useState } from "react";
import Avatar from "./Avatar.jsx";
import DeleteModal from "./DeleteModal.jsx";
import Icon from "./Icon.jsx";
import { Popover, PopoverContent, PopoverTrigger } from "./Popover.jsx";
import useImageUpload from "./useImageUpload.jsx";

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
  const {
    openFilePicker,
    selectFile,
    remove,
    isSaving,
    fileInput,
    cropDialog,
  } = useImageUpload({
    url,
    cropShape: shape === "square" ? "rect" : "round",
    cropTitle: `Crop ${label}`,
    onChange,
  });

  const rounding = shape === "square" ? "rounded-2xl!" : "rounded-full!";

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
            {badge}
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
          {badge}
        </button>
      )}
      <DeleteModal
        open={isDeleteOpen}
        onOpenChange={setIsDeleteOpen}
        itemName={`the ${label}`}
        onDelete={remove}
      />
      {fileInput}
      {cropDialog}
    </div>
  );
}
