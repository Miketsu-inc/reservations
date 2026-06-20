import { Button, Input, Modal } from "@reservations/components";
import { useToast } from "@reservations/lib";
import { useState } from "react";

export default function ChangePasswordModal({ isOpen, onClose }) {
  const [passwords, setPasswords] = useState({
    oldPassword: "",
    newPassword: "",
    confirmNewPassword: "",
  });
  const { showToast } = useToast();

  function updatePasswords(data) {
    setPasswords((prev) => ({ ...prev, ...data }));
  }

  const doNewPasswordsMatch =
    passwords.newPassword != "" &&
    passwords.newPassword === passwords.confirmNewPassword;

  async function submitHandler(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    const response = await fetch("/api/v1/users/password", {
      method: "PUT",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify({
        old_password: passwords.oldPassword,
        new_password: passwords.newPassword,
      }),
    });

    if (!response.ok) {
      const result = await response.json();
      showToast({ message: result.error.message, variant: "error" });
    } else {
      showToast({
        message: "Password updated successfully",
        variant: "success",
      });
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <form onSubmit={submitHandler} className="p-4">
        <p className="mb-6 text-xl">Change password</p>
        <div className="flex flex-col gap-4">
          <Input
            name="OldPassword"
            type="password"
            labelText="Old password"
            inputData={(data) => updatePasswords({ oldPassword: data.value })}
            value={passwords.oldPassword}
          />
          <Input
            name="NewPassword"
            type="password"
            labelText="New password"
            inputData={(data) => updatePasswords({ newPassword: data.value })}
            value={passwords.newPassword}
          />
          <Input
            name="ConfirmNewPassword"
            type="password"
            labelText="Confirm new password"
            inputData={(data) =>
              updatePasswords({ confirmNewPassword: data.value })
            }
            value={passwords.confirmNewPassword}
          />
          <div className="flex flex-row items-center justify-end gap-2 pt-4">
            <Button
              variant="tertiary"
              styles="px-4 py-2"
              buttonText="Close"
              onClick={onClose}
            />
            <Button
              type="submit"
              styles="px-4 py-2"
              buttonText="Update"
              disabled={!doNewPasswordsMatch}
            />
          </div>
        </div>
      </form>
    </Modal>
  );
}
