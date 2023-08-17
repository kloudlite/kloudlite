import { Search } from '@jengaicons/react';
import * as RovingFocusGroup from '@radix-ui/react-roving-focus';
import { IconButton } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { cn } from '~/components/utils';

const List = ({ mode = 'list' }) => {
  return (
    <RovingFocusGroup.Root
      asChild
      orientation="vertical"
      onKeyDown={(e) => {
        console.log(e);
        if (
          ['ArrowDown', 'ArrowUp'].includes(e.key) &&
          !e.target.className.includes('resource-list-item')
        ) {
          console.log('inner child');
        }
      }}
    >
      <div role="list" aria-label="list">
        <RovingFocusGroup.Item asChild>
          <a
            role="row"
            className={cn(
              'resource-list-item focus-visible:ring-2 focus:ring-border-focus z-10 outline-none ring-offset-1 relative bg-surface-basic-default hover:bg-surface-basic-hovered',
              {
                'border-b border-border-default first:rounded-t last:rounded-b flex p-2xl gap-3xl':
                  mode === 'list',
              }
            )}
          >
            hello world
            <RovingFocusGroup.Root orientation="horizontal">
              <RovingFocusGroup.Item asChild>
                <TextInput />
              </RovingFocusGroup.Item>
              <RovingFocusGroup.Item asChild>
                <IconButton icon={Search} />
              </RovingFocusGroup.Item>
            </RovingFocusGroup.Root>
          </a>
        </RovingFocusGroup.Item>
        <RovingFocusGroup.Item asChild>
          <a
            href="https://google.com"
            role="row"
            className={cn(
              'resource-list-item focus-visible:ring-2 focus:ring-border-focus z-10 outline-none ring-offset-1 relative bg-surface-basic-default hover:bg-surface-basic-hovered',
              {
                'border-b border-border-default first:rounded-t last:rounded-b flex p-2xl gap-3xl':
                  mode === 'list',
              }
            )}
          >
            hello world
            <RovingFocusGroup.Root orientation="horizontal">
              <RovingFocusGroup.Item asChild focusable={false}>
                <TextInput />
              </RovingFocusGroup.Item>
              <RovingFocusGroup.Item asChild focusable={false}>
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
