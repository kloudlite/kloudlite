import * as RovingFocusGroup from '@radix-ui/react-roving-focus';
import { useRef } from 'react';
import { cn } from '~/components/utils';

const focusableElement = 'a[href], button, input, select, textarea';

const handleKeyNavigation = (e, ref) => {
  try {
    if (
      ['ArrowLeft', 'ArrowRight'].includes(e.key) &&
      !e.target?.className.includes('resource-list-item')
    ) {
      let siblings = e.target
        ?.closest('.resource-list-item')
        ?.querySelectorAll(focusableElement);
      if (siblings) {
        siblings = Array.from(siblings);
        const currentIndex = siblings.indexOf(e.target);
        // @ts-ignore
        document.activeElement.tabIndex = -1;
        if (e.key === 'ArrowRight') {
          if (currentIndex < siblings.length - 1) {
            siblings[currentIndex + 1].tabIndex = 0;
            siblings[currentIndex + 1]?.focus();
          } else {
            siblings[0].tabIndex = 0;
            siblings[0]?.focus();
          }
        } else if (e.key === 'ArrowLeft') {
          if (currentIndex > 0) {
            siblings[currentIndex - 1].tabIndex = 0;
            siblings[currentIndex - 1]?.focus();
          } else {
            siblings[siblings.length - 1].tabIndex = 0;
            siblings[siblings.length - 1]?.focus();
          }
        }
      }
    }
    if (
      ['ArrowDown', 'ArrowUp'].includes(e.key) &&
      !e.target?.className.includes('resource-list-item')
    ) {
      // @ts-ignore
      document.activeElement.tabIndex = -1;
      if (e.key === 'ArrowDown') {
        if (e.target.closest('.resource-list-item')?.nextSibling) {
          e.target.closest('.resource-list-item')?.nextSibling?.focus();
        } else {
          ref.current?.firstElementChild.focus();
        }
      } else if (e.target.closest('.resource-list-item')?.previousSibling) {
        e.target.closest('.resource-list-item')?.previousSibling?.focus();
      } else {
        ref.current?.lastElementChild.focus();
      }
    }

    if (
      ['ArrowLeft', 'ArrowRight'].includes(e.key) &&
      e.target?.className.includes('resource-list-item')
    ) {
      let childs = e.target?.querySelectorAll(focusableElement);
      if (childs) {
        childs = Array.from(childs);
        if (childs.length < 1) return;
        if (e.key === 'ArrowRight') {
          childs[0]?.focus();
        } else {
          childs[childs.length - 1]?.focus();
        }
      }
    }
  } catch {
    console.log('Error focusing');
  }
};

const Item = ({ items, className = '', onClick = null, pressed = false }) => {
  return (
    <RovingFocusGroup.Item
      role="row"
      asChild
      className={cn(
        'resource-list-item focus-visible:ring-2 focus:ring-border-focus focus:z-10 outline-none ring-offset-1 relative [&:not(:last-child)]:border-b border-border-default first:rounded-t last:rounded-b flex flex-row items-center p-2xl gap-3xl',
        className,
        {
          'bg-surface-basic-default': !pressed,
          'cursor-pointer hover:bg-surface-basic-hovered': onClick && !pressed,
          'bg-surface-basic-active': pressed,
        }
      )}
      onClick={() => {
        if (onClick) onClick(items);
      }}
    >
      <div>
        {items.map((item) => (
          <div key={item?.key} className={cn('', item?.className, item?.width)}>
            {item?.render ? item.render() : item?.label}
          </div>
        ))}
      </div>
    </RovingFocusGroup.Item>
  );
};

const Root = ({ children, className = '' }) => {
  const ref = useRef(null);
  return (
    <RovingFocusGroup.Root
      ref={ref}
      className={cn(
        'rounded border border-border-default shadow-button',
        className
      )}
      asChild
      loop
      orientation="vertical"
      onFocus={(e) => {
        try {
          if (e.target.className.includes('resource-list-item')) {
            if (e.target.className.includes('resource-list-item')) {
              e.target.querySelectorAll(focusableElement).forEach((el) => {
                // @ts-ignore
                el.tabIndex = -1;
              });
            }
          }
        } catch {
          console.log('Error Focusing');
        }
      }}
      onKeyDown={(e) => {
        handleKeyNavigation(e, ref);
      }}
    >
      <div role="list" aria-label="list">
        {children}
      </div>
    </RovingFocusGroup.Root>
  );
};

const List = {
  Root,
  Item,
};

export default List;
