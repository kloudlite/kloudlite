import React, { forwardRef } from 'react';
import PropTypes from 'prop-types';
import classnames from "classnames";
import { Link } from '@remix-run/react';
import { BounceIt } from '../bounce-it';

export const ButtonStyles = [
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
]
export const IconButtonStyles = [
  'outline',
  'basic',
  'plain'
]

export const AriaButton = "button"

export const ButtonBase = forwardRef(({
  variant,
  size = "medium",
  onClick,
  href,
  label,
  type,
  disabled,
  sharpLeft = false,
  sharpRight = false,
  noBorder,
  className,
  noRounded,
  noRing,
  IconComp,
  DisclosureComp,
  iconOnly = false,
  selected = false,
  ...props
}, ref) => {

  let Component = "button"
  let extraProps = {}

  extraProps.onClick = onClick

  if (href) {
    Component = Link
    extraProps.to = href
  } else {
    extraProps.disabled = disabled
  }

  return (
    <BounceIt className={className}>
      <Component
        ref={ref}
        type={type}
        className={classnames(
          "w-full",
          {
            "bodyMd-medium": variant !== "primary-plain" && variant !== "secondary-plain" && variant !== "critical-plain" && variant !== "plain",
            "bodyMd": variant === "primary-plain" || variant === "secondary-plain" || variant !== "critical-plain" || variant !== "plain",
          },
          "relative ring-offset-1",
          "outline-none shadow-button",
          "flex flex-row gap-2 items-center justify-center",
          "disabled:text-text-disabled",
          {
            ...(noRing ? {} : {
              "focus-visible:ring-2 focus:ring-border-focus focus:z-10": true
            })
          },
          {
            ...(noRounded ? {} : {
              "rounded-none": sharpLeft && sharpRight,
              "rounded-r": sharpLeft && !sharpRight,
              "rounded-l": !sharpLeft && sharpRight,
              "rounded": !sharpLeft && !sharpRight,
            })
          },
          "transition-all",
          "disabled:pointer-events-none",
          {
            ...(noBorder ? { "border-none": true } : {
              "border-border-default disabled:border-border-disabled": variant === "basic" || variant === "outline" || variant === "secondary-outline",
              "border-border-primary disabled:border-border-disabled": variant === "primary" || variant === "primary-outline",
              "border-border-secondary disabled:border-border-disabled": variant === "secondary",
              "border-border-danger disabled:border-border-disabled": variant === "critical-outline" || variant === "critical",
              "border-none": variant === "plain" || variant === "primary-plain" || variant === "critical-plain" || variant === "secondary-plain",
              "border": !(variant === "plain" || variant === "primary-plain" || variant === "critical-plain" || variant === "secondary-plain"),
            })
          },
          {
            "bg-surface-default hover:bg-surface-hovered active:bg-surface-pressed": variant === "basic",
            "bg-surface-pressed hover:bg-surface-pressed active:bg-surface-pressed": variant === "basic" && selected,
            "bg-surface-primary-default hover:bg-surface-primary-hovered active:bg-surface-primary-pressed disabled:bg-surface-default": variant === "primary",
            "bg-surface-secondary-default hover:bg-surface-secondary-hovered active:bg-surface-secondary-pressed disabled:bg-surface-default": variant === "secondary",
            "bg-surface-danger-default hover:bg-surface-danger-hovered active:bg-surface-danger-pressed disabled:bg-surface-default": variant === "critical",
            "bg-none shadow-none hover:bg-surface-danger-subdued active:bg-surface-danger-pressed hover:shadow-button active:shadow-button": variant === "critical-outline",
            "bg-none shadow-none hover:bg-surface-primary-subdued active:bg-surface-primary-pressed hover:shadow-button active:shadow-button": variant === "primary-outline",
            "bg-none shadow-none hover:bg-surface-secondary-hovered active:bg-surface-secondary-pressed hover:shadow-button active:shadow-button": variant === "secondary-outline",
            "bg-none shadow-none hover:bg-surface-hovered active:bg-surface-pressed hover:shadow-button active:shadow-button": variant === "outline",
            "bg-none shadow-none active:bg-surface-pressed active:shadow-button": variant === "plain" && !iconOnly,
            "bg-none shadow-none hover:bg-surface-hovered active:bg-surface-pressed active:shadow-button": variant === "plain" && iconOnly,
            "bg-none shadow-none active:bg-surface-primary-pressed active:shadow-button": variant === "primary-plain",
            "bg-none shadow-none active:bg-surface-secondary-pressed active:shadow-button": variant === "secondary-plain",
            "bg-none shadow-none active:bg-surface-danger-pressed active:shadow-button": variant === "critical-plain",
          },
          {
            "text-text-default": (variant === "basic" || variant === "plain" || variant === "outline"),
            "active:text-text-on-primary": (variant === "primary-plain" || variant === "critical-plain" || variant === "secondary-plain"),
            "text-text-on-primary": variant === "primary" || variant === "critical" || variant === "secondary" || variant === "secondary-outline",
            "text-text-danger": (variant === "critical-outline" || variant === "critical-plain"),
            "text-text-primary": (variant === "primary-outline" || variant === "primary-plain"),
            "text-text-secondary": variant === "secondary-plain",
          },
          {
            "focus:underline": noRing
          },
          {
            "hover:underline": variant === "plain" || variant === "primary-plain" || variant === "critical-plain" || variant === "secondary-plain",
          },
          {
            ...(iconOnly ? {
              "p-2.75": size === "large" && variant != 'plain',
              "p-1.75": size === "medium" && variant != 'plain',
              "p-0.75": size === "small" && variant != 'plain',
              "p-3": size === "large" && variant == 'plain',
              "p-2": size === "medium" && variant == 'plain',
              "p-1": size === "small" && variant == 'plain'
            } : {
              "px-6 py-2.75": size === "large" && variant !== "plain" && variant !== "critical-plain" && variant !== "primary-plain" && variant !== "secondary-plain",
              "px-4 py-1.75": size === "medium" && variant !== "plain" && variant !== "critical-plain" && variant !== "primary-plain" && variant !== "secondary-plain",
              "px-2 py-0.75": size === "small" && variant !== "plain" && variant !== "critical-plain" && variant !== "primary-plain" && variant !== "secondary-plain",
              "px-1 py-0.5": variant === "plain" || variant === "primary-plain" || variant === "critical-plain" || variant === "secondary-plain",
            })
          }
        )}
        {...props}
        {...extraProps}
      >
        {IconComp && <IconComp size={iconOnly ? 20 : 16} color="currentColor" />}
        {!iconOnly && label}
        {DisclosureComp && !iconOnly && <DisclosureComp size={16} color="currentColor" />}

      </Component>
    </BounceIt >
  );
})


export const IconButton = ({
  variant,
  size = "medium",
  onClick,
  href,
  type,
  disabled,
  sharpLeft = false,
  sharpRight = false,
  className,
  noRounded,
  noRing,
  IconComp,
  ...props
}) => {

  return <ButtonBase {...props} iconOnly={true} label={''} variant={variant} size={size} onClick={onClick} href={href} type={type} disabled={disabled} sharpLeft={sharpLeft} sharpRight={sharpRight} noRing={noRing} noRounded={noRounded} IconComp={IconComp} className={className} />
}


export const Button = forwardRef(({
  label,
  variant,
  size = "medium",
  onClick,
  href,
  type,
  disabled,
  sharpLeft = false,
  sharpRight = false,
  className,
  noRounded,
  noBorder,
  noRing,
  IconComp,
  DisclosureComp,
  ...props
}, ref) => {
  console.log(variant);
  return <ButtonBase ref={ref} {...props} label={label} noBorder={noBorder} DisclosureComp={DisclosureComp} variant={variant} size={size} onClick={onClick} href={href} type={type} disabled={disabled} sharpLeft={sharpLeft} sharpRight={sharpRight} noRing={noRing} noRounded={noRounded} IconComp={IconComp} className={className} />
})

Button.propTypes = {
  /**
   * How the button looks like?
   */
  variant: PropTypes.oneOf([
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
  type: PropTypes.oneOf(["button", "submit"]),
};

Button.defaultProps = {
  variant: 'primary',
  size: 'medium',
  onClick: undefined,
  type: "button",
};



IconButton.propTypes = {
  /**
   * How the button looks like?
   */
  variant: PropTypes.oneOf(IconButtonStyles),
  /**
   * How large should the button be?
   */
  size: PropTypes.oneOf(['small', 'medium', 'large']),
  /**
   * Button contents
   */
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
  type: PropTypes.oneOf(["button", "submit"]),
};

IconButton.defaultProps = {
  variant: 'basic',
  size: 'medium',
  onClick: undefined,
  type: "button",
};
