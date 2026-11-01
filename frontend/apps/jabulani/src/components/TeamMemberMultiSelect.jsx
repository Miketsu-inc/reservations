import { MultiSelect, ServerError } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { calendarTeamMembersQueryOptions, toAvatar } from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";

export default function TeamMemberMultiSelect({
  values = [],
  onSelect,
  hideWhenSingle = false,
  labelText = "Team members",
  displayText = "member",
  placeholder = "Select team members",
  disabled,
  emptyText,
  ...props
}) {
  const { merchantId } = useAuth();
  const {
    data: team = [],
    isLoading,
    isError,
    error,
  } = useQuery(calendarTeamMembersQueryOptions(merchantId));

  if (isError) {
    return <ServerError error={error.message || error} />;
  }

  if (!isLoading && hideWhenSingle && team.length <= 1) {
    return null;
  }

  const options = team.map((member) => {
    const { img, initials, name } = toAvatar(member);
    return { value: member.id, label: name, img, initials };
  });

  return (
    <MultiSelect
      {...props}
      options={options}
      values={values}
      onSelect={onSelect}
      labelText={labelText}
      displayText={displayText}
      placeholder={isLoading ? "Loading team members..." : placeholder}
      emptyText={isLoading ? "Loading team members..." : emptyText}
      disabled={isLoading || disabled}
    />
  );
}
