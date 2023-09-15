import { SmileySad } from '@jengaicons/react';
import { ReactNode } from 'react';
import { IButton } from '~/components/atoms/button';
import { EmptyState } from './empty-state';

export interface INoResultsFound {
  title: string;
  subtitle?: string;
  action?: IButton;
  image?: ReactNode;
}
const NoResultsFound = ({
  title,
  subtitle,
  action,
  image,
}: INoResultsFound) => {
  return (
    <EmptyState
      heading={title}
      image={
        image || (
          <span className="text-icon-default">
            <SmileySad size={48} color="currentColor" />
          </span>
        )
      }
      action={action}
    >
      {subtitle}
    </EmptyState>
  );
};

export default NoResultsFound;
