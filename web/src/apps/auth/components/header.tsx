import { Button } from '~/components/atoms/button';
import { BrandLogo } from '~/components/branding/brand-logo';
import { Link } from '@remix-run/react';
import { ReactNode } from 'react';
import Wrapper from './wrapper';

const Header = ({ headerExtra }: { headerExtra?: ReactNode }) => {
  return (
    <div className="sticky top-0 bg-surface-basic-subdued w-full border-b border-border-default">
      <Wrapper className="min-h-[68px] max-h-[68px] flex flex-row items-center justify-between">
        <a href="/" aria-label="kloudlite">
          <BrandLogo size={24} detailed />
        </a>
        <div className="flex flex-row gap-xl items-center">
          <Button
            variant="plain"
            content="Contact us"
            linkComponent={Link}
            to="https://kloudlite.io/contact-us"
          />
          {headerExtra}
        </div>
      </Wrapper>
    </div>
  );
};

export default Header;
