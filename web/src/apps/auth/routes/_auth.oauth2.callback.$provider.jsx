import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { useEffect } from 'react';
import { toast } from 'react-toastify';
import { useNavigate, useLoaderData } from '@remix-run/react';
import getQueries from '~/root/lib/server/helpers/get-queries';

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

  return <div>verifying details please wait</div>;
};

export const loader = async (ctx) => {
  const { provider } = ctx.params;
  return {
    query: getQueries(ctx),
    provider,
  };
};

export default CallBack;
