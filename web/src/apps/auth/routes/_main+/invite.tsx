import { Link } from '@remix-run/react';
import { Button } from '@kloudlite/design-system/atoms/button';
import { cn } from '@kloudlite/design-system/utils';
import { getCookie } from '~/root/lib/app-setup/cookies';
import { redirectWithContext } from '~/root/lib/app-setup/with-contxt';
import { assureNotLoggedIn } from '~/root/lib/server/helpers/minimal-auth';
import { IExtRemixCtx } from '~/root/lib/types/common';
import { mainUrl } from '~/auth/consts';
import Container from '../../components/container';

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
    <Container
      headerExtra={
        <Button
          variant="outline"
          content="Sign in"
          linkComponent={Link}
          to="/login"
        />
      }
    >
      <div className="flex flex-col gap-6xl md:w-[500px] px-3xl py-5xl md:px-9xl">
        <div className="flex flex-col gap-lg items-center text-center">
          <div className={cn('text-text-strong headingXl text-center')}>
            Join Chromatic Labs on Kloudlite
          </div>
          <div className="bodyMd-medium text-text-soft">
            Simplify Collaboration and Enhance Productivity with Kloudlite
            teams.
          </div>
        </div>
        <div className="flex flex-col gap-3xl">
          <Button
            size="lg"
            variant="tertiary"
            content={<span className="bodyLg-medium">Login to Kloudlite</span>}
            block
            linkComponent={Link}
            to="/login"
          />
          <Button
            size="lg"
            variant="primary"
            content={<span className="bodyLg-medium">Signup to Kloudlite</span>}
            block
            linkComponent={Link}
            to="/signup"
          />
        </div>
        <div className="inline text-center">
          <span className="text-text-soft bodyLg">
            By continuing, you agree Kloudlite’s
          </span>
          <br />
          <span>
            <Button
              to={`${mainUrl}/terms-of-services`}
              linkComponent={Link}
              className="!inline-block align-bottom"
              variant="plain"
              content={
                <span className="bodyLg-underline text-text-strong">
                  Terms of Service
                </span>
              }
            />
            <span> and </span>
            <Button
              to={`${mainUrl}/privacy-policy`}
              linkComponent={Link}
              className="!inline-block align-bottom"
              variant="plain"
              content={
                <span className="bodyLg-underline text-text-strong">
                  Privacy Policy
                </span>
              }
            />
            .
          </span>
        </div>
      </div>
    </Container>
  );
};

export default InvitePage;
