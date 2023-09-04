import { GQLServerHandler as consoleHandler } from '~/console/server/gql/saved-queries';
import { GQLServerHandler as libHandler } from '../lib/server/gql/saved-queries';
import { loader } from './loader';

const init = () => {
  loader(consoleHandler, 'console');
  loader(libHandler, 'lib');
};

init();
