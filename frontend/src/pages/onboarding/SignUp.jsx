import { useState } from "react";
import TickIcon from "../../assets/TickIcon";
import Button from "../../components/Button";
import EmailPage from "./EmailPage";
import PasswordPage from "./PasswordPage";
import PersonalInfo from "./PersonalInfo";

export default function SingUp() {
  const [page, setpage] = useState(0);
  const [complete, setComplete] = useState(false);

  const titles = ["Enter your name", "Enter your email", "Enter your password"];

  const PageDisplay = () => {
    if (page === 0) {
      return <PersonalInfo />;
    } else if (page === 1) {
      return <EmailPage />;
    } else if (page === 2) {
      return <PasswordPage />;
    }
  };

  return (
    <div className="flex min-h-screen min-w-min items-center justify-center bg-custombg sm:bg-white">
      {/*log in container*/}
      <div
        className="flex min-h-screen w-full max-w-md flex-col bg-custombg px-10 shadow-sm sm:h-4/5
          sm:min-h-1.5 sm:rounded-md sm:pb-16 sm:pt-6 sm:shadow-lg lg:px-8"
      >
        <div className="mb-8 mt-6 flex items-center justify-center sm:mt-4">
          <div
            className={complete ? "complete" : page === 0 ? "active" : "steps"}
          >
            {complete ? (
              <TickIcon height={"20"} width={"20"} styles={"fill-white"} />
            ) : (
              "1"
            )}
            <span
              className={
                complete
                  ? "absolute top-10 text-sm text-customtxt"
                  : "absolute top-10 text-sm text-gray-300"
              }
            >
              Name
            </span>
          </div>
          <div
            className={page === 1 ? "connectComplete" : "connectSteps"}
          ></div>
          <div
            className={complete ? "complete" : page === 1 ? "active" : "steps"}
          >
            {complete ? (
              <TickIcon height={"20"} width={"20"} styles={"fill-white"} />
            ) : (
              "2"
            )}
            <span
              className={
                complete
                  ? "absolute top-10 text-sm text-customtxt"
                  : "absolute top-10 text-sm text-gray-400"
              }
            >
              Email
            </span>
          </div>
          <div
            className={page === 2 ? "connectComplete" : "connectSteps"}
          ></div>
          <div
            className={complete ? "complete" : page === 2 ? "active" : "steps"}
          >
            {complete ? (
              <TickIcon height={"20"} width={"20"} styles={"fill-white"} />
            ) : (
              "3"
            )}
            <span
              className={
                complete
                  ? "absolute top-10 text-sm text-customtxt"
                  : "absolute top-10 text-sm text-gray-300"
              }
            >
              Password
            </span>
          </div>
        </div>
        <h2 className="mt-8 py-2 text-2xl text-customtxt sm:mt-4">
          {titles[page]}
        </h2>
        <form
          className="flex flex-col"
          method="POST"
          action=""
          autoComplete="on"
        >
          {PageDisplay()}
          <div className="mt-2 flex items-center justify-between py-8 text-sm sm:mt-8 sm:pb-1 sm:pt-6">
            <Button
              styles={""}
              type={"button"}
              disabled={page === 0}
              onClickHandler={() => {
                setpage((currentPage) => currentPage - 1);
              }}
            >
              Prev
            </Button>
            {/*continue button*/}
            <Button
              styles={""}
              type={page === titles.length - 1 ? "submit" : "button"}
              onClickHandler={() => {
                page === setpage((currentPage) => currentPage + 1);
              }}
            >
              {page === titles.length - 1 ? "Submit" : "Countinue"}
            </Button>
          </div>
        </form>
        {/* Login page link */}
      </div>
    </div>
  );
}
