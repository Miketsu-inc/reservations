export default function Avatar({ styles = "", img, initials, alt = "" }) {
  return (
    <>
      {img ? (
        <img
          className={`${styles} size-16 rounded-full object-cover`}
          src={img}
          alt={alt}
        />
      ) : (
        <div
          className={`${styles} from-fg-primary to-primary bg-primary
            text-fg-on-primary flex size-16 items-center justify-center
            rounded-full text-lg font-semibold dark:bg-linear-to-br`}
        >
          {initials?.toUpperCase()}
        </div>
      )}
    </>
  );
}
