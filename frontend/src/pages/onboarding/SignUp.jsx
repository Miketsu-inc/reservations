import PersonalInfo from "./PersonalInfo";
import EmailPage from "./EmailPage";
import PasswordPage from "./PasswordPage";
import { useState } from "react";
import Button from "../../components/Button";
import TickIcon from "../../assets/TickIcon";

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
    <div className="flex justify-center items-center min-w-min min-h-screen bg-custombg sm:bg-white">
      {/*log in container*/}
      <div className="flex flex-col bg-custombg shadow-sm w-full min-h-screen max-w-md sm:shadow-lg sm:rounded-md sm:min-h-1.5 sm:h-4/5 px-10 lg:px-8 sm:pb-16 sm:pt-6">
        <div className="flex items-center justify-center mb-8 mt-6 sm:mt-4">
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
                  ? "absolute text-sm text-customtxt top-10"
                  : "absolute text-sm text-gray-300 top-10"
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
                  ? "absolute text-sm text-customtxt top-10"
                  : "absolute text-sm text-gray-400 top-10"
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
                  ? "absolute text-sm text-customtxt top-10"
                  : "absolute text-sm text-gray-300 top-10"
              }
            >
              Password
            </span>
          </div>
        </div>
        <h2 className="text-customtxt text-2xl py-2 mt-8 sm:mt-4">
          {titles[page]}
        </h2>
        <form
          className="flex flex-col "
          method="POST"
          action=""
          autoComplete="on"
        >
          {PageDisplay()}
          <div className="text-sm flex justify-between items-center mt-2 py-8 sm:mt-8 sm:pt-6 sm:pb-1">
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
