import { getCookie } from '~/root/lib/app-setup/cookies';
import withContext from '~/root/lib/app-setup/with-contxt';
import { useNavigate, useLoaderData } from '@remix-run/react';
import { useEffect } from 'react';
import { BrandLogo } from '~/components/branding/brand-logo';
import { IExtRemixCtx } from '~/root/lib/types/common';

const LogoutPage = () => {
  const navigate = useNavigate();
  const { done } = useLoaderData();

  useEffect(() => {
    if (done) {
      navigate('/');
    }
  }, [done]);
  return (
    <div className="flex flex-col items-center justify-center gap-7xl h-full">
      <BrandLogo detailed={false} size={100} />
      <span className="heading2xl text-text-strong">Logging out...</span>
    </div>
  );
};

export const loader = async (ctx: IExtRemixCtx) => {
  const cookie = getCookie(ctx);

  const keys = Object.keys(cookie.getAll());

  for (let i = 0; i < keys.length; i += 1) {
    const key = keys[i];
    if (key === 'hotspot-session') {
      cookie.remove(key);
    }
  }

  return withContext(ctx, {
    done: 'true',
  });
};

export default LogoutPage;
