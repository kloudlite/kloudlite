import { useGridList, useGridListItem } from '@react-aria/gridlist';
import { useFocusRing } from '@react-aria/focus';
import { useRef } from 'react';
import { Item, useListState } from 'react-stately';
import { cn } from '~/components/utils';

const List = (props) => {
  const state = useListState(props);
  const ref = useRef();
  const { gridProps } = useGridList(props, state, ref);
  return (
    <ul {...gridProps} ref={ref} aria-label="list">
      {[...state.collection].map((item) => (
        <ListItem key={item.key} item={item} state={state} />
      ))}
    </ul>
  );
};

const ListItem = ({ item, state }) => {
  const ref = useRef(null);
  const { rowProps, gridCellProps } = useGridListItem(
    { node: item },
    state,
    ref
  );

  const { isFocusVisible, focusProps } = useFocusRing();
  return (
    <div className="relative bg-surface-basic-default hover:bg-surface-basic-hovered p-sm focus-within:z-10">
      <li
        {...rowProps}
        {...focusProps}
        ref={ref}
        className={cn('outline-none ring-offset-0 group', {
          'focus-visible:ring-2 focus:ring-border-focus': isFocusVisible,
        })}
      >
        <div {...gridCellProps} className={cn('cursor-pointer -m-sm')}>
          {item.rendered}
        </div>
      </li>
    </div>
  );
};

export default function ResourceList({ children }) {
  return (
    <List
      selectionMode="none"
      selectionBehavior="toggle"
      onAction={(key) => {
        console.log('item clicked', key);
      }}
      aria-label="list"
    >
      {children}
    </List>
  );
}

ResourceList.ResourceItem = Item;
