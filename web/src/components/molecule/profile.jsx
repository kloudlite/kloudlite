import {Avatar} from "../atoms/avatar.jsx";
import PropTypes from "prop-types";

export const Profile = ({name, color})=>{
  return <div className={"inline-block flex py-0.5 px-1 gap-2 items-center"}>
    <Avatar label={name} color={color} size={"medium"}/>
    <div className={"bodyMd-medium"}>{name}</div>
  </div>
}

Profile.propTypes = {
  name: PropTypes.string.isRequired,
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
  color: "one"
};

