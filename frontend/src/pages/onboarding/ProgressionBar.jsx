import TickIcon from "../../assets/TickIcon";

export default function PrograssionBar(props) {
  return (
    <div className="mb-8 mt-6 flex items-center justify-center sm:mt-4">
      <div
        className={
          props.page !== 0 ? "complete" : props.page === 0 ? "active" : "steps"
        }
      >
        {props.page !== 0 ? (
          <TickIcon height={"20"} width={"20"} styles={"fill-white"} />
        ) : (
          "1"
        )}
        <span
          className={
            props.page !== 0
              ? "absolute top-10 text-sm text-gray-500"
              : "absolute top-10 text-sm text-customtxt"
          }
        >
          Name
        </span>
      </div>
      <div
        className={props.page !== 0 ? "connectComplete" : "connectSteps"}
      ></div>
      <div
        className={
          props.page === 2 ? "complete" : props.page === 1 ? "active" : "steps"
        }
      >
        {props.page === 2 ? (
          <TickIcon height="20" width="20" styles="fill-white" />
        ) : (
          "2"
        )}
        <span
          className={
            props.page === 2 || props.page === 3 || props.page === 0
              ? "absolute top-10 text-sm text-gray-500"
              : "absolute top-10 text-sm text-customtxt"
          }
        >
          Email
        </span>
      </div>
      <div
        className={props.page === 2 ? "connectComplete" : "connectSteps"}
      ></div>
      <div
        className={
          props.submitted ? "complete" : props.page === 2 ? "active" : "steps"
        }
      >
        {props.submitted ? (
          <TickIcon height="20" width="20" styles="fill-white" />
        ) : (
          "3"
        )}
        <span
          className={
            props.page === 0 || props.page === 1 || props.submitted
              ? "absolute top-10 text-sm text-gray-500"
              : "absolute top-10 text-sm text-customtxt"
          }
        >
          Password
        </span>
      </div>
    </div>
  );
}
