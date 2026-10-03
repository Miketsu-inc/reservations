import { Button, ServerError } from "@reservations/components";

function SaveButton({ buttonText, onClick, disabled, isLoading, styles = "" }) {
  return (
    <Button
      type="button"
      variant="primary"
      styles={`px-6 py-2 ${styles}`}
      buttonText={buttonText}
      onClick={onClick}
      disabled={disabled}
      isLoading={isLoading}
    />
  );
}

export default function SettingsSaveAction({
  buttonText,
  onClick,
  disabled,
  isLoading,
  error,
}) {
  return (
    <>
      <div className="hidden items-center gap-3 md:flex">
        {error && <ServerError error={error.message} />}
        <SaveButton
          buttonText={buttonText}
          onClick={onClick}
          disabled={disabled}
          isLoading={isLoading}
        />
      </div>
      <div
        className="bg-layer_bg border-border_color fixed right-0 bottom-0 left-0
          z-30 border-t px-4 pt-4 pb-[calc(1rem+env(safe-area-inset-bottom))]
          md:hidden"
      >
        {error && <ServerError error={error.message} styles="mb-2" />}
        <SaveButton
          buttonText={buttonText}
          onClick={onClick}
          disabled={disabled}
          isLoading={isLoading}
          styles="w-full"
        />
      </div>
    </>
  );
}
