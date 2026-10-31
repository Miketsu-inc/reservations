import Avatar from "./Avatar.jsx";

const sizes = {
  sm: {
    stack: "-space-x-2",
    avatar: "size-6! text-[10px]! border",
    overflow: "size-6 text-[10px] border",
  },
  md: {
    stack: "-space-x-3",
    avatar: "size-8! text-[11px]! border-2",
    overflow: "size-8 text-[11px] border-2",
  },
  lg: {
    stack: "-space-x-4",
    avatar: "size-10! text-xs! border-2",
    overflow: "size-10 text-xs border-2",
  },
};

/**
 * `avatars` are `{ id, img, initials, name }`. `borderColor` should match the
 * surface behind the stack, so the avatars look cut out of each other.
 */
export default function AvatarStack({
  avatars,
  max = 3,
  size = "sm",
  borderColor = "border-bg-raised",
}) {
  if (avatars.length === 0) return null;

  const { stack, avatar, overflow } = sizes[size];
  const remaining = avatars.length - max;

  return (
    <div className={`flex shrink-0 ${stack}`} role="list">
      {avatars.slice(0, max).map(({ id, img, initials, name }) => (
        <span key={id} className="rounded-full" role="listitem" title={name}>
          <Avatar
            img={img}
            initials={initials}
            alt={name}
            styles={`${avatar} ${borderColor}`}
          />
        </span>
      ))}
      {remaining > 0 && (
        <span
          className={`${overflow} ${borderColor} bg-bg-inactive text-fg-muted
          flex items-center justify-center rounded-full font-semibold`}
          role="listitem"
        >
          +{remaining}
        </span>
      )}
    </div>
  );
}
