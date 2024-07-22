export default function Button(props) {
  return (
    <button
      onClick={props.onClickHandler}
      className={`${props.styles} rounded-lg bg-primary shadow-md py-2 hover:bg-customhvr1 text-customtxt font-medium`}
      name={props.name}
      type={props.type}
    >
      {props.children}
    </button>
  );
}
