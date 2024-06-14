import Root, { links as baseLinks } from '~/lib/app-setup/root.jsx';
import { ChildrenProps } from '~/components/types';
import ThemeProvider from '~/root/lib/client/hooks/useTheme';
import authStylesUrl from './styles/index.css';

export { loader } from '~/lib/app-setup/root.jsx';
export { shouldRevalidate } from '~/lib/app-setup/root.jsx';

export const links = () => {
  return [...baseLinks(), { rel: 'stylesheet', href: authStylesUrl }];
};

export { ErrorBoundary } from '~/lib/app-setup/root';

const Layout = ({ children }: ChildrenProps) => {
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
