import { Link } from '@remix-run/react';
import { GoogleReCaptcha } from 'react-google-recaptcha-v3';
import { Button } from '~/components/atoms/button';
import { BrandLogo } from '~/components/branding/brand-logo.jsx';
import { cn } from '~/components/utils';
import { getCookie } from '~/root/lib/app-setup/cookies';
import { redirectWithContext } from '~/root/lib/app-setup/with-contxt';
import { assureNotLoggedIn } from '~/root/lib/server/helpers/minimal-auth';
import { IExtRemixCtx } from '~/root/lib/types/common';
import Container from '../components/container';

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
    <Container>
      <div
        className={cn(
          'flex flex-col items-center self-stretch justify-center '
        )}
      >
        <div className="flex flex-col items-stretch justify-center gap-5xl md:w-[508px]">
          <BrandLogo darkBg={false} size={60} />
          <div className="flex flex-col items-stretch gap-7xl pb-5xl">
            <div className="flex flex-col gap-lg items-center">
              <div className={cn('text-text-strong heading3xl text-center')}>
                Join Chromatic Labs on Kloudlite
              </div>
              <div className="text-text-soft bodySm text-center">
                Simplify Collaboration and Enhance Productivity with Kloudlite
                teams.
              </div>
            </div>
            <div className="flex flex-col gap-3xl md:w-[308px] self-center">
              <Button
                size="2xl"
                variant="tertiary"
                content={
                  <span className="bodyLg-medium">Login to Kloudlite</span>
                }
                block
                LinkComponent={Link}
                to="/login"
              />
              <Button
                size="2xl"
                variant="primary"
                content={
                  <span className="bodyLg-medium">Signup to Kloudlite</span>
                }
                block
                LinkComponent={Link}
                to="/signup"
              />
            </div>

            <div className="bodyMd text-text-soft text-center self-center md:w-[350px]">
              By continuing, you agree to the{' '}
              <Link
                to="https://kloudlite.io/terms-and-conditions"
                className="underline"
              >
                Terms of Service
              </Link>{' '}
              and{' '}
              <Link
                className="underline"
                to="https://kloudlite.io/privacy-policy"
              >
                Privacy Policy
              </Link>
              .
            </div>
          </div>
        </div>
        <GoogleReCaptcha onVerify={() => {}} />
      </div>
    </Container>
  );
};

export default InvitePage;
