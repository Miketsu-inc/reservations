export default function Avatar({ styles, img, initials }) {
  return (
    <>
      {img !== undefined ? (
        <img className="size-16" src={img}></img>
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
