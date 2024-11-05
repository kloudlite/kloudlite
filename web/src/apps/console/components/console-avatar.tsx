import { Avatar } from '@kloudlite/design-system/atoms/avatar';
import { titleCase } from '@kloudlite/design-system/utils';
import generateColor from '~/root/lib/utils/color-generator';
import { EnvIconComponent } from './icons';

const ConsoleAvatar = ({
  name,
  color,
  size,
  isAvatar = false,
  icon,
  className,
}: {
  name: string;
  color?: string;
  size?: string;
  isAvatar?: boolean;
  icon?: React.ReactNode;
  className?: string;
}) => {
  return (
    <Avatar
      color={color || generateColor(name, 'dark')}
      size={size || 'sm'}
      // image={<span style={{ color: 'white' }}>{titleCase(name[0])}</span>}
      image={
        isAvatar ? (
          icon || <EnvIconComponent size={20} />
        ) : (
          // <EnvIconComponent size={20} />
          <span style={{ color: 'white' }}>{titleCase(name[0])}</span>
        )
      }
      className={className}
    />
  );
};

export default ConsoleAvatar;
