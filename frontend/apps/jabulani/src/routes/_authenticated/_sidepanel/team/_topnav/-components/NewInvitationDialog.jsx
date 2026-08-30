import {
  Button,
  Input,
  ResponsiveDialog,
  Select,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { useToast } from "@reservations/lib";
import { useState } from "react";

export default function NewInvitationDialog({ Route, isOpen, onClose }) {
  const { showToast } = useToast();
  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { merchantId } = useAuth();
  const [invitation, setInvitation] = useState({ email: "", role: "staff" });
  const [isSelectOpen, setIsSelectOpen] = useState(false);

  function updateInvitation(data) {
    setInvitation((prev) => ({ ...prev, ...data }));
  }

  async function inviteMember(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    const response = await fetch(
      `/api/v1/merchants/${merchantId}/team/invitations`,
      {
        method: "POST",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          email: invitation.email,
          role: invitation.role,
        }),
      }
    );

    if (!response.ok) {
      const result = await response.json();
      showToast({ message: result.error.message, variant: "error" });
    } else {
      showToast({
        message: "Team member invited successfully",
        variant: "success",
      });

      await queryClient.invalidateQueries({
        queryKey: [merchantId, "invitations"],
      });
      setInvitation({ email: "", role: "staff" });
      onClose();
    }
  }

  return (
    <ResponsiveDialog
      isOpen={isOpen}
      onClose={onClose}
      disableFocusTrap={true}
      suspendCloseOnClickOutside={isSelectOpen}
    >
      <form
        onSubmit={inviteMember}
        className="flex h-full w-full flex-col gap-6 md:w-lg md:p-4"
      >
        <div>
          <p className="text-lg font-semibold">New Invitation</p>
          <p className="text-text_color/80">
            Invite a person to your team. They will get an email asking them to
            join.
          </p>
        </div>
        <div className="flex flex-col gap-4">
          <Input
            id="Email"
            name="Email"
            type="email"
            labelText="Email"
            placeholder="example@gmail.com"
            required={true}
            value={invitation.email}
            inputData={(data) => updateInvitation({ email: data.value })}
          />
          <Select
            required={true}
            labelText="Role"
            styles="w-full"
            value={invitation.role}
            options={[
              { value: "staff", label: "Staff" },
              { value: "admin", label: "Admin" },
            ]}
            onSelect={(option) => updateInvitation({ role: option.value })}
            onOpenChange={setIsSelectOpen}
            placeholder="Pick a role"
          />
        </div>
        <div className="flex items-center justify-end gap-2">
          <Button
            styles="py-2 px-4 hidden lg:block"
            buttonText="Cancel"
            variant="tertiary"
            type="button"
            onClick={onClose}
          />
          <Button
            styles="py-2 px-4 w-full lg:w-auto"
            buttonText="Invite"
            variant="primary"
            type="submit"
          />
        </div>
      </form>
    </ResponsiveDialog>
  );
}
