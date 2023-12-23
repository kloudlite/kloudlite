import { Button, IconButton } from '~/components/atoms/button';
import { IExtRemixCtx } from '~/root/lib/types/common';
import getQueries from '~/root/lib/server/helpers/get-queries';
import { getCookie } from '~/root/lib/app-setup/cookies';
import withContext from '~/root/lib/app-setup/with-contxt';
import { useLoaderData, useLocation, useNavigate } from '@remix-run/react';
import md5 from '~/root/lib/client/helpers/md5';
import { Avatar } from '~/components/atoms/avatar';
import { cn } from '~/components/utils';
import { Power } from '@jengaicons/react';
import { GQLServerHandler } from '../server/gql/saved-queries';

function CliLogin() {
  const { loginId, user } = useLoaderData();

  const location = useLocation();
  const navigate = useNavigate();

  const onAuthenticate = () => {
    if (!loginId) return;
    const cookie = getCookie();
    cookie.set('cliLogin', loginId);
    navigate('/');
  };

  if (!loginId) {
    return (
      <div className="flex min-h-screen justify-center items-center">
        <div className="flex flex-col items-center justify-center gap-5">
          <h1 className="text-5xl text-center font-bold text-black">
            <span className="text-primary-600">Login id</span> not provided
          </h1>
          <div className="font-medium w-full text-center text-2xl text-black tracking-wide">
            Please wait try again.
          </div>
        </div>
      </div>
    );
  }

  const { avatar, email, name } = user || {};

  const profile = () => {
    return (
      <div className="flex items-center gap-2 text-sm px-4 py-4 min-w-[15rem]">
        <div className="w-12 h-12 border border-secondary-400 p-0.5 rounded-full">
          {/* eslint-disable-next-line no-nested-ternary */}
          {avatar ? (
            <img
              alt="profile"
              src={avatar}
              className="rounded-full object-cover"
            />
          ) : email ? (
            <img
              className="rounded-full"
              alt={email}
              src={`https://www.gravatar.com/avatar/${md5(email)}?d=identicon`}
            />
          ) : (
            <Avatar size="lg" image={email} />
          )}
        </div>
        <div className="flex flex-col items-start">
          <span className="text-neutral-600 font-medium text-lg">
            {name || email}
          </span>
          <span className="text-primary-400 text-sm">
            {/* {role.replace('-', ' ')} */}
            {email}
          </span>
        </div>
      </div>
    );
  };

  const logOut = async () => {
    navigate(
      `${
        location.pathname +
        location.search +
        (location.search.trim() ? '&' : '?')
      }logout=yes`,
      { replace: true }
    );
  };

  return (
    <div className="flex min-h-screen justify-center items-center">
      <div className="flex flex-col items-center justify-center gap-12">
        <h1 className="text-5xl text-center font-bold text-black">
          Login to <span className="text-primary-600">Kloudlite CLI</span>
        </h1>
        <div
          className={cn(
            'font-medium w-full text-center text-2xl text-black tracking-wide px-6',
            {
              'bg-white border border-neutral-200 rounded-md ': user,
            }
          )}
        >
          {user ? (
            <div className="flex gap-6 justify-center items-center">
              <div>{profile()}</div>

              <Button
                variant="primary-outline"
                onClick={onAuthenticate}
                content="Authenticate"
              />

              <IconButton onClick={logOut} icon={<Power />} />
            </div>
          ) : (
            <Button
              size="lg"
              onClick={onAuthenticate}
              content="Login to Kloudlite"
            />
          )}
        </div>
      </div>
    </div>
  );
}

export const loader = async (ctx: IExtRemixCtx) => {
  const { loginId, logout } = await getQueries(ctx);
  if (logout) {
    const cookie = getCookie(ctx);

    Object.keys(cookie.getAll()).forEach((key) => {
      cookie.remove(key);
    });

    return withContext(ctx, { loginId });
  }

  const { data } = await GQLServerHandler(ctx.request).whoAmI();

  return {
    loginId: loginId || null,
    user: data || null,
  };
};

export default CliLogin;
