import { Checkbox } from '~/components/atoms/checkbox';
import { generateKey } from '~/components/utils';
import List from '~/console/components/list';

const NOTIFICATION_CATEGORY = ['Web', 'Email'];
const NOTIFICATION_SETTINGS = [
  [
    {
      title: 'Account',
      field: { checked: false, label: 'Team join requests' },
    },
    {
      title: 'Deployments',
      field: { checked: false, label: 'Deployment failures' },
    },
  ],
  [
    {
      title: 'Integrations',
      field: { checked: false, label: 'Integration updates' },
    },
    { title: 'Usage', field: { checked: false, label: 'Warnings' } },
  ],
];

const ProfileNotification = () => {
  return (
    <div className="flex flex-col gap-6xl">
      <div className="flex flex-col gap-md">
        <div className="headingMd text-text-default">Notify me about...</div>
        <div className="bodyMd text-text-soft">
          Manage your personal notification settings
        </div>
      </div>
      {NOTIFICATION_CATEGORY.map((nc) => (
        <List.Root
          key={generateKey(nc)}
          plain
          className="rounded border border-border-default"
          header={<span className="headingMd text-text-default">{nc}</span>}
        >
          {NOTIFICATION_SETTINGS.map((ns, index) => (
            <List.Row
              plain
              className="p-3xl pb-lg pt-xl first:pt-3xl last:pb-3xl"
              key={generateKey(index)}
              columns={ns.map((nss) => ({
                className: 'flex-1',
                key: generateKey('account', index, nss.title, nss.field.label),
                render: () => (
                  <div className="flex flex-col gap-lg">
                    <div className="bodySm-medium text-text-soft">
                      {nss.title}
                    </div>
                    <div>
                      <Checkbox
                        withBounceEffect={false}
                        label={nss.field.label}
                        checked={nss.field.checked}
                      />
                    </div>
                  </div>
                ),
              }))}
            />
          ))}
        </List.Root>
      ))}
    </div>
  );
};

export default ProfileNotification;
