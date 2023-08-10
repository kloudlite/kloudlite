import { useGridList, useGridListItem } from '@react-aria/gridlist';
import { useFocusRing } from '@react-aria/focus';
import { cloneElement, useRef } from 'react';
import { Item, useListState } from 'react-stately';
import { cn } from '~/components/utils';

const List = (props) => {
  const { mode } = props;
  const state = useListState(props);
  const ref = useRef();
  const { gridProps } = useGridList(props, state, ref);
  return (
    <ul
      {...gridProps}
      ref={ref}
      className={cn('flex rounded', {
        'flex-row flex-wrap gap-6xl': mode === 'grid',
        'shadow-button border border-border-default flex-col ': mode === 'list',
      })}
      aria-label="list"
    >
      {[...state.collection].map((item) => (
        <ListItem key={item.key} item={item} state={state} {...props} />
      ))}
    </ul>
  );
};

const ListItem = ({ item, state, ...props }) => {
  const ref = useRef(null);
  const { mode, linkComponent: LinkComponent, prefetchLink } = props;
  const { rowProps, gridCellProps } = useGridListItem(
    { node: item },
    state,
    ref
  );

  const { isFocusVisible, focusProps } = useFocusRing();

  const comp = () => (
    <div
      {...gridCellProps}
      className={cn('cursor-pointer flex p-2xl gap-3xl', {
        'flex-col': mode === 'grid',
        'flex-row items-center justify-between ': mode === 'list',
      })}
    >
      {cloneElement(item.rendered, { mode })}
    </div>
  );

  return (
    <li
      {...rowProps}
      {...focusProps}
      ref={ref}
      className={cn(
        'outline-none ring-offset-1 relative bg-surface-basic-default hover:bg-surface-basic-hovered',
        {
          'focus-visible:ring-2 focus:ring-border-focus z-10 ring-offset-0':
            isFocusVisible,
          'border border-border-default rounded w-92 shadow-button':
            mode === 'grid',
          'border-b border-border-default first:rounded-t last:rounded-b':
            mode === 'list',
        }
      )}
    >
      {LinkComponent && (
        <LinkComponent
          to={item?.props?.to}
          prefetch={prefetchLink ? 'intent' : 'none'}
        >
          {comp()}
        </LinkComponent>
      )}
      {!LinkComponent && comp()}
    </li>
  );
};

export default function ResourceList({
  mode = 'list',
  linkComponent = null,
  prefetchLink = true,
  children,
}) {
  return (
    <List
      selectionMode="none"
      selectionBehavior="toggle"
      onAction={(key) => {
        console.log('item clicked', key);
      }}
      mode={mode}
      linkComponent={linkComponent}
      prefetchLink={prefetchLink}
      aria-label="list"
    >
      {children}
    </List>
  );
}

ResourceList.ResourceItem = Item;
