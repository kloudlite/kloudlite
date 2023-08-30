import { Button } from '~/components/atoms/button';
import { TextArea } from '~/components/atoms/input';
import { BrandLogo } from '~/components/branding/brand-logo';
import { ProgressTracker } from '~/components/organisms/progress-tracker';
import { ArrowLeft, ArrowRight, Link } from '@jengaicons/react';
import { useParams, Link as L, useOutletContext } from '@remix-run/react';
import { Badge } from '~/components/atoms/badge';
import RawWrapper from '../components/raw-wrapper';

const InviteTeam = () => {
  const { a: accountName } = useParams();
  // @ts-ignore
  const { account: team } = useOutletContext();

  return (
    <RawWrapper
      leftChildren={
        <>
          <BrandLogo detailed={false} size={48} />
          <div className="flex flex-col gap-4xl">
            <div className="flex flex-col gap-3xl">
              <div className="text-text-default heading4xl">
                Collaborate, Invite, Achieve Together!
              </div>
              <div className="text-text-default bodyMd">
                Simplify Collaboration and Enhance Productivity with Kloudlite
                teams.
              </div>
              <div className="flex flex-row gap-md items-center">
                <Badge>
                  <span className="text-text-strong">Team:</span>
                  <span className="bodySm-semibold text-text-default">
                    {team.displayName || team.name}
                  </span>
                </Badge>
              </div>
            </div>

            <ProgressTracker
              items={[
                { label: 'Create Team', active: true, id: 1 },
                { label: 'Invite your Team Members', active: true, id: 2 },
                { label: 'Add your Cloud Provider', active: false, id: 3 },
                { label: 'Setup First Cluster', active: false, id: 4 },
                { label: 'Create your project', active: false, id: 5 },
              ]}
            />
          </div>
          <Button variant="outline" content="Skip" size="lg" />
        </>
      }
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
