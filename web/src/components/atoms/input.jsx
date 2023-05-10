import classNames from "classnames"
import PropTypes from "prop-types";


export const Input = ({label, infoContent, placeholder, value, onChange, error, type, message, Component})=>{
  const C = Component || "input"
  return <div className="flex flex-col gap-1">
    <div className="flex">
      <label className="flex-1 select-none">{label}</label>
      {infoContent && <div className="bodyMd">{infoContent}</div>}
    </div>
    <C
      type={type}
      value={value}
      onChange={onChange} 
      placeholder={placeholder} 
      className={classNames(
        "transition-all outline-none",
        "border ",
        "rounded px-3 py-2 bodyMd ", 
        "ring-offset-1 focus-visible:ring-2 focus:ring-blue-400",
        {
          "bg-surface-danger-subdued border-border-danger text-text-danger placeholder:text-critical-400":error,
          "text-text-default border-border-default":!error
        }
      )}
    />
    {message && <span className={classNames("bodySm", {
      "text-text-danger":error,
      "text-text-default":!error
    })}>{message}</span>}
  </div>
}

Input.propTypes = {
  label: PropTypes.string,
  placeholder: PropTypes.string,
  value: PropTypes.string,
  onChange: PropTypes.func,
  error: PropTypes.bool,
  type: PropTypes.oneOf(["text", "password", "number"]),
  message: PropTypes.string,
  infoContent: PropTypes.elementType,
}

Input.defaultProps = {
  label: "Label",
  placeholder: "Placeholder",
  value: "",
  onChange: ()=>{},
  error: false,
}