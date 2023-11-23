import { IRemixCtx } from '~/root/lib/types/common';
import { ArrowLeft, ArrowRight, Check, UserCircle } from '@jengaicons/react';
import { useLoaderData, useNavigate, useOutletContext } from '@remix-run/react';
import { defer } from '@remix-run/node';
import { useMapper } from '~/components/utils';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { toast } from '~/components/molecule/toast';
import { handleError } from '~/root/lib/utils/common';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { date } from 'yup';
import { Badge } from '~/components/atoms/badge';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { ensureAccountSet } from '../server/utils/auth-utils';
import { LoadingComp, pWrapper } from '../components/loading-component';
import RawWrapper, { TitleBox } from '../components/raw-wrapper';
import { parseName } from '../server/r-utils/common';
import { IAccountContext } from './_.$account';
import { FadeIn } from './_.$account.$cluster.$project.$scope.$workspace.new-app/util';
import AlertModal from '../components/alert-modal';
import CodeView from '../components/code-view';
import { asyncPopupWindow } from '../utils/commons';
import { useConsoleApi } from '../server/gql/api-provider';
import { LoadingPlaceHolder } from '../components/loading';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { cloudprovider: cp } = ctx.params;
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).getProviderSecret({
      name: cp,
    });

    if (errors) {
      return { redirect: '/teams', cloudProvider: data };
    }

    return {
      cloudProvider: data,
      redirect: '',
    };
  });
  return defer({ promise });
};

const Validator = ({ cloudProvider }: { cloudProvider: any }) => {
  const { account } = useOutletContext<IAccountContext>();
  const [showUnsavedChanges, setShowUnsavedChanges] = useState(false);
  const navigate = useNavigate();

  const api = useConsoleApi();
  const checkAwsAccess = async () => {
    const { data, errors } = await api.checkAwsAccess({
      cloudproviderName: cloudProvider.metadata?.name || '',
    });
    if (errors) {
      throw errors[0];
    }
    return data;
  };

  const [isLoading, setIsLoading] = useState(false);

  const items = useMapper(
    [
      {
        label: 'Create Team',
        active: true,
        id: 1,
        completed: false,
      },
      {
        label: 'Invite your Team Members',
        active: true,
        id: 2,
        completed: false,
      },
      {
        label: 'Add your Cloud Provider',
        active: true,
        id: 3,
        completed: false,
      },
      {
        label: 'Validate Cloud Provider',
        active: true,
        id: 4,
        completed: false,
      },
      {
        label: 'Setup First Cluster',
        active: false,
        id: 5,
        completed: false,
      },
      // {
      //   label: 'Create your project',
      //   active: false,
      //   id: 5,
      //   completed: false,
      // },
    ],
    (i) => {
      return {
        value: i.id,
        item: {
          ...i,
        },
      };
    }
  );

  const { data, isLoading: il } = useCustomSwr(
    () => cloudProvider.metadata!.name + isLoading,
    async () => {
      if (!cloudProvider.metadata!.name) {
        throw new Error('Invalid cloud provider name');
      }
      return api.checkAwsAccess({
        cloudproviderName: cloudProvider.metadata.name,
      });
    }
  );

  return (
    <>
      <RawWrapper
        title={"Validate your cloud provider's credentials"}
        subtitle="Kloudlite will help you to develop and deploy cloud native applications easily."
        progressItems={items}
        badge={{
          title: account.displayName,
          subtitle: parseName(account),
          image: <UserCircle size={20} />,
        }}
        onCancel={() => setShowUnsavedChanges(true)}
        rightChildren={
          <FadeIn notForm>
            <TitleBox
              title="Create cloudformation stack for cloudprovider"
              subtitle="Kloudlite will help you to develop and deploy cloud native applications easily."
            />
            {/* eslint-disable-next-line no-nested-ternary */}
            {il ? (
              <div className="py-2xl">
                <LoadingPlaceHolder
                  title="Validating the cloudformation stack"
                  height={250}
                />
              </div>
            ) : data?.result ? (
              <div className="py-2xl">
                <Badge type="success" icon={<Check />}>
                  Your Credential is valid
                </Badge>
              </div>
            ) : (
              <div className="flex flex-col gap-3xl p-xl border border-border-default rounded">
                <div className="flex gap-xl items-center">
                  <span>Account ID</span>
                  <span className="bodyMd-semibold text-text-primary">
                    {cloudProvider.aws?.awsAccountId}
                  </span>
                </div>
                <div className="flex flex-col gap-2xl text-start">
                  <CodeView copy data={data?.installationUrl || ''} />

                  <span className="">
                    visit the link above or
                    <button
                      className="inline-block mx-lg text-text-primary hover:underline"
                      onClick={async () => {
                        setIsLoading(true);
                        try {
                          await asyncPopupWindow({
                            url: data?.installationUrl || '',
                          });

                          const res = await checkAwsAccess();

                          if (res.result) {
                            toast.success('Aws account validated successfully');
                          } else {
                            toast.error('Aws account validation failed');
                          }
                        } catch (err) {
                          handleError(err);
                        }

                        setIsLoading(false);
                      }}
                    >
                      click here
                    </button>
                    to create AWS cloudformation stack
                  </span>
                </div>
              </div>
            )}
            <div className="flex flex-row gap-xl justify-end">
              {/* <Button */}
              {/*   variant="outline" */}
              {/*   content="Back" */}
              {/*   prefix={<ArrowLeft />} */}
              {/*   size="lg" */}
              {/* /> */}
              <Button
                variant="primary"
                content={data?.result ? 'Next' : 'Skip'}
                suffix={<ArrowRight />}
                size="lg"
                onClick={() => {
                  navigate(`/${parseName(account)}/new-cluster`);
                }}
              />
            </div>
          </FadeIn>
        }
      />
      <AlertModal
        title="Leave page with unsaved changes?"
        message="Leaving this page will delete all unsaved changes."
        okText="Leave"
        cancelText="Stay"
        variant="critical"
        show={showUnsavedChanges}
        setShow={setShowUnsavedChanges}
        onSubmit={() => {
          setShowUnsavedChanges(false);
          navigate(`/${parseName(account)}/clusters`);
        }}
      />
    </>
  );
};

const _NewCluster = () => {
  const { promise } = useLoaderData<typeof loader>();

  return (
    <LoadingComp data={promise}>
      {({ cloudProvider }) => {
        if (cloudProvider === null) {
          return null;
        }
        return <Validator cloudProvider={cloudProvider} />;
      }}
    </LoadingComp>
  );
};

export default _NewCluster;
