import TickIcon from "../../assets/TickIcon";

export default function PrograssionBar({ page }) {
  return (
    <div className="mb-8 mt-6 flex items-center justify-center sm:mt-4">
      <div
        className={
          page === 1 || page === 2 || page === 3
            ? "complete"
            : page === 0
              ? "active"
              : "steps"
        }
      >
        {page === 1 || page === 2 || page === 3 ? (
          <TickIcon height={"20"} width={"20"} styles={"fill-white"} />
        ) : (
          "1"
        )}
        <span
          className={
            page === 1 || page === 2 || page === 3
              ? "absolute top-10 text-sm text-customtxt"
              : "absolute top-10 text-sm text-gray-300"
          }
        >
          Name
        </span>
      </div>
      <div
        className={
          page === 1 || page === 2 || page === 3
            ? "connectComplete"
            : "connectSteps"
        }
      ></div>
      <div
        className={
          page === 2 || page == 3 ? "complete" : page === 1 ? "active" : "steps"
        }
      >
        {page === 2 || page === 3 ? (
          <TickIcon height="20" width="20" styles="fill-white" />
        ) : (
          "2"
        )}
        <span
          className={
            page === 2 || page === 3
              ? "absolute top-10 text-sm text-customtxt"
              : "absolute top-10 text-sm text-gray-300"
          }
        >
          Email
        </span>
      </div>
      <div
        className={
          page === 2 || page === 3 ? "connectComplete" : "connectSteps"
        }
      ></div>
      <div
        className={page === 3 ? "complete" : page === 2 ? "active" : "steps"}
      >
        {page === 3 ? (
          <TickIcon height="20" width="20" styles="fill-white" />
        ) : (
          "3"
        )}
        <span
          className={
            page === 3
              ? "absolute top-10 text-sm text-customtxt"
              : "absolute top-10 text-sm text-gray-200"
          }
        >
          Password
        </span>
      </div>
    </div>
  );
}
