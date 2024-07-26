import { useState } from "react";
import Button from "../../components/Button";
import EmailPage from "./EmailPage";
import PasswordPage from "./PasswordPage";
import PersonalInfo from "./PersonalInfo";
import PrograssionBar from "./ProgressionBar";

export default function SingUp() {
  const [page, setpage] = useState(0);
  const [errors, setErrors] = useState({});
  const [isValid, setIsValid] = useState({
    firstname: true,
    lastName: true,
    email: true,
    password: true,
    confirmPassword: true,
  });

  const [formValues, setFormValues] = useState({
    firstName: "",
    lastName: "",
    email: "",
    password: "",
    confirmPassword: "",
  });

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormValues((...prevData) => ({ ...prevData, [name]: value }));
    setErrors((...prevErrors) => ({ ...prevErrors, [name]: "" }));
    setIsValid((...prevIsValid) => ({ ...prevIsValid, [name]: true }));
  };

  const handleBlur = (e) => {
    const { name, value } = e.target;
    const Errors = { ...errors };
    const newIsValid = { ...isValid };
    if (name === "firstName") {
    } else if (name === "lastName") {
    } else if (
      name === "email" &&
      formValues.email.trim() !== "" &&
      formValues.email.includes("@") !== true
    ) {
      Errors.email = "Please enter a valid email";
      newIsValid.email = false;
    } else if (name === "password" && formValues.password.trim() !== "") {
    } else if (
      name === "confirmPassword" &&
      formValues.confirmPassword !== formValues.password &&
      formValues.password.trim() !== true &&
      formValues.confirmPassword.trim() !== ""
    ) {
      Errors.confirmPassword = "This password should match the previous one";
      newIsValid.confirmPassword = false;
    } else {
      delete Errors[name];
      newIsValid[name] = true;
    }
    setIsValid(newIsValid);
    setErrors(Errors);
  };

  const titles = ["Enter your name", "Enter your email", "Enter your password"];

  const PageDisplay = () => {
    if (page === 0) {
      return (
        <PersonalInfo
          formValues={formValues}
          handleChange={handleChange}
          handleBlur={handleBlur}
          errors={errors}
        />
      );
    } else if (page === 1) {
      return (
        <EmailPage
          formValues={formValues}
          handleChange={handleChange}
          handleBlur={handleBlur}
          errors={errors}
        />
      );
    } else if (page === 2) {
      return (
        <PasswordPage
          formValues={formValues}
          handleChange={handleChange}
          handleBlur={handleBlur}
          errors={errors}
        />
      );
    }
  };

  return (
    <div className="flex min-h-screen min-w-min items-center justify-center bg-custombg">
      {/*log in container*/}
      <div
        className="flex min-h-screen w-full max-w-md flex-col bg-custombg px-10 shadow-sm sm:h-4/5
          sm:min-h-1.5 sm:rounded-md sm:bg-slate-400 sm:bg-opacity-5 sm:pb-16 sm:pt-6
          sm:shadow-lg lg:px-8"
      >
        <PrograssionBar page={page} />

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
              styles="px-2"
              type="button"
              disabled={page === 0}
              onClick={() => {
                setpage((currentPage) => currentPage - 1);
              }}
            >
              Prev
            </Button>
            {/*continue button*/}
            <Button
              styles="px-2"
              type={page === titles.length - 1 ? "submit" : "button"}
              disabled={page === 2}
              onClick={() => {
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
