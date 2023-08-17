import { Search } from '@jengaicons/react';
import * as RovingFocusGroup from '@radix-ui/react-roving-focus';
import { IconButton } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { cn } from '~/components/utils';

const List = ({ mode = 'list' }) => {
  return (
    <RovingFocusGroup.Root asChild>
      <div role="list" aria-label="list">
        <RovingFocusGroup.Item asChild onClick={(e) => console.log(e)}>
          <a
            role="row"
            className={cn(
              'focus-visible:ring-2 focus:ring-border-focus z-10 outline-none ring-offset-1 relative bg-surface-basic-default hover:bg-surface-basic-hovered',
              {
                'border-b border-border-default first:rounded-t last:rounded-b flex p-2xl gap-3xl':
                  mode === 'list',
              }
            )}
          >
            hello world
            <RovingFocusGroup.Root>
              <RovingFocusGroup.Item asChild>
                <TextInput />
              </RovingFocusGroup.Item>
              <RovingFocusGroup.Item asChild>
                <IconButton icon={Search} />
              </RovingFocusGroup.Item>
            </RovingFocusGroup.Root>
          </a>
        </RovingFocusGroup.Item>
        <RovingFocusGroup.Item
          asChild
          onClick={(e) => console.log(e)}
          onFocus={(e) => {
            if (e.target.firstElementChild?.tabIndex) {
              e.target.firstElementChild.tabIndex = 0;
            }
          }}
        >
          <a
            href="https://google.com"
            role="row"
            className={cn(
              'focus-visible:ring-2 focus:ring-border-focus z-10 outline-none ring-offset-1 relative bg-surface-basic-default hover:bg-surface-basic-hovered',
              {
                'border-b border-border-default first:rounded-t last:rounded-b flex p-2xl gap-3xl':
                  mode === 'list',
              }
            )}
          >
            hello world
            <RovingFocusGroup.Root
              onBlur={(e) => {
                e.target.tabIndex = -1;
              }}
            >
              <RovingFocusGroup.Item asChild>
                <TextInput />
              </RovingFocusGroup.Item>
              <RovingFocusGroup.Item
                asChild
                onBlur={(e) => {
                  e.target.tabIndex = -1;
                }}
              >
                <IconButton icon={Search} />
              </RovingFocusGroup.Item>
            </RovingFocusGroup.Root>
          </a>
        </RovingFocusGroup.Item>
      </div>
    </RovingFocusGroup.Root>
  );
};

export default List;
