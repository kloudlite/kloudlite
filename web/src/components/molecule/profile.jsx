import { Avatar } from "../atoms/avatar.jsx";
import PropTypes from "prop-types";

export const Profile = ({ name, subtitle, color }) => {
  return <div className={"inline-block flex py-0.5 px-1 gap-2 items-center"}>
    <Avatar label={name} color={color} size={"medium"} />
    <div className="flex flex-col gap-y-1">
      <div className={"bodyMd-medium"}>{name}</div>
      {subtitle && <div className={"bodySm text-text-soft"}>{subtitle}</div>}
    </div>
  </div>
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
};

Avatar.defaultProps = {
  name: "test",
  subtitle: "subtitle",
  color: "one"
};

