import classNames from "classnames"
import PropTypes from "prop-types"

export const Badge = ({ type, label, IconComp }) => {
  return (
    <div className={classNames("flex gap-1 items-center py-0.5 px-2 w-fit rounded-full bodySm border select-none", {
      "border-border-default bg-surface-default text-text-default": type === "neutral",
      "border-border-primary bg-surface-primary-subdued text-text-primary": type === "info",
      "border-border-success bg-surface-success-subdued text-text-success": type === "success",
      "border-border-warning bg-surface-warning-subdued text-text-warning": type === "warning",
      "border-border-danger bg-surface-danger-subdued text-text-danger": type === "danger",
    })}>{IconComp && <IconComp size={12} color="currentColor" />}{label}</div>
  )
}


Badge.propTypes = {
  type: PropTypes.oneOf(["neutral", "info", "success", "warning", "danger"]),
  label: PropTypes.string.isRequired,
  IconComp: PropTypes.object
}


Badge.defaultProps = {
  type: "neutral",
  label: "badge"
}