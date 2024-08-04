import TickIcon from "../../assets/TickIcon";

export default function PrograssionBar({ page, submitted }) {
  return (
    <div className="mb-8 mt-6 flex items-center justify-center sm:mt-4">
      <div
        className={page !== 0 ? "complete" : page === 0 ? "active" : "steps"}
      >
        {page !== 0 ? (
          <TickIcon height={"20"} width={"20"} styles={"fill-white"} />
        ) : (
          "1"
        )}
        <span
          className={
            page !== 0
              ? "absolute top-10 text-sm text-gray-500"
              : "absolute top-10 text-sm"
          }
        >
          Name
        </span>
      </div>
      <div className={page !== 0 ? "connectComplete" : "connectSteps"}></div>
      <div
        className={page === 2 ? "complete" : page === 1 ? "active" : "steps"}
      >
        {page === 2 ? (
          <TickIcon height="20" width="20" styles="fill-white" />
        ) : (
          "2"
        )}
        <span
          className={
            page === 2 || page === 3 || page === 0
              ? "absolute top-10 text-sm text-gray-500"
              : "absolute top-10 text-sm"
          }
        >
          Email
        </span>
      </div>
      <div className={page === 2 ? "connectComplete" : "connectSteps"}></div>
      <div className={submitted ? "complete" : page === 2 ? "active" : "steps"}>
        {submitted ? (
          <TickIcon height="20" width="20" styles="fill-white" />
        ) : (
          "3"
        )}
        <span
          className={
            page === 0 || page === 1 || submitted
              ? "absolute top-10 text-sm text-gray-500"
              : "absolute top-10 text-sm"
          }
        >
          Password
        </span>
      </div>
    </div>
  );
}
