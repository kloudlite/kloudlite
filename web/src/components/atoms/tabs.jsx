import { useState, cloneElement } from "react"
import { LayoutGroup } from "framer-motion"
import { useEffect } from "react";
import { useFocusRing, useLink } from "react-aria";
import { Link } from "@remix-run/react"
import classNames from "classnames";
import { motion } from "framer-motion";
import PropTypes from "prop-types";


export const NavTab = ({ href, label, active, fitted, onClick }) => {

  return <div className={classNames("outline-none flex flex-col relative group bodyMd-medium hover:text-text-default active:text-text-default ",
    {
      "text-text-default": active,
      "text-text-soft": !active
    })}>
    <Link onClick={onClick} to={href} className={classNames("outline-none flex flex-col rounded ring-offset-1 focus-visible:ring-2 focus-visible:ring-border-focus",
      {
        "p-4": !fitted,
        "pt-2 pb-3": fitted,
      })}>
      {label}
    </Link>
    {
      active && <motion.div layoutId="underline" className={classNames("h-1 bg-surface-primary-pressed z-10 absolute bottom-0 w-full")}></motion.div>
    }
    <div className="h-1 group-hover:bg-border-default group-active:bg-border-tertiary bg-none transition-all absolute bottom-0 w-full z-0"></div>
  </div>
}

export const NavTabs = ({ items, fitted, onChange, layoutId, value }) => {

  const [active, setActive] = useState(value);
  useEffect(() => {
    if (onChange) {
      onChange(active)
    }
  }, [active])
  return <div className="flex flex-row gap-6">
    <LayoutGroup id={layoutId}>
      {items.map((child, index) => {
        return <NavTab
          onClick={() => {
            setActive(child.value)
          }}
          fitted={fitted}
          key={child.key}
          href={child.href}
          label={child.label}
          active={active === child.value}
        />
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
  items: PropTypes.arrayOf(PropTypes.object.isRequired).isRequired,
  fitted: PropTypes.bool,
  onChange: PropTypes.func,
  value: PropTypes.any
}

NavTabs.defaultProps = {
  layoutId: "nav-tabs",
}


NavTab.defaultProps = {
  label: "Item",
  active: false,
  fitted: false,
}
