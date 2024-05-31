import { redirect } from '@remix-run/node';
import { useNavigate } from '@remix-run/react';
import { BrandLogo } from '~/components/branding/brand-logo';
import { toast } from '~/components/molecule/toast';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import useForm from '~/root/lib/client/hooks/use-form';
import getQueries from '~/root/lib/server/helpers/get-queries';
import Yup from '~/root/lib/server/helpers/yup';
import { IRemixCtx } from '~/root/lib/types/common';
import { GQLServerHandler } from '~/lib/server/gql/saved-queries';
import { Badge } from '~/components/atoms/badge';
import TextInputLg from '~/console/components/text-input-lg';

const InviteCode = () => {
  const api = useConsoleApi();
  const navigate = useNavigate();

  const { values, handleChange, submit, handleSubmit } = useForm({
    initialValues: {
      inviteCode: '',
    },
    validationSchema: Yup.object({
      inviteCode: Yup.string().required('invite code is required'),
    }),
    onSubmit: async (v) => {
      try {
        const { errors } = await api.verifyInviteCode({
          invitationCode: v.inviteCode,
        });
        if (errors) {
          throw new Error(errors[0].message);
        }
        toast.success('Invitation code verification successful.');
        navigate('/teams');
      } catch (err) {
        const errorMessage =
          err instanceof Error
            ? err.message
            : 'An error occurred. Please try again.';
        toast.error(errorMessage);
      }
    },
  });

  return (
    <div className="flex flex-col items-center justify-center gap-4xl h-full max-w-[734px] m-auto">
      <BrandLogo detailed={false} size={100} />
      <Badge type="neutral">
        🔥 Amazing curated{' '}
        <span className="bodyMd-semibold text-text-strong">Open-Source</span>{' '}
        remote local envs
      </Badge>
      <span className="heading5xl-marketing text-text-strong">
        Unlock early access now!
      </span>
      <span className="text-center bodyXl">
        Don't miss the chance to try our product. Enter your referral code now
        to move up the waitlist and secure early access!
      </span>
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-[500px] flex items-center"
      >
        <div className="flex flex-col gap-3xl w-full">
          <TextInputLg
            value={values.inviteCode}
            onChange={handleChange('inviteCode')}
            onEnter={submit}
          />
        </div>
      </form>
    </div>
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
  const { email, approved } = data || {};

  if (approved) {
    return redirect('/teams');
  }

  return {
    query,
    email: email || '',
  };
};

export default InviteCode;
