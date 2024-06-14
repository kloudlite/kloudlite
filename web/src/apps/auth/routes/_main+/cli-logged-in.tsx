import { Link } from '@remix-run/react';
import Container from '~/auth/components/container';
import { Button } from '~/components/atoms/button';
import { getCookie } from '~/root/lib/app-setup/cookies';
import withContext from '~/root/lib/app-setup/with-contxt';
import { IExtRemixCtx } from '~/root/lib/types/common';

const cliLoggedIn = () => {
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
          <div className="text-text-strong headingXl text-center">
            Logged in successfully
          </div>
          <div className="bodyMd-medium text-text-soft">
            Visit your terminal.
          </div>
        </div>
      </div>
    </Container>
  );
};

export const loader = (ctx: IExtRemixCtx) => {
  const cookie = getCookie(ctx);
  cookie.remove('cliLogin');
  return withContext(ctx, {});
};

export default cliLoggedIn;
