import { useNavigate } from '@remix-run/react';
import { BrandLogo } from '~/components/branding/brand-logo';
import { IExtRemixCtx } from '~/root/lib/types/common';
import { GQLServerHandler } from '~/auth/server/gql/saved-queries';
import { handleError } from '~/root/lib/utils/common';
import useExtLoaderData from '~/root/lib/client/hooks/use-custom-loader-data';
import { useEffect } from 'react';

export const loader = async (ctx: IExtRemixCtx) => {
  const { errors } = await GQLServerHandler(ctx.request).logout();

  if (errors) {
    return handleError(errors[0]);
  }

  return {
    done: true,
  };
};

const LogoutPage = () => {
  const navigate = useNavigate();

  const { done } = useExtLoaderData<typeof loader>();

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

export default LogoutPage;
