export default function Avatar({ styles = "", img, initials, alt = "" }) {
  return (
    <>
      {img ? (
        <img
          className={`${styles} size-16 rounded-md object-cover`}
          src={img}
          alt={alt}
        />
      ) : (
        <div
          className={`${styles} from-secondary to-primary bg-primary flex
            size-16 items-center justify-center rounded-md text-lg text-white
            dark:bg-linear-to-br`}
        >
          {initials?.toUpperCase()}
        </div>
      )}
    </>
  );
}
