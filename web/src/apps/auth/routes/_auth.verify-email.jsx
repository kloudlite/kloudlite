import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import usePersistState from '~/root/lib/client/hooks/use-persist-state';
import { useEffect, useState } from 'react';
import { GQLServerHandler } from '~/auth/server/gql/saved-queries';
import getQueries from '~/root/lib/server/helpers/get-queries';
import logger from '~/root/lib/client/helpers/log';
import { useLoaderData, useNavigate } from '@remix-run/react';

import { redirect } from '@remix-run/node';
import { BrandLogo } from '~/components/branding/brand-logo';
import { Button } from '~/components/atoms/button';
import { ArrowRight } from '@jengaicons/react';
import { toast } from '~/components/molecule/toast';

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
    return (
      <div className="flex flex-col items-center justify-center gap-7xl h-full">
        <BrandLogo detailed={false} size={100} />
        <span className="heading2xl text-text-strong">
          Verifying details...
        </span>
      </div>
    );
  }

  return (
    <div className="h-full w-full flex items-center justify-center px-3xl">
      <div className="flex flex-col items-center gap-5xl md:w-[360px]">
        <BrandLogo detailed={false} size={60} />
        <div className="flex flex-col gap-5xl pb-5xl">
          <div className="flex flex-col items-center gap-2xl">
            <h3 className="heading3xl text-text-strong">Email verification</h3>
            <div className="bodyMd text-text-soft text-center">
              Please check your <span className="bodyMd-semibold">{email}</span>{' '}
              inbox to verify your account to get started.
            </div>
          </div>
          <Button
            content="Go back to Login"
            size="2xl"
            suffix={ArrowRight}
            block
          />
        </div>
        <div className="text-center">
          Didn’t get the email? Check your spam folder or{' '}
          <Button
            variant="primary-plain"
            content="Send it again"
            onClick={resendVerificationEmail}
            className="!inline-block"
          />
        </div>
      </div>
    </div>
  );
};

export const loader = async (ctx) => {
  const query = getQueries(ctx);
  const { data, errors } = await GQLServerHandler(ctx.request).whoAmI();
  if (errors) {
    console.error(errors[0].message);
    return redirect('/');
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
