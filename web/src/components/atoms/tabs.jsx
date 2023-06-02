import { useState, cloneElement } from "react"
import { LayoutGroup } from "framer-motion"
import { useEffect } from "react";
import { useLink } from "react-aria";
import { Link } from "react-router-dom";
import classNames from "classnames";
import { motion } from "framer-motion";
import PropTypes from "prop-types";


export const NavTab = ({ href, label, onPress, active, fitted }) => {
  const { linkProps, isPressed } = useLink({ href, onPress })
  return <div className={classNames("outline-none flex flex-col relative group bodyMd-medium hover:text-text-default active:text-text-default ",
    {
      "text-text-default": active,
      "text-text-soft": !active
    })}>
    <Link {...linkProps} to={href} className={classNames("outline-none flex flex-col",
      {
        "p-4": !fitted,
        "pt-2 pb-3": fitted
      })}>
      {label}
    </Link>
    {
      active && <motion.div layoutId="underline" className={classNames("h-1 bg-surface-primary-pressed z-10")}></motion.div>
    }
    <div className="h-1 group-hover:bg-surface-hovered group-active:bg-surface-pressed bg-none transition-all bottom-0 absolute w-full z-0"></div>
  </div>
}

export const NavTabs = ({ children, fitted, onChange, layoutId }) => {
  const [active, setActive] = useState(0);
  useEffect(() => {
    if (onChange) {
      onChange(active)
    }
  }, [active])
  return <div className="flex flex-row gap-6">
    <LayoutGroup id={layoutId}>
      {children.map((child, index) => {
        return cloneElement(child, {
          onPress: () => {
            setActive(index)
          },
          fitted,
          key: index,
          active: active === index,
        })
      })}
    </LayoutGroup>
  </div>
}

NavTab.propTypes = {
  href: PropTypes.string,
  label: PropTypes.string.isRequired,
  onPress: PropTypes.func,
  active: PropTypes.bool,
  fitted: PropTypes.bool
}

NavTabs.propTypes = {
   /**
   * LayoutId should be provided in order to prevent multiple tabs to share same instance.
   */
  layoutId: PropTypes.string.isRequired,
  children: PropTypes.arrayOf(PropTypes.object.isRequired).isRequired,
  fitted: PropTypes.bool,
  onChange: PropTypes.func,
}


NavTab.defaultProps = {
  label: "Item",
  active: false,
  fitted: false,
}
