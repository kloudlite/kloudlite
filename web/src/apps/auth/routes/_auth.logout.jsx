import { GQLServerHandler } from '~/auth/server/gql/saved-queries';
import { getCookie } from '~/root/lib/app-setup/cookies';
import withContext from '~/root/lib/app-setup/with-contxt';
import { useNavigate } from '@remix-run/react';
import { useEffect } from 'react';
import { BrandLogo } from '~/components/branding/brand-logo';

const LogoutPage = () => {
  const navigate = useNavigate();
  useEffect(() => {
    navigate('/login');
  }, []);
  return (
    <div className="flex flex-col items-center justify-center gap-7xl h-full">
      <BrandLogo detailed={false} size={100} />
      <span className="heading2xl text-text-strong">Logging out...</span>
    </div>
  );
};

export const loader = async (ctx) => {
  await GQLServerHandler(ctx.request).logout();

  const cookie = getCookie(ctx);

  console.log(cookie.getAll());

  Object.keys(cookie.getAll()).forEach((key) => {
    if (key !== 'url_history') cookie.remove(key);
  });

  console.log(cookie.getAll());

  return withContext(ctx, {});
};

export default LogoutPage;
