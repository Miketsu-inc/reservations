import { Avatar, Select, ServerError } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { calendarTeamMembersQueryOptions, toAvatar } from "@reservations/lib";
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

  const options = team.map((member) => {
    const { img, initials, name } = toAvatar(member);
    return {
      value: member.id,
      label: name,
      icon: (
        <Avatar styles="size-6! text-[10px]!" img={img} initials={initials} />
      ),
    };
  });

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
