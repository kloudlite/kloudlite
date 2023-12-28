import { ArrowRight, Plus, X } from '@jengaicons/react';
import { useNavigate, useParams } from '@remix-run/react';
import { Button, IconButton } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { toast } from '~/components/molecule/toast';
import { useDataFromMatches } from '~/root/lib/client/hooks/use-custom-matches';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import { UserMe } from '~/root/lib/server/gql/saved-queries';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { useEffect, useState } from 'react';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import RawWrapper, { TitleBox } from '~/console/components/raw-wrapper';
import { FadeIn } from '~/console/page-components/util';
import { IdSelector } from '~/console/components/id-selector';
import ProgressWrapper from '~/console/components/progress-wrapper';
import SelectPrimitive from '~/components/atoms/select-primitive';
import DynamicPagination from '~/console/components/dynamic-pagination';
import List from '~/console/components/list';
import {
  ListBody,
  ListItem,
} from '~/console/components/console-list-components';
import { usePagination } from '~/components/molecule/pagination';
import { ACCOUNT_ROLES } from '~/console/utils/commons';
import { Github__Com___Kloudlite___Api___Apps___Iam___Types__Role as Role } from '~/root/src/generated/gql/server';
import { titleCase } from '~/components/utils';

const InviteSection = () => {
  const { a: accountName } = useParams();
  const api = useConsoleApi();
  const [inviting, setInviting] = useState(false);

  const [inviteMembers, setInviteMembers] = useState<
    { userEmail: string; userRole: Role }[]
  >([]);

  const {
    values: valuesInvite,
    errors: errorsInvite,
    handleChange: handleChangeInvite,
    handleSubmit: handleSubmitInvite,
    resetValues: resetValuesInvite,
  } = useForm({
    initialValues: {
      userEmail: '',
      userRole: 'account_member',
    },
    validationSchema: Yup.object({
      userEmail: Yup.string()
        .required()
        .email()
        .test('is-valid', 'Email already exists.', (value) => {
          return !inviteMembers.find(
            (im) => im.userEmail.toLowerCase() === value.toLowerCase()
          );
        }),
      userRole: Yup.string().required().oneOf(Object.keys(ACCOUNT_ROLES)),
    }),
    onSubmit: async () => {
      setInviteMembers((prev) => [
        ...prev,
        {
          userEmail: valuesInvite.userEmail,
          userRole: valuesInvite.userRole as Role,
        },
      ]);
      resetValuesInvite();
    },
  });

  const { page, hasNext, hasPrevious, onNext, onPrev, setItems } =
    usePagination({
      items: inviteMembers,
      itemsPerPage: 5,
    });

  useEffect(() => {
    setItems(inviteMembers);
  }, [inviteMembers]);

  const removeMember = ({ item }: { item: (typeof inviteMembers)[number] }) => {
    setInviteMembers(inviteMembers.filter((im) => im !== item));
  };

  const sendInvitation = async () => {
    if (inviting) {
      return;
    }

    if (inviteMembers.length > 0) {
      try {
        setInviting(true);
        const { errors: e } = await api.inviteMembersForAccount({
          accountName: accountName!,
          invitations: inviteMembers,
        });
        if (e) {
          throw e[0];
        }
        toast.success('user invited');
        // navigate(`/onboarding/${accountName}/new-cloud-provider`);
      } catch (err) {
        handleError(err);
      } finally {
        setInviting(false);
      }
    } else {
      // navigate(`/onboarding/${accountName}/new-cloud-provider`);
    }
  };
  return (
    <div className="flex flex-col gap-6xl justify-center">
      <form onSubmit={handleSubmitInvite} className="flex flex-col gap-3xl">
        <div className="flex flex-col gap-xl">
          <div className="flex gap-2xl">
            <div className="flex-1">
              <TextInput
                label="Email"
                value={valuesInvite.userEmail}
                onChange={handleChangeInvite('userEmail')}
                error={!!errorsInvite.userEmail}
                message={titleCase(errorsInvite.userEmail || '')}
              />
            </div>

            <SelectPrimitive.Root
              label="Role"
              value={valuesInvite.userRole}
              onChange={handleChangeInvite('userRole')}
            >
              {Object.entries(ACCOUNT_ROLES).map(([key, value]) => {
                return (
                  <SelectPrimitive.Option key={key} value={key}>
                    {value}
                  </SelectPrimitive.Option>
                );
              })}
            </SelectPrimitive.Root>
          </div>
          <div>
            <Button
              content="Invite"
              variant="basic"
              prefix={<Plus />}
              size="sm"
              type="submit"
            />
          </div>
        </div>
      </form>
      <DynamicPagination
        {...{
          hasNext,
          hasPrevious,
          hasItems: inviteMembers.length > 0,
          noItemsMessage: '0 teammates to invite.',
          onNext,
          onPrev,
          headerClassName: 'bg-surface-basic-subdued',
          header: <div className="bodyMd-medium py-lg px-2xl">Team list</div>,
        }}
        className="rounded border border-border-default overflow-hidden min-h-[266px]"
      >
        <List.Root plain>
          {page.map((item) => {
            return (
              <List.Row
                key={item.userEmail}
                plain
                className="p-lg px-xl [&:not(:last-child)]:border-b border-border-default"
                columns={[
                  {
                    key: 1,
                    className: 'flex-1',
                    render: () => <ListItem data={item.userEmail} />,
                  },
                  {
                    key: 2,
                    render: () => <ListBody data={item.userRole} />,
                  },
                  {
                    key: 3,
                    render: () => (
                      <IconButton
                        icon={<X />}
                        variant="plain"
                        size="sm"
                        onClick={() => {
                          removeMember({ item });
                        }}
                      />
                    ),
                  },
                ]}
              />
            );
          })}
        </List.Root>
      </DynamicPagination>
    </div>
  );
};
const NewAccount = () => {
  const api = useConsoleApi();
  const navigate = useNavigate();
  const user = useDataFromMatches<UserMe>('user', {});
  const [isNameLoading, setIsNameLoading] = useState(false);
  const { values, handleChange, errors, isLoading, handleSubmit } = useForm({
    initialValues: {
      name: '',
      displayName: '',
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      displayName: Yup.string().required(),
    }),
    onSubmit: async (v) => {
      if (isNameLoading) {
        toast.error('id is being checked, please wait');
        return;
      }
      try {
        const { errors: _errors } = await api.createAccount({
          account: {
            metadata: { name: v.name },
            spec: {},
            displayName: v.displayName,
            contactEmail: user.email,
          },
        });
        if (_errors) {
          throw _errors[0];
        }
        toast.success('account created');
        navigate(`/onboarding/${v.name}/new-cloud-provider`);
      } catch (err) {
        handleError(err);
      }
    },
  });

  const progressItems = [
    {
      label: 'Create Team',
      active: true,
      completed: false,
      children: (
        <form className="py-3xl flex flex-col gap-3xl" onSubmit={handleSubmit}>
          <div className="flex flex-col">
            <TextInput
              size="lg"
              value={values.displayName}
              onChange={handleChange('displayName')}
              error={!!errors.displayName}
              message={errors.displayName}
              label="Name"
            />
            <IdSelector
              onLoad={(v) => setIsNameLoading(v)}
              name={values.displayName}
              onChange={(v) => handleChange('name')(dummyEvent(v))}
              resType="account"
              className="pt-2xl"
            />
          </div>
          {/* <InviteSection /> */}
          <div className="flex flex-row gap-xl justify-start">
            <Button
              variant="primary"
              content="Next"
              suffix={<ArrowRight />}
              size="md"
              loading={isLoading}
              type="submit"
            />
          </div>
        </form>
      ),
    },
    {
      label: 'Add your Cloud Provider',
      active: false,
      completed: false,
    },
    {
      label: 'Validate Cloud Provider',
      active: false,
      completed: false,
    },
    {
      label: 'Setup First Cluster',
      active: false,
      completed: false,
    },
  ];

  return (
    <ProgressWrapper
      title="Setup your account!"
      subTitle="Simplify Collaboration and Enhance Productivity with Kloudlite
  teams"
      progressItems={{
        items: progressItems,
      }}
    />
  );
};

export default NewAccount;
