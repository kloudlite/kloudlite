import { Avatar } from '~/components/atoms/avatar';
import { titleCase } from '~/components/utils';
import generateColor from './color-generator';

const ConsoleAvatar = ({ name }: { name: string }) => {
  return (
    <Avatar
      color={generateColor(name, 'dark')}
      size="sm"
      image={<span style={{ color: 'white' }}>{titleCase(name[0])}</span>}
    />
  );
};

export default ConsoleAvatar;
