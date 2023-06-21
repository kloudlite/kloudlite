import React, { forwardRef } from 'react';
import PropTypes from 'prop-types';
import classnames from "classnames";
import { BounceIt } from "../bounce-it.jsx";
import { Link } from 'react-router-dom';
import { useButton } from 'react-aria';
import { useFocusRing } from 'react-aria';

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

export const AriaButton = forwardRef(({ className, ...props }, ref) => {
  let { buttonProps } = useButton(props, ref);
  return <button {...buttonProps} ref={ref} className={className}>{props.children}</button>;
})


export const ButtonBase = forwardRef(({
  style,
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

  let { isFocusVisible, focusProps } = useFocusRing();


  if (href) {
    return (
      <BounceIt disable={disabled} className='focus-within:z-10'>
        <Link
          ref={ref}
          to={href}
          className={classnames(
            className,
            {
              "bodyMd-medium": style !== "primary-plain" && style !== "secondary-plain" && style !== "critical-plain" && style !== "plain",
              "bodyMd": style === "primary-plain" || style === "secondary-plain" || style !== "critical-plain" || style !== "plain",
            },
            "ring-offset-1",
            "outline-none shadow-button",
            "flex gap-2 items-center",
            {
              ...(noRing ? {} : {
                "focus-visible:ring-2 focus:ring-border-focus": true,
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
                "border-border-default disabled:border-border-disabled": style === "basic" || style === "outline",
                "border-border-primary disabled:border-border-disabled": style === "primary" || style === "primary-outline",
                "border-border-secondary disabled:border-border-disabled": style === "secondary" || style === "secondary-outline",
                "border-border-danger disabled:border-border-disabled": style === "critical-outline" || style === "critical",
                "border-none": style === "plain" || style === "primary-plain" || style === "critical-plain" || style === "secondary-plain",
                "border": !(style === "plain" || style === "primary-plain" || style === "critical-plain" || style === "secondary-plain"),
              })
            },
            {
              "bg-surface-default hover:bg-surface-hovered active:bg-surface-pressed": style === "basic",
              "bg-surface-pressed hover:bg-surface-pressed active:bg-surface-pressed": style === "basic" && selected,
              "bg-surface-primary-default hover:bg-surface-primary-hovered active:bg-surface-primary-pressed disabled:bg-surface-default": style === "primary",
              "bg-surface-secondary-default hover:bg-surface-secondary-hovered active:bg-surface-secondary-pressed disabled:bg-surface-default": style === "secondary",
              "bg-surface-danger-default hover:bg-surface-danger-hovered active:bg-surface-danger-pressed disabled:bg-surface-default": style === "critical",
              "bg-none shadow-none hover:bg-surface-danger-subdued active:bg-surface-danger-selected hover:shadow-button active:shadow-button": style === "critical-outline",
              "bg-none shadow-none hover:bg-surface-primary-subdued active:bg-surface-primary-selected hover:shadow-button active:shadow-button": style === "primary-outline",
              "bg-none shadow-none hover:bg-surface-secondary-subdued active:bg-surface-secondary-selected hover:shadow-button active:shadow-button": style === "secondary-outline",
              "bg-none shadow-none hover:bg-surface-hovered active:bg-surface-pressed hover:shadow-button active:shadow-button": style === "outline",
              "bg-none shadow-none active:bg-surface-pressed active:shadow-button": style === "plain" && !iconOnly,
              "bg-none shadow-none hover:bg-surface-hovered active:bg-surface-pressed active:shadow-button": style === "plain" && iconOnly,
              "bg-none shadow-none active:bg-surface-primary-pressed active:shadow-button": style === "primary-plain",
              "bg-none shadow-none active:bg-surface-secondary-pressed active:shadow-button": style === "secondary-plain",
              "bg-none shadow-none active:bg-surface-danger-pressed active:shadow-button": style === "critical-plain",
            },
            {
              "text-text-default disabled:text-text-disable": (style === "basic" || style === "plain" || style === "outline"),
              "active:text-text-on-primary": (style === "primary-plain" || style === "critical-plain" || style === "secondary-plain"),
              "text-text-on-primary disabled:text-text-disabled": style === "primary" || style === "critical" || style === "secondary",
              "text-text-danger disabled:text-text-disabled": ((style === "critical-outline" || style === "critical-plain")),
              "text-text-primary disabled:text-text-disabled": ((style === "primary-outline" || style === "primary-plain")),
              "text-text-secondary disabled:text-text-disabled": ((style === "secondary-outline" || style === "secondary-plain")),
            },
            {
              "focus:underline": noRing
            },
            {
              "hover:underline": style === "plain" || style === "primary-plain" || style === "critical-plain" || style === "secondary-plain",
            },
            {
              "px-6 py-3": size === "large" && style !== "plain" && style !== "critical-plain" && style !== "primary-plain" && style !== "secondary-plain",
              "px-4 py-2": size === "medium" && style !== "plain" && style !== "critical-plain" && style !== "primary-plain" && style !== "secondary-plain",
              "px-2 py-1": size === "small" && style !== "plain" && style !== "critical-plain" && style !== "primary-plain" && style !== "secondary-plain",
              "px-1 py-0.5": style === "plain" || style === "primary-plain" || style === "critical-plain" || style === "secondary-plain",
            }
          )}
          onClick={onClick}
          href={href}
        >
          {IconComp && <IconComp size={iconOnly ? 20 : 16} color="currentColor" />}
          {!iconOnly && label}
          {DisclosureComp && !iconOnly && <DisclosureComp size={16} color="currentColor" />}
        </Link>
      </BounceIt>
    )
  }
  return (
    <AriaButton
      {...focusProps}
      ref={ref}
      onPress={onClick}
      type={type}
      isDisabled={disabled}
      className={classnames(
        className,
        {
          "bodyMd-medium": style !== "primary-plain" && style !== "secondary-plain" && style !== "critical-plain" && style !== "plain",
          "bodyMd": style === "primary-plain" || style === "secondary-plain" || style !== "critical-plain" || style !== "plain",
        },
        "ring-offset-1",
        "outline-none shadow-button",
        "flex gap-2 items-center",
        "disabled:text-text-disabled",
        {
          ...(noRing ? {} : {
            "focus-visible:ring-2 focus:ring-border-focus z-10": isFocusVisible,
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
            "border-border-default disabled:border-border-disabled": style === "basic" || style === "outline" || style === "secondary-outline",
            "border-border-primary disabled:border-border-disabled": style === "primary" || style === "primary-outline",
            "border-border-secondary disabled:border-border-disabled": style === "secondary",
            "border-border-danger disabled:border-border-disabled": style === "critical-outline" || style === "critical",
            "border-none": style === "plain" || style === "primary-plain" || style === "critical-plain" || style === "secondary-plain",
            "border": !(style === "plain" || style === "primary-plain" || style === "critical-plain" || style === "secondary-plain"),
          })
        },
        {
          "bg-surface-default hover:bg-surface-hovered active:bg-surface-pressed": style === "basic",
          "bg-surface-pressed hover:bg-surface-pressed active:bg-surface-pressed": style === "basic" && selected,
          "bg-surface-primary-default hover:bg-surface-primary-hovered active:bg-surface-primary-pressed disabled:bg-surface-default": style === "primary",
          "bg-surface-secondary-default hover:bg-surface-secondary-hovered active:bg-surface-secondary-pressed disabled:bg-surface-default": style === "secondary",
          "bg-surface-danger-default hover:bg-surface-danger-hovered active:bg-surface-danger-pressed disabled:bg-surface-default": style === "critical",
          "bg-none shadow-none hover:bg-surface-danger-subdued active:bg-surface-danger-pressed hover:shadow-button active:shadow-button": style === "critical-outline",
          "bg-none shadow-none hover:bg-surface-primary-subdued active:bg-surface-primary-pressed hover:shadow-button active:shadow-button": style === "primary-outline",
          "bg-none shadow-none hover:bg-surface-secondary-hovered active:bg-surface-secondary-pressed hover:shadow-button active:shadow-button": style === "secondary-outline",
          "bg-none shadow-none hover:bg-surface-hovered active:bg-surface-pressed hover:shadow-button active:shadow-button": style === "outline",
          "bg-none shadow-none active:bg-surface-pressed active:shadow-button": style === "plain" && !iconOnly,
          "bg-none shadow-none hover:bg-surface-hovered active:bg-surface-pressed active:shadow-button": style === "plain" && iconOnly,
          "bg-none shadow-none active:bg-surface-primary-pressed active:shadow-button": style === "primary-plain",
          "bg-none shadow-none active:bg-surface-secondary-pressed active:shadow-button": style === "secondary-plain",
          "bg-none shadow-none active:bg-surface-danger-pressed active:shadow-button": style === "critical-plain",
        },
        {
          "text-text-default": (style === "basic" || style === "plain" || style === "outline"),
          "active:text-text-on-primary": (style === "primary-plain" || style === "critical-plain" || style === "secondary-plain"),
          "text-text-on-primary": style === "primary" || style === "critical" || style === "secondary" || style === "secondary-outline",
          "text-text-danger": (style === "critical-outline" || style === "critical-plain"),
          "text-text-primary": (style === "primary-outline" || style === "primary-plain"),
          "text-text-secondary": style === "secondary-plain",
        },
        {
          "focus:underline": noRing
        },
        {
          "hover:underline": style === "plain" || style === "primary-plain" || style === "critical-plain" || style === "secondary-plain",
        },
        {
          ...(iconOnly ? {
            "p-2.75": size === "large" && style != 'plain',
            "p-1.75": size === "medium" && style != 'plain',
            "p-0.75": size === "small" && style != 'plain',
            "p-3": size === "large" && style == 'plain',
            "p-2": size === "medium" && style == 'plain',
            "p-1": size === "small" && style == 'plain'
          } : {
            "px-6 py-2.75": size === "large" && style !== "plain" && style !== "critical-plain" && style !== "primary-plain" && style !== "secondary-plain",
            "px-4 py-1.75": size === "medium" && style !== "plain" && style !== "critical-plain" && style !== "primary-plain" && style !== "secondary-plain",
            "px-2 py-0.75": size === "small" && style !== "plain" && style !== "critical-plain" && style !== "primary-plain" && style !== "secondary-plain",
            "px-1 py-0.5": style === "plain" || style === "primary-plain" || style === "critical-plain" || style === "secondary-plain",
          })
        }
      )}
      href={href}
      {...props}
    >
      {IconComp && <IconComp size={iconOnly ? 20 : 16} color="currentColor" />}
      {!iconOnly && label}
      {DisclosureComp && !iconOnly && <DisclosureComp size={16} color="currentColor" />}
    </AriaButton>
  );
})


export const IconButton = forwardRef(({
  style,
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
}, ref) => {
  return <ButtonBase ref={ref} iconOnly={true} label={''} style={style} size={size} onClick={onClick} href={href} type={type} disabled={disabled} sharpLeft={sharpLeft} sharpRight={sharpRight} noRing={noRing} noRounded={noRounded} IconComp={IconComp} className={className} />
})


export const Button = forwardRef(({
  label,
  style,
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
  console.log(ref);
  return <ButtonBase {...props} ref={ref} label={label} noBorder={noBorder} DisclosureComp={DisclosureComp} style={style} size={size} onClick={onClick} href={href} type={type} disabled={disabled} sharpLeft={sharpLeft} sharpRight={sharpRight} noRing={noRing} noRounded={noRounded} IconComp={IconComp} className={className} />
})

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
  type: PropTypes.oneOf(["button", "submit"]),
};

Button.defaultProps = {
  style: 'primary',
  size: 'medium',
  onClick: undefined,
  link: false,
  type: "button",
};



IconButton.propTypes = {
  /**
   * How the button looks like?
   */
  style: PropTypes.oneOf(IconButtonStyles),
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
  style: 'basic',
  size: 'medium',
  onClick: undefined,
  link: false,
  type: "button",
};
