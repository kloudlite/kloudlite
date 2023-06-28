import { AvatarBase } from "../atoms/avatar.jsx";
import PropTypes from "prop-types";
import classNames from "classnames";
import { forwardRef } from "react";
import { BounceIt } from "../bounce-it.jsx";

export const Profile = forwardRef(({ name, subtitle, color, size, ...props }, ref) => {
  return <BounceIt className="w-fit">
    <button
      {...props}
      ref={ref}
      className={classNames("flex py-0.5 px-1 gap-2 items-center ring-offset-1 outline-none transition-all rounded focus-visible:ring-2 focus-visible:ring-border-focus")}>
      <AvatarBase label={name} color={color} size={size} />
      <div className="flex flex-col gap-y-1  items-start">
        <div className={"bodyMd-medium"}>{name}</div>
        {subtitle && <div className={"bodySm text-text-soft"}>{subtitle}</div>}
      </div>
    </button>
  </BounceIt>
})

Profile.propTypes = {
  name: PropTypes.string.isRequired,
  subtitle: PropTypes.string,
  color: PropTypes.oneOf([
    "one",
    "two",
    "three",
    "four",
    "five"
  ]),
  size: PropTypes.oneOf([
    "large", "medium", "small", "extra-small",
  ]),
};

Profile.defaultProps = {
  name: "test",
  subtitle: "subtitle",
  color: "one",
  size: "medium"
};

