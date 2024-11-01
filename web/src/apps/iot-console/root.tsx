/* eslint-disable react/jsx-no-useless-fragment */
import Root, { links as baseLinks } from '~/lib/app-setup/root';
import { ChildrenProps } from '@kloudlite/design-system/types';
import { SockProvider } from '~/root/lib/client/helpers/socket/context';
import authStylesUrl from './styles/index.css';
import highlightCss from './styles/hljs/tokyo-night-dark.min.css';
import { DataContextProvider } from './page-components/common-state';

export { loader } from '~/lib/app-setup/root.jsx';
export { shouldRevalidate } from '~/lib/app-setup/root.jsx';

const tagId = 'G-HCK0NJNTJ6';

export const links = () => {
  return [
    ...baseLinks(),
    { rel: 'stylesheet', href: authStylesUrl },
    { rel: 'stylesheet', href: highlightCss },
  ];
};

export { ErrorBoundary } from '~/lib/app-setup/root';

const Layout = ({ children }: ChildrenProps) => {
  return <>{children}</>;
};

const _Root = ({ ...props }) => {
  return (
    <SockProvider>
      <DataContextProvider>
        <Root {...props} Wrapper={Layout} tagId={tagId} />
      </DataContextProvider>
    </SockProvider>
  );
};

export default _Root;
