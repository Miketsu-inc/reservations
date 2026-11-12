import {
  Button,
  ResponsiveDialog,
  ResponsiveDialogClose,
  ResponsiveDialogContent,
  Select,
  Textarea,
} from "@reservations/components";
import { useState } from "react";

const REPORT_REASONS = [
  { value: "spam", label: "Spam or fake review" },
  { value: "offensive", label: "Offensive, abusive language" },
  { value: "conflict", label: "Not about my business" },
  { value: "private_info", label: "Contains personal information" },
  { value: "other", label: "Other" },
];

export default function ReportReviewModal({
  review,
  isOpen,
  onClose,
  onSubmit,
}) {
  const [selectedReason, setSelectedReason] = useState("spam");
  const [details, setDetails] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (!review) return null;

  async function handleSubmit() {
    const reasonText = REPORT_REASONS.find(
      (r) => r.value === selectedReason
    )?.label;
    const fullReason = details.trim()
      ? `${reasonText}: ${details.trim()}`
      : reasonText;

    setIsSubmitting(true);
    try {
      await onSubmit(review.id, fullReason);
      onClose();
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <ResponsiveDialog
      open={isOpen}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <ResponsiveDialogContent
        popUpStyles="h-[calc(80vh+3rem)]!"
        styles="w-full lg:min-w-lg max-w-xl p-4 lg:p-6 flex flex-col gap-5"
      >
        <div className="flex h-full w-full flex-col justify-between">
          <div className="flex flex-col gap-6">
            <p className="text-xl font-semibold">Report Review</p>
            <p className="text-fg-muted text-sm">
              Reporting a review will flag it for moderation. While under
              review, it will be marked as under moderation.
            </p>

            <div className="flex flex-col gap-8">
              <Select
                labelText="Reason for reporting"
                options={REPORT_REASONS}
                value={selectedReason}
                onSelect={(opt) => setSelectedReason(opt.value)}
                required={false}
              />

              <Textarea
                labelText="Additional details"
                placeholder="Provide any context that helps moderators understand why this review should be removed"
                styles="p-2 min-h-32"
                value={details}
                inputData={(e) => setDetails(e.value)}
                required={false}
              />
            </div>
          </div>
          <div className="flex flex-row items-center justify-end gap-3 md:mt-8">
            <ResponsiveDialogClose asChild>
              <Button
                variant="secondary"
                styles="py-1.5 px-3"
                buttonText="Cancel"
                type="button"
                onClick={onClose}
              />
            </ResponsiveDialogClose>
            <Button
              variant="danger"
              styles="py-1.5 px-3"
              buttonText="Report Review"
              type="button"
              isLoading={isSubmitting}
              onClick={handleSubmit}
            />
          </div>
        </div>
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}
