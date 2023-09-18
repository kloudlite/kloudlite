import { ReactNode } from 'react';
import { IButton } from '~/components/atoms/button';
import { EmptyState } from './empty-state';

export interface INoResultsFound {
  title: ReactNode;
  subtitle?: ReactNode;
  action?: IButton;
  image?: ReactNode;
  shadow?: boolean;
  border?: boolean;
  compact?: boolean;
}
const NoResultsFound = ({
  title,
  subtitle,
  action,
  image,
  shadow = true,
  border = true,
  compact = false,
}: INoResultsFound) => {
  return (
    <EmptyState
      compact={compact}
      shadow={shadow}
      border={border}
      heading={title}
      image={image}
      action={action}
    >
      {subtitle}
    </EmptyState>
  );
};

export default NoResultsFound;
