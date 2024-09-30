import {
  GithubLogoFill,
  LinkedinLogoFill,
  TwitterNewLogoFill,
} from '@jengaicons/react';
import { Link } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import { BrandLogo } from '~/components/branding/brand-logo';
import { mainUrl } from '../consts';
import Wrapper from './wrapper';

const linkedinUrl = 'https://linkedin.com/company/kloudlite-io';
const gitUrl = 'https://github.com/kloudlite/kloudlite';
const xUrl = 'https://x.com/kloudlite';

const menu = [
  {
    title: 'Home',
    to: mainUrl,
  },
  {
    title: 'Documents',
    to: `${mainUrl}/docs`,
  },
  {
    title: 'Help & support',
    to: `${mainUrl}/help-and-support`,
  },
  {
    title: 'Contact us',
    to: `${mainUrl}/contact-us`,
  },
  {
    title: 'Blog',
    to: `${mainUrl}/blog`,
  },
  {
    title: 'Changelog',
    to: 'https://github.com/kloudlite/kloudlite/releases',
  },
  {
    title: 'Pricing',
    to: `${mainUrl}/pricing`,
  },
  {
    title: 'Legal',
    to: `${mainUrl}/legal/privacy-policy`,
  },
];

const SocialMenu = () => {
  const socialIconSize = 18;
  return (
    <div className="flex flex-row items-center gap-2xl text-icon-soft shrink-0">
      <a href={gitUrl} aria-label="kloudlite-github">
        <GithubLogoFill size={socialIconSize} />
      </a>
      <a href={xUrl} aria-label="kloudlite-x">
        <TwitterNewLogoFill size={socialIconSize} />
      </a>
      <a href={linkedinUrl} aria-label="kloudlite-linkedin">
        <LinkedinLogoFill size={socialIconSize} />
      </a>
    </div>
  );
};

const Footer = () => {
  return (
    <div className="bg-surface-basic-default">
      <Wrapper className="py-6xl flex flex-col gap-4xl">
        <div className="flex flex-row items-center justify-between">
          <div className="flex flex-col md:flex-row md:items-center gap-lg">
            <a href={mainUrl} aria-label="kloudlite">
              <BrandLogo size={24} detailed />
            </a>
            <span className="text-text-soft bodyMd">
              © {new Date().getFullYear()}
            </span>
          </div>
          <div className="flex flex-row items-center gap-3xl">
            <div className="hidden md:block lg:hidden">
              <SocialMenu />
            </div>
            {/* <ThemeSwitcher /> */}
          </div>
        </div>
        <div className="flex flex-col-reverse gap-4xl md:gap-0 md:flex-row md:items-center lg:justify-between">
          <div className="md:hidden lg:block">
            <SocialMenu />
          </div>
          <div className="flex flex-row flex-wrap gap-y-md lg:basis-auto items-center lg:gap-lg w-full lg:w-auto">
            {menu.map((m) => (
              <div
                key={m.title}
                className="basis-1/2 md:basis-1/4 lg:basis-auto"
              >
                <Button
                  to={m.to}
                  content={<span className="text-text-strong">{m.title}</span>}
                  variant="plain"
                  linkComponent={Link}
                />
              </div>
            ))}
          </div>
        </div>
      </Wrapper>
    </div>
  );
};

export default Footer;
