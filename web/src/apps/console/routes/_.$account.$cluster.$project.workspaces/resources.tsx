import { DotsThreeVerticalFill } from '@jengaicons/react';
import { Link, useParams } from '@remix-run/react';
import { IconButton } from '~/components/atoms/button';
import { Thumbnail } from '~/components/atoms/thumbnail';
import { dayjs } from '~/components/molecule/dayjs';
import { titleCase } from '~/components/utils';
import {
  ListBody,
  ListItemWithSubtitle,
  ListTitleWithSubtitleAvatar,
} from '~/console/components/console-list-components';
import List from '~/console/components/list';
import { IWorkspace } from '~/console/server/gql/queries/workspace-queries';
import { parseFromAnn, parseName } from '~/console/server/r-utils/common';
import { keyconstants } from '~/console/server/r-utils/key-constants';

const Resources = ({ items = [] }: { items: IWorkspace[] }) => {
  const { account, project } = useParams();

  return (
    <List.Root linkComponent={Link}>
      {items.map((item) => {
        const { name, id, cluster, path, updateInfo } = {
          name: item.displayName,
          id: parseName(item),
          cluster: item.clusterName,
          path: `/workspaces/${parseName(item)}`,
          updateInfo: {
            author: titleCase(
              `${parseFromAnn(item, keyconstants.author)} updated the workspace`
            ),
            time: dayjs(item.updateTime).fromNow(),
          },
        };

        return (
          <List.Row
            key={id}
            className="!p-3xl"
            to={`/${account}/${cluster}/${project}/workspace/${id}`}
            columns={[
              {
                key: 1,
                className: 'flex-1',
                render: () => (
                  <ListTitleWithSubtitleAvatar
                    title={name}
                    subtitle={id}
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
                className: 'w-[230px] text-start',
                render: () => <ListBody data={path} />,
              },
              {
                key: 3,
                className: 'w-[120px] text-start',
                render: () => <ListBody data={item.clusterName} />,
              },
              {
                key: 4,
                render: () => (
                  <ListItemWithSubtitle
                    data={updateInfo.author}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              {
                key: 5,
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
