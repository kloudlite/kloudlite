import { Avatar, AvatarBase } from "../atoms/avatar.jsx";
import PropTypes from "prop-types";
import { AriaButton } from "../atoms/button.jsx";
import classNames from "classnames";
import { useFocusRing } from "react-aria";

export const Profile = ({ name, subtitle, color, size }) => {
  let { isFocusVisible, focusProps } = useFocusRing()
  return <AriaButton
    {...focusProps}
    className={classNames("outline-none flex py-0.5 px-1 gap-2 items-center ring-offset-1 outline-none transition-all rounded",
      {
        "focus:ring-2 focus:ring-border-focus": isFocusVisible
      })}>
    <AvatarBase label={name} color={color} size={size} renderAs="div" />
    <div className="flex flex-col gap-y-1  items-start">
      <div className={"bodyMd-medium"}>{name}</div>
      {subtitle && <div className={"bodySm text-text-soft"}>{subtitle}</div>}
    </div>
  </AriaButton>
}

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

