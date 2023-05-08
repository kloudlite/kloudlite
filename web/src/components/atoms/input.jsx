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
          "bg-critical-50 border-critical-600 text-critical-600 placeholder:text-critical-400":error,
          "text-grey-900 border-grey-300":!error
        }
      )}
    />
    {message && <span className={classNames("bodySm", {
      "text-critical-600":error,
      "text-grey-900":!error
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