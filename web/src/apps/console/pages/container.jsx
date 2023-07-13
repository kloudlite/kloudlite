import { BellSimpleFill, CaretDownFill } from '@jengaicons/react';
import {
  Link,
  Links,
  LiveReload,
  Outlet,
  useLocation,
  useMatch,
} from '@remix-run/react';
import classNames from 'classnames';
import { Button, IconButton } from '~/components/atoms/button.jsx';
import { BrandLogo } from '~/components/branding/brand-logo.jsx';
import { Profile } from '~/components/molecule/profile.jsx';
import { TopBar } from '~/components/organisms/top-bar.jsx';

const Container = ({ children }) => {
  const fixedHeader = true;

  const location = useLocation();
  console.log('location', location.pathname);
  const match = useMatch(
    {
      path: '/:path/*',
    },
    location.pathname
  );

  console.log('match', match);
  return (
    <div className="px-2.5">
      {'' !== 'newproject' && (
        <TopBar
          linkComponent={Link}
          fixed={fixedHeader}
          logo={<BrandLogo detailed size={20} />}
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
              <Button
                content="Nuveo"
                variant="basic"
                DisclosureComp={CaretDownFill}
              />
              <div className="h-[15px] w-xs bg-border-default" />
              <div className="flex flex-row gap-lg items-center justify-center">
                <IconButton icon={BellSimpleFill} variant="plain" />
                <Profile name="Astroman" size="small" subtitle={null} />
              </div>
            </div>
          }
        />
      )}
      <div
        className={classNames('max-w-[1184px] m-auto', {
          'pt-23.75': fixedHeader && !('' === 'newproject'),
          'pt-15': '' === 'newproject',
        })}
      >
        {children}
      </div>
    </div>
  );
};

export default Container;
