import Root, { links as baseLinks } from '~/lib/app-setup/root.jsx';
import authStylesUrl from './styles/index.css';

export { loader } from '~/lib/app-setup/root.jsx';

export const links = () => {
  return [...baseLinks(), { rel: 'stylesheet', href: authStylesUrl }];
};

export default Root;
