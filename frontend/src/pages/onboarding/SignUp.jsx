import { useRef, useState } from "react";
import TickIcon from "../../assets/TickIcon";
import Button from "../../components/Button";
import EmailPage from "./EmailPage";
import PasswordPage from "./PasswordPage";
import PersonalInfo from "./PersonalInfo";
import PrograssionBar from "./ProgressionBar";

const defaultSignUpData = {
  firstName: {
    value: "",
    isValid: false,
  },
  lastName: {
    value: "",
    isValid: false,
  },
  email: {
    value: "",
    isValid: false,
  },
  password: {
    value: "",
    isValid: false,
  },
  confirmPassword: {
    value: "",
    isValid: false,
  },
};

const titles = ["Enter your name", "Enter your email", "Enter your password"];

export default function SingUp() {
  const firstNameRef = useRef();
  const lastNameRef = useRef();
  const emailRef = useRef();
  const passwordRef = useRef();
  const confirmPasswordRef = useRef();
  const [page, setpage] = useState(0);
  const [signUpData, setSignUpData] = useState(defaultSignUpData);
  const [submitted, setSubmitted] = useState(false);

  function handleInputData(data) {
    setSignUpData((prevSignUpData) => ({
      ...prevSignUpData,
      [data.name]: {
        value: data.value,
        isValid: data.isValid,
      },
    }));
  }

  function handleClick() {
    let hasError = false;
    if (page === 0 && !signUpData.firstName.isValid) {
      firstNameRef.current.triggerValidationError();
      hasError = true;
    }
    if (page === 0 && !signUpData.lastName.isValid) {
      lastNameRef.current.triggerValidationError();
      hasError = true;
    }
    if (page === 1 && !signUpData.email.isValid) {
      emailRef.current.triggerValidationError();
      hasError = true;
    }

    if (!hasError) {
      setpage((currentPage) => currentPage + 1);
    }
  }

  function handleSubmit(e) {
    e.preventDefault();
    let hasError = false;
    if (!signUpData.password.isValid) {
      passwordRef.current.triggerValidationError();
      hasError = true;
    }
    if (!signUpData.confirmPassword.isValid) {
      confirmPasswordRef.current.triggerValidationError();
      hasError = true;
    }
    if (
      signUpData.password.value !== signUpData.confirmPassword.value &&
      signUpData.password.isValid
    ) {
      confirmPasswordRef.current.triggerValidationError();
      hasError = true;
    }
    if (!hasError) {
      setSubmitted(true);
      console.log(signUpData);
    }
  }

  function changeInputField() {
    if (page === 0 && !submitted) {
      return (
        <PersonalInfo
          handleInputData={handleInputData}
          firstNameRef={firstNameRef}
          lastNameRef={lastNameRef}
        />
      );
    } else if (page === 1 && !submitted) {
      return (
        <EmailPage handleInputData={handleInputData} emailRef={emailRef} />
      );
    } else if (page === 2 && !submitted) {
      return (
        <PasswordPage
          handleInputData={handleInputData}
          passwordRef={passwordRef}
          confirmPasswordRef={confirmPasswordRef}
        />
      );
    } else {
      return (
        <div className="flex flex-col items-center justify-center">
          <div className="my-4 rounded-full border-4 border-green-600 p-6">
            <TickIcon height="60" width="60" styles="fill-green-600" />
          </div>
          <div className="mt-10 text-center text-xl font-semibold text-customtxt">
            You signed up successfully
          </div>
        </div>
      );
    }
  }

  return (
    <div className="flex min-h-screen min-w-min items-center justify-center bg-custombg">
      {/*log in container*/}
      <div
        className="flex min-h-screen w-full max-w-md flex-col bg-custombg px-10 shadow-sm sm:h-4/5
          sm:min-h-1.5 sm:rounded-md sm:bg-slate-400 sm:bg-opacity-5 sm:pb-16 sm:pt-6
          sm:shadow-lg lg:px-8"
      >
        <PrograssionBar page={page} submitted={submitted} />

        <h2 className="mt-8 py-2 text-2xl text-customtxt sm:mt-4">
          {!submitted ? titles[page] : ""}
        </h2>
        <form
          className="flex flex-col"
          method="POST"
          action=""
          autoComplete="on"
          onSubmit={handleSubmit}
        >
          {changeInputField()}
          <div className="flex items-center justify-center">
            {page < titles.length - 1 ? (
              <Button
                styles="mt-10 w-3/4 font-semibold"
                type="button"
                key="continueButton"
                onClick={handleClick}
              >
                Continue
              </Button>
            ) : !submitted ? (
              <Button
                styles="mt-10 w-3/4 font-semibold"
                type="submit"
                key="submitButton"
              >
                Finish
              </Button>
            ) : (
              ""
            )}
          </div>
        </form>
        {/* Login page link */}
      </div>
    </div>
  );
}
