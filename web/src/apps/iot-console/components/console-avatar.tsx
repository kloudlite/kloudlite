import { Avatar } from '@kloudlite/design-system/atoms/avatar';
import { titleCase } from '@kloudlite/design-system/utils';
import generateColor from '~/root/lib/utils/color-generator';

const ConsoleAvatar = ({
  name,
  color,
  size,
}: {
  name: string;
  color?: string;
  size?: string;
}) => {
  return (
    <Avatar
      color={color || generateColor(name, 'dark')}
      size={size || 'sm'}
      image={<span style={{ color: 'white' }}>{titleCase(name[0])}</span>}
    />
  );
};

export default ConsoleAvatar;
