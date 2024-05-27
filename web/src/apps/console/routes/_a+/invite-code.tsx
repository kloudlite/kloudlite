import { redirect } from '@remix-run/node';
import { GQLServerHandler } from '~/auth/server/gql/saved-queries';
import { BrandLogo } from '~/components/branding/brand-logo';
import getQueries from '~/root/lib/server/helpers/get-queries';
import { IRemixCtx } from '~/root/lib/types/common';

const InviteCode = () => {
  return (
    <div className="flex flex-col items-center justify-center gap-7xl h-full">
      <BrandLogo detailed={false} size={100} />
      <span className="heading2xl text-text-strong">Invite Code Details:</span>
      <div className="bodyLg text-text-default">
        Thank you for sign up, we have received your request and we are on it.
      </div>
    </div>
  );
};

export const loader = async (ctx: IRemixCtx) => {
  const query = getQueries(ctx);
  const { data, errors } = await GQLServerHandler(ctx.request).whoAmI();
  if (errors) {
    return {
      query,
    };
  }
  const { email, approved } = data || {};

  if (approved) {
    return redirect('/teams');
  }

  return {
    query,
    email: email || '',
  };
};

export default InviteCode;
