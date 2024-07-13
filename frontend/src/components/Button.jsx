function Button(props) {
    return (
        <button name={props.name} type={props.type}>{props.text}</button>
    )
}

export default Button