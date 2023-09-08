import { DotsThreeVerticalFill } from '@jengaicons/react';
import { IconButton } from '~/components/atoms/button';
import { Thumbnail } from '~/components/atoms/thumbnail';
import List from '~/console/components/list';
import {
  ListBody,
  ListItemWithSubtitle,
  ListTitleWithSubtitleAvatar,
} from '~/console/components/console-list-components';
import { titleCase } from '~/components/utils';
import { Link, useParams } from '@remix-run/react';

const Resources = ({
  items = [],
}: {
  items: {
    name: string;
    displayName: string;
    providerRegion: string;
    updateInfo: { author: string; time: string };
  }[];
}) => {
  const { account } = useParams();
  return (
    <List.Root linkComponent={Link}>
      {items.map((item) => {
        return (
          <List.Row
            key={item.name}
            className="!p-3xl"
            to={`/${account}/${item.name}/nodepools`}
            columns={[
              {
                key: 1,
                className: 'flex-1',
                render: () => (
                  <ListTitleWithSubtitleAvatar
                    title={item.displayName}
                    subtitle={item.name}
                    avatar={
                      <Thumbnail
                        size="sm"
                        rounded
                        src="https://images.unsplash.com/photo-1600716051809-e997e11a5d52?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxzZWFyY2h8NHx8c2FtcGxlfGVufDB8fDB8fA%3D%3D&auto=format&fit=crop&w=800&q=60"
                      />
                    }
                  />
                ),
              },
              {
                key: 2,
                className: 'w-[120px] text-start',
                render: () => <ListBody data={item.providerRegion} />,
              },
              {
                key: 3,
                render: () => (
                  <ListItemWithSubtitle
                    data={titleCase(item.updateInfo.author)}
                    subtitle={item.updateInfo.time}
                  />
                ),
              },
              {
                key: 4,
                render: () => (
                  <IconButton
                    icon={<DotsThreeVerticalFill />}
                    variant="plain"
                    onClick={(e) => e.stopPropagation()}
                  />
                ),
              },
            ]}
          />
        );
      })}
    </List.Root>
  );
};

export default Resources;
