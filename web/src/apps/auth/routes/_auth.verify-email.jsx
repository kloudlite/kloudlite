import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import usePersistState from '~/root/lib/client/hooks/use-persist-state';
import { useEffect, useState } from 'react';
import { toast } from 'react-toastify';
import { GQLServerHandler } from '~/auth/server/gql/saved-queries';
import getQueries from '~/root/lib/server/helpers/get-queries';
import logger from '~/root/lib/client/helpers/log';
import { useLoaderData, useNavigate } from '@remix-run/react';

import { redirect } from '@remix-run/node';

const VerifyEmail = () => {
  const { query, email } = useLoaderData();
  const navigate = useNavigate();
  const { token } = query;
  const api = useAPIClient();

  const [rateLimiter, setRateLimiter] = usePersistState('rateLimiter', {});

  useEffect(() => {
    (async () => {
      try {
        if (!token) return;
        const { _, errors } = await api.verifyEmail({
          token,
        });
        if (errors) {
          throw errors[0];
        }

        toast.success('Email verified successfully');
        navigate('/');
      } catch (error) {
        logger.error(error);
        toast.error(error.message);
      }
    })();
    // (async () => {
    //   try {
    //     const { errors } = await api.whoAmI();
    //     if (errors) {
    //       throw errors[0];
    //     }
    //   } catch (err) {
    //     toast.error(err.message);
    //   }
    // })();
  }, []);

  const [isSending, setSending] = useState(false);

  const resendVerificationEmail = () => {
    (async () => {
      try {
        if (!email) {
          toast.error('Something went wrong! Please try again.');
          return;
        }
        const { lastSent } = rateLimiter;
        if (lastSent) {
          const diff = Date.now() - lastSent;
          if (diff < 60000) {
            toast.error('Please wait for 60 seconds before resending email');
            return;
          }
        }

        if (isSending) {
          toast.error('Please Wait we are sending email');
          return;
        }

        setSending(true);

        const { errors } = await api.resendVerificationEmail({ email });

        setSending(false);

        if (errors) {
          toast.error(errors[0].message);
          return;
        }

        setRateLimiter((s) => ({ ...s, lastSent: Date.now() }));

        toast.success("We've sent you a new verification email");
      } catch (err) {
        setSending(false);
        toast.error(err.message);
      }
    })();
  };

  if (token) {
    return <div>verifying details please wait</div>;
  }

  return (
    <div className="justify-center flex-1">
      <div className="flex flex-col items-center p-12 h-full overflow-auto">
        <div className="flex flex-col flex-1 w-[434px] justify-center gap-12">
          <div className="flex flex-col gap-4">
            <div className="font-bold text-5xl pr-12 leading-tight text-black">
              We sent you an
              <br />
              <span className="text-primary">Email!</span>
            </div>
            <div className="text-secondary">Please check your inbox .</div>
          </div>
          <div className="flex justify-start gap-2 items-center">
            <span className="text-secondary">Didn&apos;t get Email? </span>
            <div
              onClick={resendVerificationEmail}
              className="text-primary font-medium cursor-pointer"
            >
              Resend verification email.
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export const loader = async (ctx) => {
  const query = getQueries(ctx);
  const { data, errors } = await GQLServerHandler({
    headers: ctx.request.headers,
  }).whoAmI();
  if (errors) {
    logger.error(errors[0].message);
  }
  const { email, verified } = data || {};

  if (verified) {
    return redirect('/');
  }

  return {
    query,
    email: email || '',
  };
};

export default VerifyEmail;
