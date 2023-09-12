import { useFocusRing } from '@react-aria/focus';
import { useGridList, useGridListItem } from '@react-aria/gridlist';
import { ReactElement, cloneElement, useRef } from 'react';
import {
  Item,
  ItemProps,
  ListProps,
  ListState,
  useListState,
} from 'react-stately';
import { cn } from '~/components/utils';

type ListModes = 'list' | 'grid' | (string & NonNullable<unknown>);
interface IList extends ListProps<object> {
  mode: ListModes;
  onAction: (key: any) => void;
  linkComponent: any;
  prefetchLink: boolean;
}

interface IListItem extends IList {
  item: any;
  state: ListState<unknown>;
}
function ListItem({ item, state, ...props }: IListItem) {
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
}

function List(props: IList) {
  const { mode } = props;
  const state = useListState(props);
  const ref = useRef<HTMLUListElement>(null);
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
}

interface IResourceList {
  mode: ListModes;
  linkComponent?: any;
  prefetchLink?: boolean;
  children: ReactElement[];
}
export default function ResourceList({
  mode = 'list',
  linkComponent = null,
  prefetchLink = true,
  children,
}: IResourceList) {
  return (
    <List
      selectionMode="none"
      selectionBehavior="toggle"
      onAction={(key: any) => {
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

interface extra {
  to?: string;
}

type IResourceItem = <T>(props: ItemProps<T> & extra) => JSX.Element;
const ri: IResourceItem = Item;

ResourceList.ResourceItem = ri;
