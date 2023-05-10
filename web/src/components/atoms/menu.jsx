import classNames from 'classnames';
import { PropTypes } from 'prop-types';
import { useState, useRef, useEffect } from 'react';
import { BounceIt } from '../bounce-it';
import { AnimatePresence, motion } from 'framer-motion';
import { createPortal } from 'react-dom';

const Popover = ({children})=>{
  return createPortal(<div>{children}</div>, document.body)
}

export const Menu = ({items, value, onChange, placeholder})=>{
  const [open, setOpen] = useState(false)
  const ref = useRef()
  const [currentRect, setCurrentRect] = useState(null)
  const [currentValue, setCurrentValue] = useState(value)
  useEffect(()=>{
    onChange(currentValue)
  }, [value])
  useEffect(()=>{
    if(ref.current){
      setCurrentRect(ref.current.getBoundingClientRect())
      console.log(currentRect)
    }
  }, [ref.current])
  return <div className="flex flex-col gap-2 max-w-md">
    <BounceIt onClick={()=>{
      setOpen(!open)
    }}>
      <div ref={ref} className={classNames(
        'select-none transition-all cursor-pointer px-3 py-2 hover:bg-zinc-200 rounded',
        {
          'bg-primary-200': open,
        }
      )}>{currentValue || placeholder}</div>
    </BounceIt>
    <div className='relative'>
      <Popover>
        <AnimatePresence>
        {
          open && <motion.div 
          initial={{
            opacity: 0,
            translateY: 10,
          }}
          animate={{ opacity: 1, translateY: 0 }}
          exit={{ opacity: 0, translateY: 10 }}
          transition={{
            ease: 'anticipate',
            duration: 0.3,
          }}
          className={classNames(
            "absolute top-0 left-0 right-0 z-10",
            "bg-grey-50 rounded-md shadow-md",
            "overflow-hidden",
            "border border-grey-300",
            "divide-y divide-grey-300",
            "divide-solid",
            "divide-opacity-50",
          )} style={{width:currentRect.width, x:currentRect.x, y:currentRect.y+currentRect.height+3}}>
            {items.map((item)=>{
              return <div key={item.value} className={
                classNames(
                  "flex gap-2 px-3 py-2 hover:bg-zinc-100 transition-all cursor-pointer",
                  {
                    "bg-zinc-100": item.value === currentValue,
                  }
                )
              } onClick={()=>{
                setCurrentValue(item.value)
                setOpen(false)
              }}>
                <div>{item.label}</div>
              </div>
            })}
          </motion.div>
        }
        </AnimatePresence>
      </Popover>
    </div>
    
  </div>
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