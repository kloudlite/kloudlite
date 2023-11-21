import Pulsable from 'react-pulsable';
import { Button, IconButton } from '~/components/atoms/button';
import { ReactNode } from 'react';
import { generateKey } from '~/components/utils';
import { DotsSix } from '@jengaicons/react';
import Wrapper from '../components/wrapper';
import CommonTools from '../components/common-tools';
import ListGridView from '../components/list-grid-view';
import Grid from '../components/grid';
import {
  ListBody,
  ListItemWithSubtitle,
  ListTitleWithSubtitleAvatar,
} from '../components/console-list-components';
import ConsoleAvatar from '../components/console-avatar';
import List from '../components/list';

function NewArray(size: number) {
  const x = [];
  for (let i = 0; i < size; i += 1) {
    x[i] = i;
  }
  return x;
}

const GridView = ({ itemsCount }: { itemsCount: number }) => {
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3">
      {NewArray(itemsCount).map((id) => {
        return (
          <Grid.Column
            key={id}
            rows={[
              {
                key: generateKey('grid-item1', id),
                render: () => (
                  <ListTitleWithSubtitleAvatar
                    title="Awesome Title"
                    subtitle="Awesome Subtitle"
                    action={<IconButton icon={<DotsSix />} />}
                    avatar={<ConsoleAvatar name="awesomeid" />}
                  />
                ),
              },
              {
                key: generateKey('grid-item2', id),
                render: () => (
                  <div className="flex flex-col gap-md">
                    {/* <ListItem data={path} /> */}
                    <ListBody data="Awesome Data" />
                  </div>
                ),
              },
              {
                key: generateKey('grid-item3', id),
                render: () => (
                  <ListItemWithSubtitle
                    data="Awesome Data"
                    subtitle="Awesome Subtitle"
                  />
                ),
              },
            ]}
          />
        );
      })}
    </Grid.Root>
  );
};

const ListView = ({ itemsCount }: { itemsCount: number }) => {
  return (
    <List.Root>
      {NewArray(itemsCount).map((id) => {
        return (
          <List.Row
            key={id}
            className="!p-3xl"
            columns={[
              {
                key: generateKey('grid-1', id),
                className: 'flex-1',
                render: () => (
                  <ListTitleWithSubtitleAvatar
                    title={`Awesome Title ${id}`}
                    subtitle={`Awesome Subtitle ${id}`}
                    avatar={<ConsoleAvatar name="awesomeid" />}
                  />
                ),
              },
              {
                key: generateKey('grid-2', id),
                className: 'w-[120px] text-start',
                render: () => <ListBody data="awesome data" />,
              },
              {
                key: generateKey('grid-3', id),
                className: 'w-[180px]',
                render: () => (
                  <ListItemWithSubtitle
                    data={`Awesome Data ${id}`}
                    subtitle={`Awesome Subtitle ${id}`}
                  />
                ),
              },
              {
                key: generateKey('grid-4', 'action', id),
                render: () => <IconButton icon={<DotsSix />} />,
              },
            ]}
          />
        );
      })}
    </List.Root>
  );
};

const Resources = ({ itemsCount = 5 }: { itemsCount?: number }) => {
  return (
    <ListGridView
      listView={<ListView itemsCount={itemsCount} />}
      gridView={<GridView itemsCount={itemsCount} />}
    />
  );
};

export const MainLayoutSK = ({
  title,
  filterOptionsCount = 1,
  noHeader = false,
  itemsCount = 4,
}: {
  title: ReactNode;
  filterOptionsCount?: number;
  noHeader?: boolean;
  itemsCount?: number;
}) => {
  return (
    <Pulsable isLoading>
      <Wrapper
        {...(noHeader
          ? {}
          : {
              header: {
                title,
                action: (
                  <div className="pulsable">
                    <Button content="Awesome Button" />
                  </div>
                ),
              },
            })}
        tools={
          <CommonTools
            options={Array(filterOptionsCount)
              .fill(0)
              .map((_, i) => ({
                name: `Cluster${i}`,
                type: 'text',
                search: true,
                dataFetcher: async () => {
                  return [
                    {
                      content: 'test',
                      value: 'test',
                    },
                  ];
                },
              }))}
          />
        }
      >
        <Resources itemsCount={itemsCount} />
      </Wrapper>
    </Pulsable>
  );
};
