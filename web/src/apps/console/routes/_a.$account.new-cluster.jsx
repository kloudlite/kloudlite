import { ArrowLeft, ArrowRight, PencilLine } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { BrandLogo } from '~/components/branding/brand-logo';
import { ProgressTracker } from '~/components/organisms/progress-tracker';
import * as AlertDialog from '~/components/molecule/alert-dialog';
import { useEffect, useState } from 'react';
import { useParams } from '@remix-run/react';
import * as SelectInput from '~/components/atoms/select';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { toast } from 'react-toastify';
import { getCookie } from '~/root/lib/app-setup/cookies';
import * as Tooltip from '~/components/atoms/tooltip';
import { IdSelector } from '../components/id-selector';

const NewCluster = () => {
  const [showUnsavedChanges, setShowUnsavedChanges] = useState(false);

  const cookie = getCookie();
  const { account } = useParams();

  const { values, handleSubmit, handleChange } = useForm({
    initialValues: {
      provider: '',
      region: 'ap-south-1',
      displayName: '',
      name: '',
    },
    validationSchema: Yup.object({}),
    onSubmit: async () => {
      try {
        console.log(values);
      } catch (err) {
        toast.error(err.message);
      }
    },
  });

  useEffect(() => {
    if (account) {
      cookie.set('kloudlite-account', account);
    }
  }, []);

  return (
    <Tooltip.TooltipProvider>
      <div className="h-full flex flex-row">
        <div className="h-full w-[571px] flex flex-col bg-surface-basic-subdued py-11xl px-10xl">
          <div className="flex flex-col gap-8xl">
            <div className="flex flex-col gap-4xl items-start">
              <BrandLogo detailed={false} size={48} />
              <div className="flex flex-col gap-3xl">
                <div className="text-text-default heading4xl">
                  Let’s create new cluster.
                </div>
                <div className="text-text-default bodyLg">
                  Create your cluster to production effortlessly
                </div>
              </div>
            </div>
            <ProgressTracker
              items={[
                { label: 'Configure cluster', active: true, id: 1 },
                { label: 'Review', active: false, id: 2 },
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
        <form className="py-11xl px-10xl flex-1" onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4xl">
            <div className="h-7xl" />
            <div className="flex flex-col gap-3xl p-3xl">
              <TextInput
                label="Cluster name"
                name="name"
                onChange={handleChange('name')}
                value={values.name}
              />
              <IdSelector
                name={values.name}
                onChange={(v) => {
                  console.log('hello', v);
                  handleChange('clusterId')({ target: { value: v } });
                }}
              />

              <SelectInput.Select
                label="Provider"
                value={values.provider}
                onChange={(v) => {
                  handleChange('provider')({ target: { value: v } });
                }}
              >
                <SelectInput.Option value="aws">
                  Amazon Web Services
                </SelectInput.Option>
              </SelectInput.Select>
              <SelectInput.Select
                label="Region"
                value={values.region}
                onChange={(v) => {
                  handleChange('region')({ target: { value: v } });
                }}
              >
                <SelectInput.Option value="ap-south-1">
                  Mumbai(ap-south-1)
                </SelectInput.Option>
              </SelectInput.Select>
            </div>
          </div>
          <div className="flex flex-row justify-end px-3xl">
            <Button
              variant="primary"
              content="Create"
              suffix={ArrowRight}
              type="submit"
            />
          </div>
        </form>

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
    </Tooltip.TooltipProvider>
  );
};

export default NewCluster;
