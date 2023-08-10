import { BellSimpleFill, CaretDownFill } from '@jengaicons/react';
import { Link, useLocation, useMatch } from '@remix-run/react';
import { useParams } from 'react-router-dom';
import { Button, IconButton } from '~/components/atoms/button.jsx';
import OptionList from '~/components/atoms/option-list';
import { BrandLogo } from '~/components/branding/brand-logo.jsx';
import { Profile } from '~/components/molecule/profile.jsx';
import { TopBar } from '~/components/organisms/top-bar.jsx';

const Container = ({ children }) => {
  const location = useLocation();
  const match = useMatch(
    {
      path: '/:path/*',
    },

    location.pathname
  );

  return (
    <div className="flex flex-col">
      <TopBar
        linkComponent={Link}
        fixed
        logo={
          <div>
            <div className="hidden md:block">
              <BrandLogo detailed size={24} />
            </div>
            <div className="block md:hidden">
              <BrandLogo size={24} />
            </div>
          </div>
        }
        tab={{
          value: match?.params?.path,
          fitted: true,
          layoutId: 'project',
          onChange: (e) => {
            console.log(e);
          },
          items: [
            {
              label: 'Project',
              href: '/project',
              key: 'project',
              value: 'project',
            },
            {
              label: 'Cluster',
              href: '/cluster',
              key: 'cluster',
              value: 'cluster',
            },
            {
              label: 'Cloud provider',
              href: '#',
              key: 'cloudprovider',
              value: 'cloudprovider',
            },
            {
              label: 'Domains',
              href: '#',
              key: 'domains',
              value: 'domains',
            },
            {
              label: 'Container registry',
              href: '#',
              value: 'containerregistry',
              key: 'containerregistry',
            },
            {
              label: 'VPN',
              href: '#',
              key: 'vpn',
              value: 'vpn',
            },
            {
              label: 'Settings',
              href: '/settings/general',
              key: 'settings',
              value: 'settings',
            },
          ],
        }}
        actions={
          <div className="flex flex-row gap-2xl items-center">
            <TopBarMenu />
            <div className="h-[15px] w-xs bg-border-default" />
            {/* for screens md or larger */}
            <div className="hidden md:flex flex-row gap-lg items-center justify-center">
              <IconButton icon={BellSimpleFill} variant="plain" />
              <Profile name="Astroman" size="small" subtitle={null} />
            </div>
            {/* for screen smaller than md */}
            <div className="flex md:hidden flex-row gap-lg items-center justify-center">
              <IconButton icon={BellSimpleFill} variant="plain" />
              <Profile name="" size="small" subtitle={null} />
            </div>
          </div>
        }
      />
      <div className="w-full max-w-8xl mx-auto">{children}</div>
    </div>
  );
};

export default Container;

// OptionList for various actions
const TopBarMenu = ({ open, setOpen }) => {
  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <Button content="Nuveo" variant="outline" suffix={CaretDownFill} />
      </OptionList.Trigger>
      <OptionList.Content />
    </OptionList.Root>
  );
};
