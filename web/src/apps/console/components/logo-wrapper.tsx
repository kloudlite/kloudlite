import { Link } from '@remix-run/react';
import { ReactNode } from 'react';

const LogoWrapper = ({
  children,
  to,
}: {
  children?: ReactNode;
  to: string;
}) => {
  return (
    <Link
      className="rounded outline-none ring-offset-1 focus-visible:ring-2 focus:ring-border-focus"
      to={to}
      prefetch="intent"
    >
      {children}
    </Link>
  );
};

export default LogoWrapper;
