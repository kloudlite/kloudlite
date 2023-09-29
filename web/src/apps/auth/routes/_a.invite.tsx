import { BrandLogo } from '~/components/branding/brand-logo.jsx';
import { ArrowRight } from '@jengaicons/react';
import { assureNotLoggedIn } from '~/root/lib/server/helpers/minimal-auth';
import { IExtRemixCtx } from '~/root/lib/types/common';
import { getCookie } from '~/root/lib/app-setup/cookies';
import { redirectWithContext } from '~/root/lib/app-setup/with-contxt';
import { Button } from '~/components/atoms/button';
import { Link } from '@remix-run/react';

export const loader = async (ctx: IExtRemixCtx) => {
  const loggedIn = await assureNotLoggedIn(ctx);
  const cookie = getCookie(ctx);

  if (loggedIn) {
    cookie.set('url_history', '/invites');
    return redirectWithContext(ctx, '/');
  }

  return {};
};

const InvitePage = () => {
  return (
    <div className="h-full w-full flex items-center justify-center px-3xl">
      <div className="flex flex-col items-center gap-5xl md:w-[360px]">
        <BrandLogo detailed={false} size={60} />
        <div className="flex flex-col gap-5xl pb-5xl">
          <div className="flex flex-col items-center gap-2xl">
            <h3 className="heading3xl text-text-strong">Accept Invitation</h3>
            <div className="bodyMd text-text-soft text-center">
              You need to be logged in to accept the invitation.
            </div>
          </div>
          <Button
            LinkComponent={Link}
            to="/login"
            content="Login & Accept"
            size="2xl"
            suffix={<ArrowRight />}
            block
          />
        </div>
        <div className="text-center">
          Don&apos;t have an account?{' '}
          <Button
            variant="primary-plain"
            content="Sign Up & Accept"
            LinkComponent={Link}
            to="/signup"
            // onClick={resendVerificationEmail}
            className="!inline-block"
          />
        </div>
      </div>
    </div>
  );
};

export default InvitePage;
