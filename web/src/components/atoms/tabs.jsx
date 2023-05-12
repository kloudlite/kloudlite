import classNames from "classnames"
import { useState, cloneElement } from "react"
import { motion } from "framer-motion"
import { LayoutGroup } from "framer-motion"
import {useLink, useButton} from 'react-aria';
import { useRef, useEffect } from "react";
import {
  TabContent,
  TabIndicator,
  TabList,
  Tabs,
  TabTrigger,
} from '@ark-ui/react'


// export const Tab = ({label, href, onPress, fitted, active})=>{
//   const ref = useRef();
//   const {linkProps} = useLink({label, href}, ref);
//   const {buttonProps} = useButton({label, href, onPress}, ref);
//   const C = href ? "a" : "button";
//   return <div className="flex flex-col">
//       <C 
//       {...(href ? linkProps : buttonProps)}
//       ref={ref}
//       href={href} 
//       className={classNames(
//         "focus-visible:ring-2 ring-offset-1 rounded",
//         " hover:text-text-default transition-all active:text-text-default outline-none cursor-pointer",
//         {
//           "text-text-default bodyMd": active,
//           "text-text-soft bodyMd": !active,
//         },
//         "flex flex-col",
//         "py-4",
//         {
//           "px-0": fitted,
//           "px-4": !fitted,
//         },
//       )}>
//       {label}
//     </C>
//       {
//         active && <motion.div 
//           className="bg-surface-primary-default h-[3px] hover:bg-surface-primary-hovered active:bg-surface-primary-pressed"
//           layoutId="underline"
//         ></motion.div>
//       }
//   </div>
// }

export const TabsX = ()=>{
  return (<Tabs>
    <TabList>
      <TabTrigger value="react">
        <button>React</button>
      </TabTrigger>
      <TabTrigger value="solid">
        <button>Solid</button>
      </TabTrigger>
      <TabTrigger value="vue">
        <button>Vue</button>
      </TabTrigger>
    </TabList>
    <TabContent value="react">
      A JavaScript library for building user interfaces
    </TabContent>
    <TabContent value="solid">
      Simple and performant reactivity for building user interfaces.
    </TabContent>
    <TabContent value="vue">
      An approachable, performant and versatile framework for building web user
      interfaces.
    </TabContent>
  </Tabs>);
}

export const Tabs2 = ({children, fitted, onChange})=>{
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
