import { ArrowLeft, ArrowRight } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { BrandLogo } from '~/components/branding/brand-logo';
import { ProgressTracker } from '~/components/organisms/progress-tracker';
import * as AlertDialog from '~/components/molecule/alert-dialog';
import { useState } from 'react';
import * as SelectInput from '~/components/atoms/select';

const NewCluster = () => {
  const [showUnsavedChanges, setShowUnsavedChanges] = useState(false);
  return (
    <div className="h-full flex flex-row">
      <div className="h-full w-[571px] flex flex-col bg-surface-basic-subdued py-11xl px-10xl">
        <div className="flex flex-col gap-8xl">
          <div className="flex flex-col gap-4xl items-start">
            <BrandLogo detailed={false} size={48} />
            <div className="flex flex-col gap-3xl">
              <div className="text-text-default heading4xl">
                Let’s create new project.
              </div>
              <div className="text-text-default bodyLg">
                Create your project to production effortlessly
              </div>
            </div>
          </div>
          <ProgressTracker
            items={[
              { label: 'Configure projects', active: true, id: 1 },
              { label: 'Publish', active: false, id: 2 },
            ]}
          />
          <Button
            variant="outline"
            content="Back"
            prefix={ArrowLeft}
            onClick={() => setShowUnsavedChanges(true)}
          />
        </div>
      </div>
      <div className="py-11xl px-10xl flex-1">
        <div className="flex flex-col gap-4xl">
          <div className="h-7xl" />
          <div className="flex flex-col gap-3xl p-3xl">
            <TextInput label="Cluster name" />
            <TextInput label="Cluster ID" />
            <SelectInput.Select label="Provider" value="aws">
              <SelectInput.Option>AWS</SelectInput.Option>
            </SelectInput.Select>
            <SelectInput.Select label="Region" value="india">
              <SelectInput.Option>India</SelectInput.Option>
            </SelectInput.Select>
          </div>
        </div>
        <div className="flex flex-row justify-end px-3xl">
          <Button variant="primary" content="Publish" suffix={ArrowRight} />
        </div>
      </div>

      {/* Unsaved change alert dialog */}
      <AlertDialog.Dialog
        show={showUnsavedChanges}
        onOpenChange={setShowUnsavedChanges}
      >
        <AlertDialog.Header>
          Leave page with unsaved changes?
        </AlertDialog.Header>
        <AlertDialog.Content>
          Leaving this page will delete all unsaved changes.
        </AlertDialog.Content>
        <AlertDialog.Footer>
          <AlertDialog.Button variant="basic" content="Cancel" />
          <AlertDialog.Button variant="critical" content="Delete" />
        </AlertDialog.Footer>
      </AlertDialog.Dialog>
    </div>
  );
};

export default NewCluster;
