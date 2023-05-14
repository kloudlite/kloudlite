import { useState, cloneElement } from "react"
import { LayoutGroup } from "framer-motion"
import { useEffect } from "react";
import { useLink } from "react-aria";
import { Link } from "react-router-dom";
import classNames from "classnames";
import {motion} from "framer-motion";


export const NavTab = ({href, label, onPress, active})=>{
  const {linkProps, isPressed} = useLink({href, onPress})
  return <div className="flex flex-col relative group">
    <Link {...linkProps} to={href} className="flex flex-col py-4">
      {label}    
    </Link>
    {
      active && <motion.div layoutId="underline" className={classNames("h-1 bg-surface-primary-default z-10")}></motion.div>
    }
    <div class="h-1 group-hover:bg-surface-hovered group-active:bg-surface-pressed bg-none transition-all bottom-0 absolute w-full z-0"></div>
  </div>
}

export const NavTabs = ({children, fitted, onChange})=>{
  const [active, setActive] = useState(0);
  useEffect(()=>{
    if (onChange){
      onChange(active)
    }
  }, [active])
  return <div className="flex flex-row gap-6">
    <LayoutGroup>
    {children.map((child, index)=> {
      return cloneElement(child, {
        onPress: ()=> {
          setActive(index)
        },
        fitted,
        active: active === index,
      })
    })}
    </LayoutGroup>
  </div>
}
