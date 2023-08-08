import { ArrowLeft, ArrowRight, CircleDashed, Search } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { BrandLogo } from '~/components/branding/brand-logo';
import { ProgressTracker } from '~/components/organisms/progress-tracker';
import * as AlertDialog from '~/components/molecule/alert-dialog';
import { useState } from 'react';
import { useParams, useLoaderData } from '@remix-run/react';
import * as Radio from '~/components/atoms/radio';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { getCookie } from '~/root/lib/app-setup/cookies';
import * as Tooltip from '~/components/atoms/tooltip';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { toast } from '~/components/molecule/toast';
import logger from '~/root/lib/client/helpers/log';
import { dayjs } from '~/components/molecule/dayjs';
import { ensureAccountSet } from '../server/utils/auth-utils';
import {
  getPagination,
  getSearch,
  parseName,
  parseUpdationTime,
} from '../server/r-urils/common';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { IdSelector } from '../components/id-selector';

const NewProject = () => {
  const { clustersData } = useLoaderData();
  const clusters = clustersData?.edges?.map(({ node }) => node || []);

  const [showUnsavedChanges, setShowUnsavedChanges] = useState(false);

  const { values, handleSubmit, handleChange } = useForm({
    initialValues: {
      name: '',
      displayName: '',
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

  return (
    <Tooltip.TooltipProvider>
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
                { label: 'Configure project', active: true, id: 1 },
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
          <div className="gap-6xl flex flex-col p-3xl">
            <div className="flex flex-col gap-4xl">
              <div className="h-7xl" />
              <div className="flex flex-col gap-3xl">
                <TextInput
                  label="Project name"
                  name="name"
                  value={values.displayName}
                  onChange={handleChange('displayName')}
                />
                <IdSelector
                  name={values.displayName}
                  onChange={(v) => {
                    handleChange('name')(dummyEvent(v));
                  }}
                />
              </div>
            </div>
            <div className="flex flex-col border border-border-disabled bg-surface-basic-default rounded-md">
              <TextInput
                prefixIcon={Search}
                placeholder="Cluster(s)"
                className="bg-surface-basic-subdued rounded-none rounded-t-md border-0 border-b border-border-disabled"
              />
              <Radio.RadioGroup
                className="flex flex-col pr-2xl !gap-y-0"
                labelPlacement="left"
              >
                {clusters.map((cluster) => {
                  return (
                    <Radio.RadioItem
                      value={parseName(cluster)}
                      withBounceEffect={false}
                      className="justify-between w-full"
                      key={parseName(cluster)}
                    >
                      <div className="p-2xl pl-lg flex flex-row gap-lg items-center">
                        <CircleDashed size={24} />
                        <div className="flex flex-row flex-1 items-center gap-lg">
                          <span className="headingMd text-text-default">
                            {parseName(cluster)}
                          </span>
                          <span className="bodyMd text-text-default ">
                            {dayjs(parseUpdationTime(cluster)).fromNow()}
                          </span>
                        </div>
                      </div>
                    </Radio.RadioItem>
                  );
                })}
              </Radio.RadioGroup>
            </div>
            <div className="flex flex-row justify-end">
              <Button
                variant="primary"
                content="Create"
                suffix={ArrowRight}
                type="submit"
              />
            </div>
          </div>
        </form>

        {/* Unsaved change alert dialog */}
        <AlertDialog.DialogRoot
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
        </AlertDialog.DialogRoot>
      </div>
    </Tooltip.TooltipProvider>
  );
};

export const loader = async (ctx) => {
  ensureAccountSet(ctx);
  const { data, errors } = await GQLServerHandler(ctx.request).listClusters({
    pagination: getPagination(ctx),
    search: getSearch(ctx),
  });

  if (errors) {
    logger.error(errors);
  }

  return {
    clustersData: data,
  };
};

export default NewProject;
