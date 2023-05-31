import React from 'react';
import PropTypes from 'prop-types';
import { Button } from "./button.jsx";
import classnames from "classnames";

export const ButtonGroup = ({ items, size, fullWidth, style }) => {
  return (
    <div className={classnames("flex w-max rounded overflow-hidden border divide-x", {
      "bg-primary-700": style === "primary",
      "bg-secondary-700": style === "secondary",
      "bg-zinc-300": style === "basic",
      "bg-critical-700": style === "critical",
    }, {
      "divide-zinc-300 border-zinc-300 disabled:border-zinc-200": style === "basic" || style === "outline",
      "divide-primary-600 border-primary-600 disabled:border-zinc-200": style === "primary" || style === "primary-outline",
      "divide-secondary-600 border-secondary-600 disabled:border-zinc-200": style === "secondary" || style === "secondary-outline",
      "border-none": style === "plain" || style === "primary-plain" || style === "critical-plain" || style === "secondary-plain",
      "border": !(style === "plain" || style === "primary-plain" || style === "critical-plain" || style === "secondary-plain"),
      "divide-critical-600 border-critical-600 disabled:border-zinc-200": style === "critical-outline" || style === "critical",
    }, {
      "border-zinc-300": style === "basic" || style === "outline",
      "border-primary-600": style === "primary" || style === "primary-outline",
      "border-secondary-600": style === "secondary" || style === "secondary-outline",
      "border-critical-600": style === "critical-outline" || style === "critical",
    })}>
      {
        items.map((item) => {
          return <Button
            noBorder={true}
            key={item.label}
            label={item.label}
            size={size}
            noRounded
            style={style}
          // noRing 
          />
        })
      }
    </div>
  );
};


ButtonGroup.propTypes = {
  items: PropTypes.arrayOf(PropTypes.shape({
    label: PropTypes.string.isRequired,
    onClick: PropTypes.func,
  })).isRequired,
  fullWidth: PropTypes.bool,
  style: PropTypes.oneOf([
    "basic",
    "outline",
    "primary",
    "primary-outline",
    "secondary",
    "secondary-outline",
    "critical",
    "critical-outline"
  ]),
  size: PropTypes.oneOf(["small", "medium", "large"]),
};

ButtonGroup.defaultProps = {
  style: 'primary',
  size: 'medium',
  fullWidth: false,
  items: [
    {
      label: "test",
    },
    {
      label: "test2",
    },
    {
      label: "test3",
    }
  ],
};
