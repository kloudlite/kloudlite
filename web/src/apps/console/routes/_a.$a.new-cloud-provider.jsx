import { Button } from '~/components/atoms/button';
import { PasswordInput, TextArea, TextInput } from '~/components/atoms/input';
import { BrandLogo } from '~/components/branding/brand-logo';
import { ProgressTracker } from '~/components/organisms/progress-tracker';
import { ArrowLeft, ArrowRight, Link } from '@jengaicons/react';
import Select from '~/components/atoms/select';
import RawWrapper from '../components/raw-wrapper';
import { IdSelector } from '../components/id-selector';

const NewCloudProvider = () => {
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
        <div className="flex flex-col gap-3xl justify-center">
          <div className="text-text-soft headingLg">Cloud provider details</div>
          <div className="flex flex-col gap-3xl">
            <TextInput label="Name" value="" size="lg" />
            <IdSelector name="id" />
            <Select.Root value="" label="Provider" size="lg">
              <Select.Option value="aws">Amazon Web Services</Select.Option>
            </Select.Root>
            <PasswordInput
              name="accessKey"
              value=""
              label="Access Key ID"
              size="lg"
            />
            <PasswordInput
              name="accessSecret"
              label="Access Key Secret"
              value=""
              size="lg"
            />
          </div>
          <div className="flex flex-row gap-xl justify-end">
            <Button
              variant="outline"
              content="Back"
              prefix={ArrowLeft}
              size="lg"
            />
            <Button
              variant="primary"
              content="Continue"
              suffix={ArrowRight}
              size="lg"
            />
          </div>
        </div>
      }
    />
  );
};

export default NewCloudProvider;
