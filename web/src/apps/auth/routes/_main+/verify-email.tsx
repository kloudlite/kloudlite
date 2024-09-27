import usePersistState from '~/root/lib/client/hooks/use-persist-state';
import { useEffect, useState } from 'react';
import { GQLServerHandler } from '~/auth/server/gql/saved-queries';
import getQueries from '~/root/lib/server/helpers/get-queries';
import { Link, useLoaderData, useNavigate } from '@remix-run/react';

import { redirect } from '@remix-run/node';
import { BrandLogo } from '~/components/branding/brand-logo';
import { Button } from '~/components/atoms/button';
import { toast } from '~/components/molecule/toast';
import { handleError } from '~/root/lib/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import { ArrowLeft } from '~/components/icons';
import { cn } from '~/components/utils';
import Container from '~/auth/components/container';
import { useAuthApi } from '~/auth/server/gql/api-provider';

const VerifyEmail = () => {
  const { query, email } = useLoaderData();
  const navigate = useNavigate();
  const { token } = query;
  const api = useAuthApi();

  const [rateLimiter, setRateLimiter] = usePersistState('rateLimiter', {});

  useEffect(() => {
    (async () => {
      try {
        if (!token) return;
        const { errors } = await api.verifyEmail({
          token,
        });
        if (errors) {
          throw errors[0];
        }

        toast.success('Email verified successfully');
        navigate('/');
      } catch (error) {
        handleError(error);
      }
    })();
  }, []);

  const [isSending, setSending] = useState(false);

  const resendVerificationEmail = () => {
    (async () => {
      try {
        if (!email) {
          // TODO: handle this case, by taking email from user
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

        const { errors } = await api.resendVerificationEmail();

        setSending(false);

        if (errors) {
          toast.error(errors[0].message);
          return;
        }

        setRateLimiter((s: any) => ({
          ...s,
          lastSent: Date.now(),
        }));

        toast.success("We've sent you a new verification email");
      } catch (err) {
        setSending(false);
        handleError(err);
      }
    })();
  };

  if (token) {
    return (
      <div className="flex flex-col items-center justify-center gap-3xl h-full">
        <BrandLogo detailed={false} size={56} />
        <span className="headingLg text-text-strong">Verifying details...</span>
      </div>
    );
  }

  return (
    <Container>
      <div className="flex flex-col gap-6xl md:w-[500px] px-3xl py-5xl md:px-9xl">
        <div className="flex flex-col gap-lg items-center text-center">
          <div className={cn('text-text-strong headingXl text-center')}>
            Email verification
          </div>
          <div className="bodyMd-medium text-text-soft">
            Please check your <span className="text-text-default">{email}</span>{' '}
            inbox to verify your account to get started.
          </div>
        </div>
        <div className="flex flex-col gap-3xl">
          <Button
            size="lg"
            block
            variant="primary"
            content={
              <span className="bodyLg-medium">Resend verification link</span>
            }
            onClick={resendVerificationEmail}
          />
          <Button
            size="lg"
            variant="basic"
            content={<span className="bodyLg-medium">Go back to login</span>}
            prefix={<ArrowLeft />}
            block
            type="submit"
            linkComponent={Link}
            to="/logout"
          />
        </div>
      </div>
    </Container>
  );
};

export const loader = async (ctx: IRemixCtx) => {
  const query = getQueries(ctx);
  const { data, errors } = await GQLServerHandler(ctx.request).whoAmI();
  if (errors) {
    return {
      query,
    };
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
