import { useGridList, useGridListItem } from '@react-aria/gridlist';
import { useFocusRing } from '@react-aria/focus';
import { useRef } from 'react';
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
        'flex-row flex-wrap gap-6xl ': mode === 'grid',
        'shadow-base border-border-default flex-col': mode === 'list',
      })}
    >
      {[...state.collection].map((item) => (
        <ListItem key={item.key} item={item} state={state} mode={mode} />
      ))}
    </ul>
  );
};

const ListItem = ({ item, state, mode }) => {
  const ref = useRef(null);
  const { rowProps, gridCellProps } = useGridListItem(
    { node: item },
    state,
    ref
  );

  const { isFocusVisible, focusProps } = useFocusRing();
  return (
    <li
      {...rowProps}
      {...focusProps}
      ref={ref}
      className={cn(
        'outline-none ring-offset-1 relative bg-surface-basic-default hover:bg-surface-basic-hovered',
        {
          'focus-visible:ring-2 focus:ring-border-focus z-10 ring-offset-0 border-surface-default':
            isFocusVisible,
          'border border-border-default rounded w-92 shadow-base':
            mode === 'grid',
          'border-b border-border-disabled first:rounded-t last:rounded-b':
            mode === 'list',
        }
      )}
    >
      <div
        {...gridCellProps}
        className={cn('cursor-pointer flex p-3xl gap-3xl', {
          'flex-col': mode === 'grid',
          'flex-row items-center justify-between ': mode === 'list',
        })}
      >
        {item.rendered}
      </div>
    </li>
  );
};

export default function ResourceList({ mode = 'list', children }) {
  return (
    <List
      selectionMode="none"
      selectionBehavior="toggle"
      onAction={(key) => {
        console.log('item clicked', key);
      }}
      mode={mode}
    >
      {children}
    </List>
  );
}

ResourceList.ResourceItem = Item;
