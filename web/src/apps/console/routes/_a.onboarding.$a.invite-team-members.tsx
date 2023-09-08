import { Button } from '~/components/atoms/button';
import { TextArea } from '~/components/atoms/input';
import { ArrowLeft, ArrowRight, Link } from '@jengaicons/react';
import { useParams, Link as L } from '@remix-run/react';
import { useMapper } from '~/components/utils';
import RawWrapper from '../components/raw-wrapper';

const InviteTeam = () => {
  const { a: accountName } = useParams();
  // @ts-ignore
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

  const items = useMapper(progressItems, (i) => {
    return {
      value: i.id,
      item: {
        ...i,
      },
    };
  });
  return (
    <RawWrapper
      title="Collaborate, Invite, Achieve Together!"
      subtitle="Simplify Collaboration and Enhance Productivity with Kloudlite
    teams."
      progressItems={items}
      rightChildren={
        <div className="flex flex-col p-3xl gap-6xl justify-center h-[549px]">
          <div className="flex flex-col gap-3xl">
            <div className="flex flex-row items-center justify-between">
              <div className="text-text-default headingXl">
                Invite teammates
              </div>
              <Button
                variant="primary-plain"
                content="Copy invite link"
                prefix={<Link />}
              />
            </div>
            <TextArea
              value=""
              rows="5"
              resize={false}
              placeholder="Ex. ellis@gmail.com, maria@gmail.com"
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
