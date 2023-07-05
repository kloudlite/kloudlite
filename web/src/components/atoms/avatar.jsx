import React from 'react';
import PropTypes from 'prop-types';
import classNames from "classnames";

const colors = {
  "one": [
    "#D4D4D8",
    "#FFFFFF",
    "#8C9196",
    "#111827"
  ],
  "two": [
    "#D97706",
    "#FDE68A",
    "#b45309",
    "#b45309"
  ],
  "three": [
    "#16A34A",
    "#BBF7D0",
    "#15803d",
    "#15803d",
  ],
  "four": [
    "#DC2626",
    "#FECACA",
    "#B91C1C",
    "#B91C1C",
  ],
  "five": [
    "#0d9488",
    "#99F6E4",
    "#0f766e",
    "#0f766e",
  ],
}


export const AvatarBase = ({ label, size, color }) => {
  const ab = (() => {
    const words = label.split(' ');
    return words.map(word => word.charAt(0).toUpperCase()).join('');
  })

  return <div
    className={
      classNames(
        "relative",
        "outline-none transition-all",
        "rounded-full",
        {
          "w-16 h-16": size === "large",
          "w-10 h-10": size === "medium",
          "w-8 h-8": size === "small",
          "w-6 h-6": size === "extra-small",
        }
      )
    }
  >
    {!label && <svg className={classNames(
      {
        "w-16 h-16": size === "large",
        "w-10 h-10": size === "medium",
        "w-8 h-8": size === "small",
        "w-6 h-6": size === "extra-small",
      }
    )} viewBox="0 0 62 61" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="1" y="0.5" width="60" height="60" rx="30" fill={colors[color][1]} />
      <path fillRule="evenodd" clipRule="evenodd" d="M31.0002 27.3571C36.0446 27.3571 40.1339 23.1358 40.1339 17.9286C40.1339 12.7213 36.0446 8.5 31.0002 8.5C25.9558 8.5 21.8665 12.7213 21.8665 17.9286C21.8665 23.1358 25.9558 27.3571 31.0002 27.3571ZM31.0002 52.5C38.6749 52.5 45.5322 48.8162 50 43.0717C45.5322 37.3269 38.6747 33.6429 30.9998 33.6429C23.3251 33.6429 16.4678 37.3267 12 43.0712C16.4678 48.816 23.3253 52.5 31.0002 52.5Z" fill={colors[color][2]} />
      <rect x="1" y="0.5" width="60" height="60" rx="30" stroke={colors[color][0]} />
    </svg>}
    {label &&
      <div className={classNames(
        {
          "w-16 h-16": size === "large",
          "w-10 h-10": size === "medium",
          "w-8 h-8": size === "small",
          "w-6 h-6": size === "extra-small",
        }, "relative"
      )}>
        <svg className={classNames(
          {
            "w-16 h-16": size === "large",
            "w-10 h-10": size === "medium",
            "w-8 h-8": size === "small",
            "w-6 h-6": size === "extra-small",
          }
        )} viewBox="0 0 62 61" fill="none" xmlns="http://www.w3.org/2000/svg">
          <rect x="1" y="0.5" width="60" height="60" rx="30" fill={colors[color][1]} />
          <rect x="1" y="0.5" width="60" height="60" rx="30" stroke={colors[color][0]} />
        </svg>
        <div className={classNames("absolute top-0 bottom-0 left-0 right-0 flex justify-center items-center", {
          "headingLg": size === "large",
          "bodyLg": size === "medium" || size === "small",
          "bodySm": size === "extra-small",
        })} style={{ color: colors[color][3] }}>{ab()}</div>
      </div>
    }
  </div>
}

export const Avatar = ({ label, size, color }) => {
  return <AvatarBase label={label} size={size} color={color} />
};

Avatar.propTypes = {
  label: PropTypes.string.isRequired,
  size: PropTypes.oneOf([
    "large", "medium", "small", "extra-small",
  ]),
  color: PropTypes.oneOf([
    "one",
    "two",
    "three",
    "four",
    "five"
  ]),
};

Avatar.defaultProps = {
  label: "test",
  size: "medium",
  color: "one"
};
