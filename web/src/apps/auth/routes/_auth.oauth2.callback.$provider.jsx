import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { useEffect } from 'react';
import { useNavigate, useLoaderData } from '@remix-run/react';
import getQueries from '~/root/lib/server/helpers/get-queries';
import { BrandLogo } from '~/components/branding/brand-logo';
import { toast } from '~/components/molecule/toast';

const CallBack = () => {
  const { query, provider } = useLoaderData();
  const api = useAPIClient();
  const navigate = useNavigate();
  useEffect(() => {
    (async () => {
      try {
        const { errors } = await api.oauthLogin({
          ...query,
          provider,
        });

        if (errors && errors.length > 0) {
          toast.error(errors[0].message);
          navigate('/');
        } else {
          toast.success('Login Successful');
          navigate('/');
        }
      } catch (e) {
        toast.error(e.message);
        navigate('/');
      }
    })();
  }, [query]);

  return (
    <div className="flex flex-col items-center justify-center gap-7xl h-full">
      <BrandLogo detailed={false} size={100} />
      <span className="heading2xl text-text-strong">Verifying details...</span>
    </div>
  );
};

export const loader = async (ctx) => {
  const { provider } = ctx.params;
  return {
    query: getQueries(ctx),
    provider,
  };
};

export default CallBack;
