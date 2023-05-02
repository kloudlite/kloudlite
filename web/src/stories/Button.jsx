import React from 'react';
import PropTypes from 'prop-types';
import classnames from "classnames";
import BounceIt from "../components/bounce-it.jsx";

const Anchor = ({href, children, ...props}) => {
  return (
    <a href={href} {...props}>
      {children}
    </a>
  )
}

const ButtonElement = ({type, children, ...props}) => {
  return (
    <button type={type} {...props}>
      {children}
    </button>
  )
}

/**
 * Button component for user interaction
 */
export const Button = ({style, size="medium", onClick, href, label, type, disabled, Component, sharpLeft, sharpRight, nobounce}) => {
  const C = Component || (href ? Anchor : ButtonElement)
  return (
    <BounceIt disable={nobounce}>
      <C
        type={type}
        disabled={disabled}
        className={classnames(
          "bodyMd-medium",
          {
            "rounded-none":sharpLeft && sharpRight,
            "rounded-l":sharpLeft && !sharpRight,
            "rounded-r":!sharpLeft && sharpRight,
            "rounded-full":!sharpLeft && !sharpRight,
          },
          {
            "border-l border-r border-t border-b":sharpLeft && sharpRight,
            "border-l border-t border-b":sharpLeft && !sharpRight,
            "border-r border-t border-b":!sharpLeft && sharpRight,
            "border-t border-b":!sharpLeft && !sharpRight,
          },
          "transition-all",
          "disabled:pointer-events-none",
          {
            "shadow-button":style !== "plain" && style !== "outline" && style !== "primary-plain" && style !== "secondary-plain" && style !== "critical-plain",
          },
          {
            "border-border-default disabled:border-border-disabled":style === "basic" || style === "outline",
            "border-primary disabled:border-border-disabled":style === "primary"||style === "primary-outline",
            "border-secondary disabled:border-border-disabled":style === "secondary"||style === "secondary-outline",
            "border-none":style === "plain" || style === "primary-plain" || style === "critical-plain" || style === "secondary-plain",
            "border-critical disabled:border-border-disabled":style === "critical-outline" || style === "critical",
          },
          {
            "bg-surface hover:bg-surface-hovered active:bg-surface-pressed":style === "basic",
            "bg-surface-primary-default hover:bg-surface-primary-hovered active:bg-surface-primary-pressed disabled:bg-surface":style === "primary",
            "bg-surface-secondary-default hover:bg-surface-secondary-hovered active:bg-surface-secondary-pressed disabled:bg-surface":style === "secondary",
            "bg-surface-critical-default hover:bg-surface-critical-hovered active:bg-surface-critical-pressed disabled:bg-surface":style === "critical",
            "bg-none hover:bg-surface-critical-subdued active:bg-surface-critical-pressed":style === "critical-outline",
            "bg-none hover:bg-surface-primary-subdued active:bg-surface-primary-pressed":style === "primary-outline",
            "bg-none hover:bg-surface-secondary-subdued active:bg-surface-secondary-pressed":style === "secondary-outline",
            "bg-none hover:bg-surface-hovered active:bg-surface-pressed":style === "outline",
            "bg-none active:bg-surface-pressed":style === "plain",
            "bg-none active:bg-surface-primary-pressed":style === "primary-plain",
            "bg-none active:bg-surface-secondary-pressed":style === "secondary-plain",
            "bg-none active:bg-surface-critical-pressed":style === "critical-plain",
          },
          {
            "text-text-default disabled:text-text-disabled":style === "basic" || style==="plain" || style === "outline",
            "active:text-text-on-primary":style === "primary-plain" || style === "critical-plain" || style === "secondary-plain",
            "text-text-on-primary disabled:text-text-disabled":style === "primary" || style === "critical",
            "text-text-on-secondary disabled:text-text-disabled":style === "secondary",
            "text-text-critical active:text-text-on-primary disabled:text-text-disabled":style === "critical-outline" || style === "critical-plain",
            "text-text-primary active:text-text-on-primary disabled:text-text-disabled":style === "primary-outline"|| style === "primary-plain",
            "text-text-secondary active:text-text-on-secondary disabled:text-text-disabled":style === "secondary-outline" || style === "secondary-plain",
          },
          {
            "underline":style === "plain"|| style === "primary-plain" || style === "critical-plain" || style === "secondary-plain",
          },
          {
            "px-6 py-3" : size === "large" && style !== "plain" && style!== "critical-plain" && style !== "primary-plain" && style !== "secondary-plain",
            "px-4 py-2" : size === "medium" && style !== "plain"&& style!== "critical-plain" && style !== "primary-plain" && style !== "secondary-plain",
            "px-2 py-1" : size === "small" && style !== "plain"&& style!== "critical-plain" && style !== "primary-plain" && style !== "secondary-plain",
            "px-1 py-0.5" : style === "plain"|| style === "primary-plain" || style === "critical-plain" || style === "secondary-plain",
          }
        )}
        onClick={onClick}
        href={href}
      >
        {label}
      </C>
    </BounceIt>
  );
};

Button.propTypes = {
  /**
   * How the button looks like?
   */
  style: PropTypes.oneOf([
    'outline',
    'basic',
    'plain',
    'primary',
    'primary-outline',
    'secondary',
    'secondary-outline',
    'critical',
    'critical-outline',
    'primary-plain',
    'secondary-plain',
    'critical-plain',
  ]),
  /**
   * How large should the button be?
   */
  size: PropTypes.oneOf(['small', 'medium', 'large']),
  /**
   * Button contents
   */
  label: PropTypes.string.isRequired,
  /**
   * Optional click handler
   */
  onClick: PropTypes.func,
  /**
   * Href for link
   */
  href: PropTypes.string,
  /**
   * Disable button
   */
  disabled: PropTypes.bool,
};

Button.defaultProps = {
  style: 'primary',
  size: 'medium',
  onClick: undefined,
  link: false,
};
