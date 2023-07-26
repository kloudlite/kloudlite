import { GQLServerHandler } from '~/auth/server/gql/saved-queries';
import { getCookie } from '~/root/lib/app-setup/cookies';
import withContext from '~/root/lib/app-setup/with-contxt';
import { useNavigate } from '@remix-run/react';
import { useEffect } from 'react';

const LogoutPage = () => {
  const navigate = useNavigate();
  useEffect(() => {
    navigate('/login');
  }, []);
  return <div>Logging Out</div>;
};

export const loader = async (ctx) => {
  await GQLServerHandler(ctx.request).logout();

  const cookie = getCookie(ctx);

  console.log(cookie.getAll());

  Object.keys(cookie.getAll()).forEach((key) => {
    cookie.remove(key);
  });

  console.log(cookie.getAll());

  return withContext(ctx, {});
};

export default LogoutPage;
