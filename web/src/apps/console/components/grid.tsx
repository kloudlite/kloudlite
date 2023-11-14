import * as RovingFocusGroup from '@radix-ui/react-roving-focus';
import React, {
  Key,
  KeyboardEvent,
  ReactElement,
  ReactNode,
  useRef,
} from 'react';
import { cn } from '~/components/utils';

const focusableElement = 'a[href], button, input, select, textarea';

const handleKeyNavigation = (
  e: KeyboardEvent<HTMLDivElement>,
  current: any
) => {
  const target = e.target as any;
  try {
    if (
      ['ArrowUp', 'ArrowDown'].includes(e.key) &&
      !target.className.includes('resource-list-item')
    ) {
      let siblings = target
        ?.closest('.resource-list-item')
        ?.querySelectorAll(focusableElement);
      if (siblings) {
        siblings = Array.from(siblings);
        const currentIndex = siblings.indexOf(e.target);
        // @ts-ignore
        document.activeElement.tabIndex = -1;
        if (e.key === 'ArrowDown') {
          if (currentIndex < siblings.length - 1) {
            siblings[currentIndex + 1].tabIndex = 0;
            siblings[currentIndex + 1]?.focus();
          } else {
            siblings[0].tabIndex = 0;
            siblings[0]?.focus();
          }
        } else if (e.key === 'ArrowUp') {
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
      ['ArrowRight', 'ArrowLeft'].includes(e.key) &&
      !target.className.includes('resource-list-item')
    ) {
      // @ts-ignore
      document.activeElement.tabIndex = -1;
      if (e.key === 'ArrowRight') {
        if (target.closest('.resource-list-item')?.nextSibling) {
          target.closest('.resource-list-item')?.nextSibling?.focus();
        } else {
          current?.firstElementChild.focus();
        }
      } else if (target.closest('.resource-list-item')?.previousSibling) {
        target.closest('.resource-list-item')?.previousSibling?.focus();
      } else {
        current?.lastElementChild.focus();
      }
    }

    if (
      ['ArrowUp', 'ArrowDown'].includes(e.key) &&
      target?.className.includes('resource-list-item')
    ) {
      let childs = target?.querySelectorAll(focusableElement);
      if (childs) {
        childs = Array.from(childs);
        if (childs.length < 1) return;
        if (e.key === 'ArrowDown') {
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

interface IRow {
  render?: () => ReactNode;
  key: Key;
  className?: string;
  width?: string;
  label?: ReactNode;
}

interface IMain {
  rows: IRow[];
  className?: string;
  onClick?: ((item?: IRow[]) => void) | null;
  pressed?: boolean;
  to?: string;
}

interface IColumnBase extends IMain {
  linkComponent?: any;
}

const ColumnBase = ({
  rows = [],
  to = '',
  linkComponent = 'div',
  className = '',
  onClick = null,
  pressed = false,
}: IColumnBase) => {
  let Component: any = linkComponent;

  if (to) {
    if (linkComponent === 'div') {
      Component = 'a';
    } else {
      Component = linkComponent;
    }
  }

  return (
    <RovingFocusGroup.Item
      role="row"
      asChild
      className={cn(
        'resource-list-item focus-visible:shadow-focus focus:z-10 outline-none ring-offset-1 relative p-2xl flex flex-col gap-3xl rounded border border-border-default bg-surface-basic-default shadow-button',
        className,
        {
          'bg-surface-basic-default': !pressed,
          'cursor-pointer hover:bg-surface-basic-hovered':
            (!!onClick || linkComponent !== 'div') && !pressed,
          'bg-surface-basic-active': pressed,
        }
      )}
      onClick={() => {
        if (onClick) onClick(rows);
      }}
      onKeyDown={(e) => {
        if (['Enter', ' '].includes(e.key) && onClick) {
          onClick(rows);
        }
      }}
    >
      <Component {...(Component === 'a' ? { href: to } : { to })}>
        {rows.map((item) => (
          <div key={item?.key} className={cn('', item?.className, item?.width)}>
            {item?.render ? item.render() : item?.label}
          </div>
        ))}
      </Component>
    </RovingFocusGroup.Item>
  );
};

type IColumn = IMain;

const Column = ({
  rows = [],
  className = '',
  onClick,
  pressed = false,
  to = '',
}: IColumn) => {
  return (
    <ColumnBase
      rows={rows}
      className={className}
      onClick={onClick}
      pressed={pressed}
      to={to}
    />
  );
};

interface IRoot {
  children: ReactNode;
  className?: string;
  linkComponent?: any;
}

const Root = ({ children, className = '', linkComponent }: IRoot) => {
  const ref = useRef<HTMLDivElement>(null);
  return (
    <RovingFocusGroup.Root
      ref={ref}
      className={cn('grid grid-cols-4 gap-6xl', className)}
      asChild
      loop
      orientation="horizontal"
      onFocus={(e) => {
        try {
          if (e.target.className.includes('resource-list-item')) {
            if (e.target.className.includes('resource-list-item')) {
              e.target.querySelectorAll(focusableElement).forEach((el) => {
                (el as HTMLButtonElement).tabIndex = -1;
              });
            }
          }
        } catch {
          console.log('Error Focusing');
        }
      }}
      onKeyDown={(e) => {
        handleKeyNavigation(e, ref.current);
      }}
    >
      <div role="list" aria-label="list">
        {React.Children.map(children as ReactElement[], (child) => (
          <ColumnBase {...child.props} linkComponent={linkComponent} />
        ))}
      </div>
    </RovingFocusGroup.Root>
  );
};

const Grid = {
  Root,
  Column,
};

export default Grid;
