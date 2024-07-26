export default function Button(props) {
  return (
    <button
      onClick={props.onClick}
      className={`${props.styles} rounded-lg bg-primary py-2 font-medium text-customtxt shadow-md
        hover:bg-customhvr1`}
      name={props.name}
      type={props.type}
    >
      {props.children}
    </button>
  );
}
