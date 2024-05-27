import { redirect } from '@remix-run/node';
import { useNavigate } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { BrandLogo } from '~/components/branding/brand-logo';
import { toast } from '~/components/molecule/toast';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import useForm from '~/root/lib/client/hooks/use-form';
import getQueries from '~/root/lib/server/helpers/get-queries';
import Yup from '~/root/lib/server/helpers/yup';
import { IRemixCtx } from '~/root/lib/types/common';
import { handleError } from '~/root/lib/utils/common';
import { GQLServerHandler } from '~/lib/server/gql/saved-queries';

const InviteCode = () => {
  const api = useConsoleApi();
  const navigate = useNavigate();

  const { values, handleChange, isLoading, errors, handleSubmit } = useForm({
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
          throw errors[0];
        }
        toast.success('Invitation code verification successfull.');
        navigate('/teams');
      } catch (err) {
        handleError(err);
      }
    },
  });

  return (
    <div className="flex flex-col items-center justify-center gap-7xl h-full">
      <BrandLogo detailed={false} size={100} />
      <span className="heading2xl text-text-strong">
        Validate Your Invite Code
      </span>
      <div className="bodyLg text-text-default text-center">
        Thank you for sign up, we have received your request and we are on it.
        <br />
        Currently, You are on waiting list. Thanks for your patience.
        <br />
      </div>
      <div className="bodyLg text-text-default text-center">
        If, You have invite code please proceed. Thanks.
      </div>

      <form
        onSubmit={handleSubmit}
        className="w-full max-w-[500px] flex items-center"
      >
        <div className="flex flex-col gap-3xl w-full">
          <TextInput
            label="Invite Code"
            size="lg"
            placeholder="Invitation Code"
            value={values.inviteCode}
            onChange={handleChange('inviteCode')}
            error={!!errors.inviteCode}
            message={errors.inviteCode}
          />
          <div className="flex flex-col items-center">
            <Button
              {...{
                variant: 'primary',
                content: 'Verify',
                loading: isLoading,
                type: 'submit',
              }}
            />
          </div>
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
