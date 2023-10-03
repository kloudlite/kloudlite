import { ArrowLeft, ArrowRight, Plus, X } from '@jengaicons/react';
import { Link as L, useParams } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { Button, IconButton } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import SelectPrimitive from '~/components/atoms/select-primitive';
import { usePagination } from '~/components/molecule/pagination';
import { titleCase, useMapper } from '~/components/utils';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { ListBody, ListItem } from '../components/console-list-components';
import DynamicPagination from '../components/dynamic-pagination';
import List from '../components/list';
import RawWrapper, { TitleBox } from '../components/raw-wrapper';
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

  const [inviteMembers, setInviteMembers] = useState<
    {
      email: string;
      role: string;
    }[]
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
      email: '',
      role: 'account-member',
    },
    validationSchema: Yup.object({
      email: Yup.string()
        .required()
        .email()
        .test('is-valid', 'Email already exists.', (value) => {
          return !inviteMembers.find(
            (im) => im.email.toLowerCase() === value.toLowerCase()
          );
        }),
      role: Yup.string().required().oneOf(Object.keys(ACCOUNT_ROLES)),
    }),
    onSubmit: async () => {
      setInviteMembers((prev) => [
        ...prev,
        { email: values.email, role: values.role },
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
                    value={values.email}
                    onChange={handleChange('email')}
                    error={!!errors.email}
                    message={titleCase(errors.email || '')}
                  />
                </div>

                <SelectPrimitive.Root
                  label="Role"
                  value={values.role}
                  onChange={handleChange('role')}
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
                    key={item.email}
                    plain
                    className="p-lg px-xl [&:not(:last-child)]:border-b border-border-default"
                    columns={[
                      {
                        key: 1,
                        className: 'flex-1',
                        render: () => <ListItem data={item.email} />,
                      },
                      {
                        key: 2,
                        render: () => <ListBody data={item.role} />,
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
              to={`/onboarding/${accountName}/new-cloud-provider`}
              LinkComponent={L}
              variant="primary"
              content="Continue"
              suffix={<ArrowRight />}
              size="lg"
            />
          </div>
        </div>
      }
    />
  );
};

export default InviteTeam;
