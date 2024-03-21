import * as RovingFocusGroup from '@radix-ui/react-roving-focus';
import { KeyboardEvent, ReactNode, useRef } from 'react';
import { cn } from '~/components/utils';
import logger from '~/root/lib/client/helpers/log';
import { LoadingPlaceHolder } from './loading';

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
    logger.error('Error focusing');
  }
};

type IHeader = {
  className?: string;
  render: () => ReactNode;
  name: string;
};

interface IColumn {
  render: () => ReactNode;
}

interface IMain {
  columns?: Record<string, IColumn>;
  className?: string;
  onClick?: ((item?: Record<string, IColumn>) => void) | null;
  pressed?: boolean;
  to?: string;
  plain?: boolean;
}

interface IRowBase extends IMain {
  linkComponent?: any;
  headers?: IHeader[];
  disabled?: boolean;
  detail?: ReactNode;
  hideDetailSeperator?: boolean;
}

const RowBase = ({
  columns,
  to = '',
  linkComponent = 'div',
  className = '',
  onClick = null,
  pressed = false,
  headers,
  disabled,
  detail,
  hideDetailSeperator,
}: IRowBase) => {
  let Component: any = linkComponent;

  if (to) {
    if (linkComponent === 'div') {
      Component = 'a';
    } else {
      Component = linkComponent;
    }
  } else {
    Component = 'div';
  }

  const commonCss = cn(
    {
      'bg-surface-basic-default': !pressed,
      'cursor-pointer hover:bg-surface-basic-hovered':
        (!!onClick || linkComponent !== 'div') && !pressed && !disabled,
      'bg-surface-basic-pressed': pressed,
      'cursor-default': !!disabled,
    },
    '[&:not(:last-child)]:border-b border-border-default'
  );

  const css = cn(
    'w-full overflow-hidden resource-list-item focus-visible:ring-2 focus:ring-border-focus focus:z-10 outline-none ring-offset-1 relative flex flex-row items-center gap-3xl',
    className
  );

  if (!disabled && !pressed) {
    return (
      <RovingFocusGroup.Item
        role="row"
        asChild
        onClick={() => {
          if (onClick) onClick(columns);
        }}
        onKeyDown={(e) => {
          if (['Enter', ' '].includes(e.key) && onClick) {
            onClick(columns);
          }
        }}
      >
        <Component
          {...(Component === 'a' ? { href: to } : { to })}
          className={cn('flex flex-col last:rounded-b p-2xl', commonCss)}
        >
          <div className={css}>
            {headers?.map((item) => (
              <div key={item.name} className={cn(item.className)}>
                {columns?.[item.name]?.render()}
              </div>
            ))}
          </div>
          <div className={cn('px-2xl')}>{detail}</div>
        </Component>
      </RovingFocusGroup.Item>
    );
  }

  return (
    <div className={cn(css, commonCss, 'p-2xl')}>
      <div role="row">
        {headers?.map((item) => (
          <div key={item.name} className={cn(item.className)}>
            {columns?.[item.name]?.render()}
          </div>
        ))}
      </div>
      {detail}
    </div>
  );
};

type IRow = IMain;

const Row = ({
  columns,
  className = '',
  onClick,
  pressed = false,
  to = '',
  plain = false,
}: IRow) => {
  return (
    <RowBase
      columns={columns}
      className={className}
      onClick={onClick}
      pressed={pressed}
      to={to}
      plain={plain}
    />
  );
};

interface IRoot {
  className?: string;
  linkComponent?: any;
  plain?: boolean;
  loading?: boolean;
  data?: {
    headers: IHeader[];
    rows: Array<{
      columns: Record<string, IColumn>;
      to?: string;
      disabled?: boolean;
      detail?: ReactNode;
      hideDetailSeperator?: boolean;
      onClick?: ((item?: Record<string, IColumn>) => void) | null;
      pressed?: boolean;
    }>;
    className?: Array<string>;
  };
  headerClassName?: string;
}

const Root = ({
  className = '',
  linkComponent,
  plain,
  loading = false,
  data,
  headerClassName,
}: IRoot) => {
  const ref = useRef<HTMLDivElement>(null);

  // logger.log(data);
  return (
    <>
      {!loading && (
        <RovingFocusGroup.Root
          ref={ref}
          className={cn(
            'w-full',
            {
              'rounded border border-border-default shadow-button': !plain,
            },
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
              logger.error('Error Focusing');
            }
          }}
          onKeyDown={(e) => {
            handleKeyNavigation(e, ref.current);
          }}
        >
          <div className="flex flex-col overflow-hidden">
            <div
              className={cn(
                'text-text-strong flex flex-row items-center py-xl px-2xl gap-3xl bodyMd-medium bg-surface-basic-active',
                headerClassName
              )}
            >
              {data?.headers.map((h, index) => (
                <div key={`${index + h.name}`} className={cn(h.className)}>
                  {h.render()}
                </div>
              ))}
            </div>
            <div role="list" aria-label="list" className="w-full">
              {data?.rows.map((r, index) => (
                <RowBase
                  linkComponent={linkComponent}
                  key={`${index + (r.to || '')}`}
                  columns={r.columns}
                  to={r.to}
                  headers={data.headers}
                  disabled={r.disabled}
                  detail={r.detail}
                  hideDetailSeperator={r.hideDetailSeperator}
                  onClick={r.onClick}
                  pressed={r.pressed}
                />
              ))}
            </div>
          </div>
        </RovingFocusGroup.Root>
      )}
      {loading && (
        <div
          className={cn(
            'flex items-center justify-center h-full',
            {
              'rounded border border-border-default shadow-button': !plain,
            },
            className
          )}
        >
          <LoadingPlaceHolder />
        </div>
      )}
    </>
  );
};

const ListV2 = {
  Root,
  Row,
};

export default ListV2;
