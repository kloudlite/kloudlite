import { PropTypes } from 'prop-types';
import {Menu as MenuComp} from "@headlessui/react";
import classNames from 'classnames';
import { createPortal } from 'react-dom';
import React, { useEffect, useRef, useState } from 'react';
import {motion} from "framer-motion";
import { AnimatePresence } from 'framer-motion';



export const Menu = ({items, children})=>{
  const ref = useRef();
  const [pos, setPos] = useState({left: 0, top: 0})
  useEffect(()=>{
    if(ref.current){
      const rect = ref.current.getBoundingClientRect();
      setPos({
        left: rect.x,
        // + rect.width / 2,
        top: rect.y + rect.height + 4
        // + window.scrollY
      });
    }
  }, [ref.current])
  return (<MenuComp>
    <MenuComp.Button as={"div"} ref={ref}>
      {children}
    </MenuComp.Button>
    {createPortal(
      <MenuComp.Items static={true}>
          {({open})=>{
            return <AnimatePresence>
              {open && <motion.div
                style={pos}
                className={"flex flex-col rounded border bg-surface-default border-border-default absolute shadow-popover outline-none overflow-clip"} 
                initial={{opacity: 0,  translateY: 3, scale: 0.99}}
                animate={{opacity: 1, translateY: 0, scale: 1}}
                exit={{opacity: 0,  translateY: 3, scale: 0.99}}
                transition={{
                  duration: 0.3,
                  ease: 'anticipate',
                }}
              >
                {
                    items.map((item)=>{
                      return <MenuComp.Item key={item.value} className={"px-3 py-2 cursor-pointer bodyMd"}>
                        {
                          ({active})=>{
                            return <div className={classNames({
                              "bg-surface-hovered": active,
                            })}>{item.Element}</div>
                          }
                        }
                      </MenuComp.Item>
                    })
                  }
              </motion.div>}
            </AnimatePresence>
            
          }}
      </MenuComp.Items>
      , document.body)}
  </MenuComp>)
}

Menu.propTypes = {
  items: PropTypes.arrayOf(PropTypes.shape({
    label: PropTypes.string,
    value: PropTypes.string,
  })),
  value: PropTypes.string,
  placeholder: PropTypes.string,
  onChange: PropTypes.func,
}

Menu.defaultProps = {
  items: [],
  value: "",
  placeholder: "",
  onChange: ()=>{},
}