import { Avatar } from '~/components/atoms/avatar';
import { titleCase } from '~/components/utils';
import generateColor from './color-generator';

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
