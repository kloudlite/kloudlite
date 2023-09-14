import Root, { links as baseLinks } from '~/lib/app-setup/root';
import { ChildrenProps } from '~/components/types';
import authStylesUrl from './styles/index.css';
import highlightCss from './styles/hljs/tokyo-night-dark.min.css';

export { loader } from '~/lib/app-setup/root.jsx';
export { shouldRevalidate } from '~/lib/app-setup/root.jsx';

export const links = () => {
  return [
    ...baseLinks(),
    { rel: 'stylesheet', href: authStylesUrl },
    { rel: 'stylesheet', href: highlightCss },
  ];
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
  return <Root {...props} Wrapper={Layout} />;
};

export default _Root;
