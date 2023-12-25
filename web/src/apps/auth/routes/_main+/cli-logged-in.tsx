import { getCookie } from '~/root/lib/app-setup/cookies';
import withContext from '~/root/lib/app-setup/with-contxt';
import { IExtRemixCtx } from '~/root/lib/types/common';

const cliLoggedIn = () => {
  return (
    <div className="flex min-h-screen justify-center items-center">
      <div className="flex flex-col items-center justify-center gap-5">
        <h1 className="text-5xl text-center font-bold text-black">
          <span className="text-primary-600">Logged in</span> Successfully
        </h1>
        <div className="font-medium w-full text-center text-2xl text-black tracking-wide">
          Visit your terminal.
        </div>
      </div>
    </div>
  );
};

export const loader = (ctx: IExtRemixCtx) => {
  const cookie = getCookie(ctx);
  cookie.remove('cliLogin');
  return withContext(ctx, {});
};

export default cliLoggedIn;
