//@ts-ignore
import { Button } from 'kl-design-system/atoms/button';
//@ts-ignore
import { BrandLogo } from 'kl-design-system/branding/brand-logo';
import { ReactNode } from 'react';
import Wrapper from './wrapper';

const Header = ({ headerExtra }: { headerExtra?: ReactNode }) => {
    return (
        <div className="sticky top-0 bg-surface-basic-subdued w-full border-b border-border-default z-50">
            <Wrapper className="min-h-[68px] max-h-[68px] flex flex-row items-center justify-between">
                <a href="/" aria-label="kloudlite">
                    <div className="hidden md:block">
                        <BrandLogo size={24} detailed />
                    </div>
                    <div className="md:hidden">
                        <BrandLogo size={24} />
                    </div>
                </a>
                <div className="flex flex-row gap-xl items-center">
                    <Button
                        variant="plain"
                        content="Contact us"
                        // linkComponent={Link}
                        to="https://kloudlite.io/contact-us"
                    />
                    {headerExtra}
                </div>
            </Wrapper>
        </div>
    );
};

export default Header;
