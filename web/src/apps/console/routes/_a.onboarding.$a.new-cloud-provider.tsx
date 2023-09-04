import { Button } from '~/components/atoms/button';
import { PasswordInput, TextInput } from '~/components/atoms/input';
import { BrandLogo } from '~/components/branding/brand-logo';
import { ProgressTracker } from '~/components/organisms/progress-tracker';
import { ArrowLeft, ArrowRight } from '@jengaicons/react';
import Select from '~/components/atoms/select';
import { useOutletContext, useNavigate, useParams } from '@remix-run/react';
import { Badge } from '~/components/atoms/badge';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { toast } from '~/components/molecule/toast';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { handleError } from '~/root/lib/utils/common';
import { UserMe } from '~/root/lib/server/gql/saved-queries';
import { parseName } from '~/root/src/generated/r-types/utils';
import RawWrapper from '../components/raw-wrapper';
import { IdSelector } from '../components/id-selector';
import { keyconstants } from '../server/r-urils/key-constants';
import { getMetadata } from '../server/r-urils/common';
import { getSecretRef } from '../server/r-urils/secret-ref';
import { ensureAccountClientSide } from '../server/utils/auth-utils';
import { Account } from '../server/gql/queries/account-queries';

const NewCloudProvider = () => {
  const { account, user } = useOutletContext<{
    account: Account;
    user: UserMe;
  }>();
  const { a: accountName } = useParams();
  const api = useAPIClient();

  const navigate = useNavigate();
  const { values, errors, handleSubmit, handleChange, isLoading } = useForm({
    initialValues: {
      displayName: '',
      name: '',
      provider: 'aws',
      accessKey: '',
      accessSecret: '',
    },
    validationSchema: Yup.object({
      displayName: Yup.string().required(),
      name: Yup.string().required(),
      provider: Yup.string().required(),
      accessKey: Yup.string().required(),
      accessSecret: Yup.string().required(),
    }),
    onSubmit: async (val) => {
      try {
        console.log(val);
        ensureAccountClientSide({ account: accountName });
        const { errors: e } = await api.createProviderSecret({
          secret: getSecretRef({
            metadata: getMetadata({
              name: val.name,
              annotations: {
                [keyconstants.displayName]: val.displayName,
                [keyconstants.author]: user.name,
              },
            }),
            stringData: {
              accessKey: val.accessKey,
              accessSecret: val.accessSecret,
            },
            cloudProviderName: val.provider,
          }),
        });
        if (e) {
          throw e[0];
        }
        toast.success('provider secret created successfully');
        navigate(`/onboarding/${accountName}/${val.name}/new-cluster`);
      } catch (err) {
        handleError(err);
      }
    },
  });

  return (
    <RawWrapper
      leftChildren={
        <>
          <BrandLogo detailed={false} size={48} />
          <div className="flex flex-col gap-4xl">
            <div className="flex flex-col gap-3xl">
              <div className="text-text-default heading4xl">
                Integrate Cloud Provider
              </div>
              <div className="text-text-default bodyMd">
                Kloudlite will help you to develop and deploy cloud native
                applications easily.
              </div>
            </div>
            <div className="flex flex-row gap-md items-center">
              <Badge>
                <span className="text-text-strong">Team:</span>
                <span className="bodySm-semibold text-text-default">
                  {account.displayName || parseName(account)}
                </span>
              </Badge>
            </div>
            <ProgressTracker
              items={[
                { label: 'Create Team', active: true, id: 1 },
                { label: 'Invite your Team Members', active: true, id: 2 },
                { label: 'Add your Cloud Provider', active: true, id: 3 },
                { label: 'Setup First Cluster', active: false, id: 4 },
                { label: 'Create your project', active: false, id: 5 },
              ]}
            />
          </div>
          <Button variant="outline" content="Skip" size="lg" />
        </>
      }
      rightChildren={
        <form
          onSubmit={handleSubmit}
          className="flex flex-col gap-3xl justify-center"
        >
          <div className="text-text-soft headingLg">Cloud provider details</div>
          <div className="flex flex-col gap-3xl">
            <TextInput
              label="Name"
              size="lg"
              value={values.displayName}
              onChange={handleChange('displayName')}
              error={!!errors.displayName}
              message={errors.displayName}
            />
            <IdSelector
              resType="providersecret"
              name={values.displayName}
              onChange={(v) => handleChange('name')(dummyEvent(v))}
            />

            <Select.Root
              label="Provider"
              error={!!errors.provider}
              message={errors.provider}
              value={values.provider}
              onChange={(provider: string) => {
                handleChange('provider')(dummyEvent(provider));
              }}
            >
              <Select.Option value="aws">Amazon Web Services</Select.Option>
            </Select.Root>

            <PasswordInput
              name="accessKey"
              label="Access Key ID"
              size="lg"
              onChange={handleChange('accessKey')}
              error={!!errors.accessKey}
              message={errors.accessKey}
              value={values.accessKey}
            />
            <PasswordInput
              name="accessSecret"
              label="Access Key Secret"
              size="lg"
              onChange={handleChange('accessSecret')}
              error={!!errors.accessSecret}
              message={errors.accessSecret}
              value={values.accessSecret}
            />
          </div>
          <div className="flex flex-row gap-xl justify-end">
            <Button
              variant="outline"
              content="Back"
              prefix={<ArrowLeft />}
              size="lg"
            />
            <Button
              loading={isLoading}
              type="submit"
              variant="primary"
              content="Continue"
              suffix={<ArrowRight />}
              size="lg"
            />
          </div>
        </form>
      }
    />
  );
};

export default NewCloudProvider;
