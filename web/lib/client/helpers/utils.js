const { getCookie } = require('../../app-setup/cookies');
const { default: getQueries } = require('../../server/helpers/get-queries');

export const getNamespace = (ctx) => {
  const { namespace: p } = getQueries(ctx);
  const cookie = getCookie(ctx);
  const namespace = p || cookie.get('current_namespace');
  cookie.set('current_namespace', namespace);

  return namespace;
};
