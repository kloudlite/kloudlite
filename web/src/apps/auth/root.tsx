import Root, { links as baseLinks } from '~/lib/app-setup/root.jsx';
import { ChildrenProps } from '~/components/types';
import ThemeProvider from '~/root/lib/client/hooks/useTheme';
import authStylesUrl from './styles/index.css';
import { RECAPTCHA_SITE_KEY } from './consts';
import { useEffect } from 'react';
import { useLocation } from '@remix-run/react';
import grecaptcha from '~/root/lib/client/helpers/g-recaptcha';

export { loader } from '~/lib/app-setup/root.jsx';
export { shouldRevalidate } from '~/lib/app-setup/root.jsx';

export const links = () => {
  return [...baseLinks(), { rel: 'stylesheet', href: authStylesUrl }];
};

export { ErrorBoundary } from '~/lib/app-setup/root';

const Layout = ({ children }: ChildrenProps) => {
  const { pathname, search } = useLocation();
  const hideCaptcha = () => {
    let x = document.querySelector('.grecaptcha-badge') as HTMLDivElement;
    if (x) {
      if (
        (pathname.startsWith('/signup') && search === '?mode=email') ||
        (pathname.startsWith('/login') && search === '?mode=email') ||
        pathname.startsWith('/forgot-password')
      ) {
        x.style.setProperty('display', 'block', 'important');
        x.style.setProperty('visibility', 'visible');
      } else {
        x.style.display = 'none';
        x.style.setProperty('visibility', 'hidden');
      }
    }
  };
  useEffect(() => {
    const script = document.createElement('script');
    script.async = true;
    script.src = `https://www.google.com/recaptcha/enterprise.js?render=${RECAPTCHA_SITE_KEY}`;
    script.onload = () => {
      grecaptcha.onReady(async () => {
        hideCaptcha();
      });
    };
    document.head.appendChild(script);
  }, []);

  useEffect(() => {
    hideCaptcha();
  }, [pathname, search]);

  return (
    // <SSRProvider>
    // eslint-disable-next-line react/jsx-no-useless-fragment
    <>{children}</>
    // </SSRProvider>
  );
};

const _Root = ({ ...props }) => {
  // @ts-ignore
  return (
    <ThemeProvider>
      <Root {...props} Wrapper={Layout} />
    </ThemeProvider>
  );
};

export default _Root;
