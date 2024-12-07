import { useLocation } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import Button from "../../components/Button";
import FloatingLabelInput from "../../components/FloatingLabelInput";
import ServerError from "../../components/ServerError";
import { MAX_INPUT_LENGTH } from "../../lib/constants";

const defaultEmailData = {
  email: {
    value: "",
    isValid: false,
  },
};

export default function EmailForm({ isCompleted, sendInputData }) {
  const emailRef = useRef();
  const [emailData, setEmailData] = useState(defaultEmailData);
  const [errorMessage, setErrorMessage] = useState("Please enter your email!");
  const [serverError, setServerError] = useState(undefined);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [verifying, setVerifying] = useState(false);
  const [verified, setIsVerified] = useState(false);

  const location = useLocation();

  const token = new URLSearchParams(location.search).get("token");

  useEffect(() => {
    if (token) {
      verifyToken(token);
    }
  }, [token]);

  const verifyToken = async (token) => {
    try {
      const response = await fetch("/api/v1/auth/user/verify-email", {
        method: "POST",
        headers: {
          "Content-type": "application/json; charset=UTF-8",
        },
        body: JSON.stringify({ Token: token }),
      });

      if (!response.ok) {
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        sendInputData({
          email: emailData.email.value,
        });
        setIsVerified(true);
      }
    } catch (err) {
      setServerError(err.message);
    }
  };

  useEffect(() => {
    if (verified) {
      isCompleted(true);
    }
  }, [verified]);

  function emailValidation(email) {
    if (email.length > MAX_INPUT_LENGTH) {
      setErrorMessage(`Inputs must be ${MAX_INPUT_LENGTH} characters or less!`);
      return false;
    }
    if (!email.includes("@")) {
      setErrorMessage("Please enter a valid email!");
      return false;
    }
    return true;
  }

  function handleInputData(data) {
    setEmailData((prevEmailData) => ({
      ...prevEmailData,
      [data.name]: {
        value: data.value,
        isValid: data.isValid,
      },
    }));
  }

  useEffect(() => {
    if (isSubmitting) {
      const sendRequest = async () => {
        try {
          const response = await fetch("/api/v1/auth/user/email", {
            method: "POST",
            headers: {
              "Content-type": "application/json; charset=UTF-8",
            },
            body: JSON.stringify({ Email: emailData.email.value }),
          });

          if (!response.ok) {
            const result = await response.json();
            setServerError(result.error.message);
          } else {
            setVerifying(true);
          }
        } catch (err) {
          setServerError(err.message);
        } finally {
          setIsSubmitting(false);
        }
      };

      sendRequest();
    }
  }, [isSubmitting, emailData]);

  function handleClick() {
    if (!emailData.email.isValid) {
      emailRef.current.triggerValidationError();
    } else {
      setIsSubmitting(true);
    }
  }

  if (verifying) {
    return (
      <div className="flex flex-col items-center justify-center">
        <div className="w-full max-w-md rounded-lg p-6 text-center shadow-md">
          <h1 className="mb-4 text-xl font-semibold text-text_color">
            Check Your Email
          </h1>
          <p className="mb-6 text-gray-200">
            We’ve sent a verification link to your email. Please check your
            inbox and click the link to verify your account.
          </p>
          <p className="mb-4 text-sm text-gray-400">
            Didn’t receive the email? Check your spam folder or
            <a href="#" className="pl-1 text-blue-500 hover:underline">
              resend verification email
            </a>
            .
          </p>
        </div>
      </div>
    );
  }

  return (
    <>
      <ServerError error={serverError} />
      <h2 className="mt-8 py-2 text-center text-xl sm:mt-4">Email</h2>
      <p className="py-2 text-center">
        Enter your email to get started with creating your account
      </p>
      <FloatingLabelInput
        styles=""
        ref={emailRef}
        type="text"
        name="email"
        id="emailInput"
        ariaLabel="Email"
        autoComplete="email"
        labelText="Email"
        errorText={errorMessage}
        inputValidation={emailValidation}
        inputData={handleInputData}
      />
      <p className="px-1 pt-4 text-center text-sm tracking-tight">
        After giving your email you'll get a
        <span className="underline"> verification email</span>. Please check
        your inbox and verify your email.
      </p>
      <div className="flex items-center justify-center">
        <Button
          styles="mt-6 w-full font-semibold focus-visible:outline-1 hover:bg-hvr_primary
            text-white"
          type="button"
          onClick={handleClick}
          buttonText="Verify email"
          isLoading={isSubmitting}
        />
      </div>
    </>
  );
}
