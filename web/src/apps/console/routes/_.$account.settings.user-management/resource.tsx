import { DotsThreeVerticalFill } from '@jengaicons/react';
import { Avatar } from '~/components/atoms/avatar';
import { IconButton } from '~/components/atoms/button';
import {
  ListBody,
  ListTitleWithSubtitleAvatar,
} from '~/console/components/console-list-components';
import List from '~/console/components/list';

interface IResource {
  items: {
    id: string;
    name: string;
    role: string;
    email: string;
  }[];
}
const Resources = ({ items = [] }: IResource) => {
  return (
    <List.Root>
      {items.map((item) => (
        <List.Row
          key={item.id}
          className="!p-3xl"
          columns={[
            {
              key: 1,
              className: 'flex-1',
              render: () => (
                <ListTitleWithSubtitleAvatar
                  avatar={<Avatar size="sm" />}
                  subtitle={item.email}
                  title={item.name}
                />
              ),
            },
            {
              key: 2,
              render: () => <ListBody data={item.role} />,
            },
            {
              key: 3,
              render: () => (
                <IconButton
                  className="pulsable pulsable-hidden"
                  icon={<DotsThreeVerticalFill />}
                  variant="plain"
                  onClick={(e) => e.stopPropagation()}
                />
              ),
            },
          ]}
        />
      ))}
    </List.Root>
  );
};

export default Resources;
