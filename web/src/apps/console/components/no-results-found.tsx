import { ReactNode } from 'react';
import { IButton } from '@kloudlite/design-system/atoms/button';
import { EmptyState } from './empty-state';

export interface INoResultsFound {
  title: ReactNode;
  subtitle?: ReactNode;
  action?: IButton;
  image?: ReactNode;
  shadow?: boolean;
  border?: boolean;
  compact?: boolean;
  padding?: boolean;
}
const NoResultsFound = ({
  title,
  subtitle,
  action,
  image,
  shadow = true,
  border = true,
  compact = false,
  padding = true,
}: INoResultsFound) => {
  return (
    <EmptyState
      compact={compact}
      shadow={shadow}
      border={border}
      heading={title}
      image={image}
      action={action}
      padding={padding}
    >
      {subtitle}
    </EmptyState>
  );
};

export default NoResultsFound;
