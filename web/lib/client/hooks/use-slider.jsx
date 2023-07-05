import classNames from 'classnames';
import { motion } from 'framer-motion';
import { useState } from 'react';
import { flattenChildren } from '~/root/lib/client/helpers/flat-children';

export const SlideDirectins = {
  TopToBottom: 'tb',
  BottomToTop: 'bt',
  LeftToRight: 'lr',
  RightToLeft: 'rl',
};

export const useSlider = ({
  direction = SlideDirectins.LeftToRight,
  width = '10rem',
  height = '10rem',
  duration = 1,
} = {}) => {
  const [childrenCount, setChildrenCount] = useState(0);
  const [activeItem, setActiveItem] = useState(0);
  const getPosition = () => {
    switch (direction) {
      case 'lr':
        return {
          right: `calc(${width} * ${activeItem})`,
        };

      case 'rl':
        return {
          left: `calc(${width} * ${activeItem})`,
        };

      case 'tb':
        return {
          bottom: `calc(${height} * ${activeItem})`,
        };

      case 'bt':
        return {
          top: `calc(${height} * ${activeItem})`,
        };

      default:
        return {};
    }
  };
  const renderComp = ({ children }) => {
    const _children = flattenChildren(children);
    setChildrenCount(_children.length);
    return (
      <div
        className="w-full h-full flex flex-col gap-2 justify-center items-center"
        key="component"
      >
        <div className="flex overflow-hidden" style={{ width, height }}>
          <motion.div
            layoutId="slider"
            className={classNames('flex relative', {
              'flex-row': direction === 'lr',
              'flex-row-reverse': direction === 'rl',
              'flex-col': direction === 'tb',
              'flex-col-reverse': direction === 'bt',
            })}
            style={{
              width,
              height,
            }}
            animate={getPosition()}
            initial={getPosition()}
            transition={{
              ease: 'easeIn',
              duration,
            }}
          >
            {_children.map((child) => {
              if (!child) return null;
              return (
                <div
                  key={child.key}
                  style={{
                    minWidth: width,
                    minHeight: height,
                    maxWidth: width,
                    maxHeight: height,
                    width,
                    height,
                  }}
                  className="flex flex-col"
                >
                  {child}
                </div>
              );
            })}
          </motion.div>
        </div>

        {/* <div className="flex gap-2"> */}
        {/*   {Array.from(Array(_children.length).keys()).map((item, index) => { */}
        {/*     return ( */}
        {/*       <div */}
        {/*         onClick={() => setActiveItem(index)} */}
        {/*         onDragCapture={console.log} */}
        {/*         key={item} */}
        {/*         draggable */}
        {/*         className={classNames( */}
        {/*           'flex justify-center items-center border font-bold cursor-pointer aspect-square w-6 rounded-full text-xs text-black active:text-primary-900 select-none', */}
        {/*           { */}
        {/*             'bg-primary-200': index === activeItem, */}
        {/*             'bg-neutral-200': index !== activeItem, */}
        {/*           } */}
        {/*         )} */}
        {/*       > */}
        {/*         {item + 1} */}
        {/*       </div> */}
        {/*     ); */}
        {/*   })} */}
        {/* </div> */}
      </div>
    );
  };

  const setActive = (fn) => {
    if (typeof fn === 'function') {
      setActiveItem((s) => {
        const res = fn(s);
        if (res > childrenCount - 1) {
          return childrenCount - 1;
        }
        if (res < 0) {
          return 0;
        }
        return res;
      });
    } else if (typeof fn === 'number') {
      console.log(childrenCount, fn);
      if (fn >= childrenCount) {
        console.log(childrenCount - 1, fn);
        setActiveItem(childrenCount - 1);
        return;
      }
      if (fn < 0) {
        setActiveItem(0);
        return;
      }
      setActiveItem(fn);
    }
  };

  return {
    renderComp,
    activeItem,
    setActiveItem: setActive,
  };
};
