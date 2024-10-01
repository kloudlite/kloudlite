import { Avatar } from '~/components/atoms/avatar';
import { titleCase } from '~/components/utils';
import generateColor from '~/root/lib/utils/color-generator';

const TemplateAvatar = ({
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
      image={<span style={{ color: 'black' }}>{titleCase(name)}</span>}
      isTemplate
    />
  );
};

export default TemplateAvatar;
