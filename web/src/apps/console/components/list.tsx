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
      ['ArrowLeft', 'ArrowRight'].includes(e.key) &&
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
      !target.className.includes('resource-list-item')
    ) {
      // @ts-ignore
      document.activeElement.tabIndex = -1;
      if (e.key === 'ArrowDown') {
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
      ['ArrowLeft', 'ArrowRight'].includes(e.key) &&
      target?.className.includes('resource-list-item')
    ) {
      let childs = target?.querySelectorAll(focusableElement);
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

interface IColumn {
  render?: () => ReactNode;
  key: Key;
  className?: string;
  width?: string;
  label?: ReactNode;
}

interface IMain {
  columns: IColumn[];
  className?: string;
  onClick?: ((item?: IColumn[]) => void) | null;
  pressed?: boolean;
  to?: string;
}

interface IRowBase extends IMain {
  linkComponent?: any;
}

const RowBase = ({
  columns = [],
  to = '',
  linkComponent = 'div',
  className = '',
  onClick = null,
  pressed = false,
}: IRowBase) => {
  let Component: any = linkComponent;

  if (to) {
    if (linkComponent === 'div') {
      Component = 'a';
    } else {
      Component = linkComponent;
    }
  }

  const css = cn(
    'overflow-hidden resource-list-item focus-visible:ring-2 focus:ring-border-focus focus:z-10 outline-none ring-offset-1 relative [&:not(:last-child)]:border-b border-border-default first:rounded-t last:rounded-b flex flex-row items-center p-2xl gap-3xl',
    className,
    {
      'bg-surface-basic-default': !pressed,
      'cursor-pointer hover:bg-surface-basic-hovered':
        (!!onClick || linkComponent !== 'div') && !pressed,
      'bg-surface-basic-active': pressed,
    }
  );

  if (!!onClick || linkComponent !== 'div') {
    return (
      <RovingFocusGroup.Item
        role="row"
        asChild
        className={css}
        onClick={() => {
          if (onClick) onClick(columns);
        }}
        onKeyDown={(e) => {
          if (['Enter', ' '].includes(e.key) && onClick) {
            onClick(columns);
          }
        }}
      >
        <Component {...(Component === 'a' ? { href: to } : { to })}>
          {columns.map((item) => (
            <div
              key={item?.key}
              className={cn('', item?.className, item?.width)}
            >
              {item?.render ? item.render() : item?.label}
            </div>
          ))}
        </Component>
      </RovingFocusGroup.Item>
    );
  }
  return (
    <div className={css} role="row">
      {columns.map((item) => (
        <div key={item?.key} className={cn('', item?.className, item?.width)}>
          {item?.render ? item.render() : item?.label}
        </div>
      ))}
    </div>
  );
};

type IRow = IMain;

const Row = ({
  columns = [],
  className = '',
  onClick,
  pressed = false,
  to = '',
}: IRow) => {
  return (
    <RowBase
      columns={columns}
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
  header?: ReactNode;
}

const Root = ({ children, header, className = '', linkComponent }: IRoot) => {
  const ref = useRef<HTMLDivElement>(null);
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
        {header && (
          <div
            aria-label="list-header"
            className="px-xl py-lg gap-lg bg-surface-basic-subdued rounded-t"
          >
            {header}
          </div>
        )}
        {React.Children.map(children as ReactElement[], (child) => (
          <RowBase {...child.props} linkComponent={linkComponent} />
        ))}
      </div>
    </RovingFocusGroup.Root>
  );
};

const List = {
  Root,
  Row,
};

export default List;
