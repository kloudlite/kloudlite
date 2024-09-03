import { useLoaderData, useNavigate, useSearchParams } from '@remix-run/react';
import { useAuthApi } from '~/auth/server/gql/api-provider';
import { BrandLogo } from '~/components/branding/brand-logo';
import { toast } from '~/components/molecule/toast';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import getQueries from '~/root/lib/server/helpers/get-queries';
import { IRemixCtx } from '~/root/lib/types/common';
import { handleError } from '~/root/lib/utils/common';

export const decodeState = (str: string) =>
  Buffer.from(str, 'base64url').toString('utf8');

const CallBack = () => {
  const { query, state, provider, setupAction } = useLoaderData();
  const api = useAuthApi();
  const navigate = useNavigate();
  const [searchParams, _setSearchParams] = useSearchParams();

  useDebounce(
    () => {
      (async () => {
        if (setupAction === 'install') {
          if (window.opener) {
            window.opener.postMessage(
              {
                type: 'install',
                query,
              },
              '*'
            );
            window.close();
            return;
          }
          toast.error('Install failed');
        }
        try {
          if (state === 'redirect:add-provider') {
            const { errors } = await api.addOauthCredientials({
              ...query,
              provider,
            });

            if (errors && errors.length) {
              toast.error(errors[0].message);
            } else {
              if (window.opener) {
                window.opener.postMessage(
                  {
                    type: 'add-provider',
                    query,
                    status: 'success',
                    provider,
                  },
                  '*'
                );
                window.close();
                return;
              }
              window.close();
            }
            return;
          }

          const { errors } = await api.oauthLogin({
            ...query,
            provider,
          });

          if (errors && errors.length > 0) {
            toast.error(errors[0].message);
            navigate('/');
          } else {
            toast.success('Login Successful');
            const callback = searchParams.get('callback');
            if (callback) {
              const {
                data: { email, name },
              } = await api.whoAmI({});
              const encodedData = btoa(`email=${email}&name=${name}`);
              window.location.href = `${callback}?${encodedData}`;
              return;
            }
            navigate('/');
          }
        } catch (err) {
          handleError(err);
          navigate('/');
        }
      })();
    },
    200,
    []
  );

  return (
    <div className="flex flex-col items-center justify-center gap-3xl h-full">
      <BrandLogo detailed={false} size={56} />
      <span className="headingLg text-text-strong">Verifying details...</span>
    </div>
  );
};

export const loader = async (ctx: IRemixCtx) => {
  const { provider } = ctx.params;
  const queries = getQueries(ctx);
  const {
    state,
    setup_action: setupAction,
    installation_id: installationId,
  } = queries;

  let queryData;
  if (state && provider === 'github')
    queryData = JSON.parse(decodeState(state));

  return {
    setupAction,
    installationId,
    query: queries,
    state: queryData?.state || state,
    provider,
  };
};

export default CallBack;
