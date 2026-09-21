import { Avatar, Select, ServerError } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { calendarTeamMembersQueryOptions } from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";

export default function TeamMemberSelect({
  value,
  onSelect,
  hideWhenSingle = true,
  labelText = "Employee",
  placeholder = "Select an employee",
  disabled,
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

  const options = team.map((member) => ({
    value: member.id,
    label: `${member.first_name} ${member.last_name}`,
    icon: (
      <Avatar
        styles="size-6! rounded-full! text-[10px]!"
        initials={`${member.first_name[0]}${member.last_name[0]}`}
      />
    ),
  }));

  return (
    <Select
      {...props}
      options={options}
      value={value}
      onSelect={onSelect}
      labelText={labelText}
      placeholder={isLoading ? "Loading team members..." : placeholder}
      disabled={isLoading || disabled}
    />
  );
}
