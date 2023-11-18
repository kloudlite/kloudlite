import { ArrowLeft, ArrowRight, Plus, X } from '@jengaicons/react';
import { useNavigate, useParams } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { Button, IconButton } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import SelectPrimitive from '~/components/atoms/select-primitive';
import { usePagination } from '~/components/molecule/pagination';
import { toast } from '~/components/molecule/toast';
import { titleCase, useMapper } from '~/components/utils';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { Kloudlite_Io__Apps__Iam__Types_Role as Role } from '~/root/src/generated/gql/server';
import { ListBody, ListItem } from '../components/console-list-components';
import DynamicPagination from '../components/dynamic-pagination';
import List from '../components/list';
import RawWrapper, { TitleBox } from '../components/raw-wrapper';
import { useConsoleApi } from '../server/gql/api-provider';
import { ACCOUNT_ROLES } from '../utils/commons';

const progressItems = [
  { label: 'Create Team', active: true, id: 1, completed: false },
  {
    label: 'Invite your Team Members',
    active: true,
    id: 2,
    completed: false,
  },
  {
    label: 'Add your Cloud Provider',
    active: false,
    id: 3,
    completed: false,
  },
  {
    label: 'Setup First Cluster',
    active: false,
    id: 4,
    completed: false,
  },
  {
    label: 'Create your project',
    active: false,
    id: 5,
    completed: false,
  },
];

const InviteTeam = () => {
  const { a: accountName } = useParams();

  const api = useConsoleApi();

  const navigate = useNavigate();

  const [inviting, setInviting] = useState(false);

  const [inviteMembers, setInviteMembers] = useState<
    { userEmail: string; userRole: Role }[]
  >([]);

  const items = useMapper(progressItems, (i) => {
    return {
      value: i.id,
      item: {
        ...i,
      },
    };
  });

  const { values, errors, handleChange, handleSubmit, resetValues } = useForm({
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
        { userEmail: values.userEmail, userRole: values.userRole as Role },
      ]);
      resetValues();
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
          invitation: inviteMembers,
        });
        if (e) {
          throw e[0];
        }
        toast.success('user invited');
        navigate(`/onboarding/${accountName}/new-cloud-provider`);
      } catch (err) {
        handleError(err);
      } finally {
        setInviting(false);
      }
    } else {
      navigate(`/onboarding/${accountName}/new-cloud-provider`);
    }
  };

  return (
    <RawWrapper
      title="Collaborate, Invite, Achieve Together!"
      subtitle="Simplify Collaboration and Enhance Productivity with Kloudlite
    teams."
      progressItems={items}
      rightChildren={
        <div className="flex flex-col p-3xl gap-6xl justify-center">
          <form onSubmit={handleSubmit} className="flex flex-col gap-3xl">
            <TitleBox
              title="Invite teammates"
              subtitle="Invite teammates to collaborate and contribute."
            />
            <div className="flex flex-col gap-xl">
              <div className="flex gap-2xl">
                <div className="flex-1">
                  <TextInput
                    label="Email"
                    value={values.userEmail}
                    onChange={handleChange('userEmail')}
                    error={!!errors.userEmail}
                    message={titleCase(errors.userEmail || '')}
                  />
                </div>

                <SelectPrimitive.Root
                  label="Role"
                  value={values.userRole}
                  onChange={handleChange('userRole')}
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
              header: (
                <div className="bodyMd-medium py-lg px-2xl">Team list</div>
              ),
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
          <div className="flex flex-row gap-xl justify-end">
            <Button
              variant="outline"
              content="Back"
              prefix={<ArrowLeft />}
              size="lg"
            />
            <Button
              variant="primary"
              content="Continue"
              suffix={<ArrowRight />}
              size="lg"
              loading={inviting}
              onClick={sendInvitation}
            />
          </div>
        </div>
      }
    />
  );
};

export default InviteTeam;
