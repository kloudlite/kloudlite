import Container from '~/auth/components/container';
import { BrandLogo } from '~/components/branding/brand-logo';
import { getCookie } from '~/root/lib/app-setup/cookies';
import withContext from '~/root/lib/app-setup/with-contxt';
import { IExtRemixCtx } from '~/root/lib/types/common';

const cliLoggedIn = () => {
  return (
    <Container>
      <div className="flex flex-col gap-5xl">
        <BrandLogo darkBg={false} size={60} />
        <div className="flex flex-col gap-lg text-center max-w-[400px] items-center">
          <h1 className="headingXl">
            <span className="">Logged in</span> Successfully
          </h1>
          <div className="bodyMd text-text-strong">Visit your terminal.</div>
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
