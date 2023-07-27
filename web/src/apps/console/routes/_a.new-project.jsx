import { ArrowLeftFill, CircleDashed, Info, Search } from '@jengaicons/react';
import { Link } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { RadioGroup } from '~/components/atoms/radio';
import { ContextualSaveBar } from '~/components/organisms/contextual-save-bar.jsx';
import { ProgressTracker } from '~/components/organisms/progress-tracker';
import { cn } from '~/components/utils';

const NewProject = () => {
  const [clusters, _setClusters] = useState([
    {
      label: 'Plaxonic',
      time: '. 197d ago',
    },
    {
      label: 'Plaxonic',
      time: '. 197d ago',
    },
    {
      label: 'Plaxonic',
      time: '. 197d ago',
    },
    {
      label: 'Plaxonic',
      time: '. 197d ago',
    },
    {
      label: 'Plaxonic',
      time: '. 197d ago',
    },
    {
      label: 'Plaxonic',
      time: '. 197d ago',
    },
  ]);

  return (
    <div>
      <ContextualSaveBar message="Unsaved changes" fixed />
      <div
        className={cn(
          'transition-all flex justify-between',
          'flex-col md:flex-row',
          'py-6xl md:py-8xl',
          'mx-auto max-w-8xl w-full',
          'px-3xl md:px-0',
          'gap-3xl md:gap-6xl lg:gap-9xl xl:gap-10xl'
        )}
      >
        <div className="flex flex-col gap-3xl items-start">
          <Button
            content="Back"
            prefix={ArrowLeftFill}
            variant="plain"
            href="/project"
            LinkComponent={Link}
          />
          <span className="heading2xl text-text-default">
            Let’s create new project.
          </span>
          <ProgressTracker
            items={[
              {
                label: 'Configure projects',
                active: true,
                key: 'configureprojects',
              },
              {
                label: 'Publish',
                active: false,
                key: 'publish',
              },
            ]}
          />
        </div>
        <div className="flex flex-col border border-border-default bg-surface-basic-default shadow-card rounded-md flex-1">
          <div className="bg-surface-basic-subdued p-3xl text-text-default headingXl rounded-t-md">
            Configure Projects
          </div>
          <div className="flex flex-col gap-5xl px-3xl pt-3xl pb-5xl">
            <div className="flex flex-col md:flex-row gap-3xl">
              <div className="flex-1">
                <TextInput label="Project Name" placeholder="" />
              </div>
              <div className="flex-1">
                <TextInput
                  label="Project ID"
                  suffixIcon={Info}
                  placeholder=""
                />
              </div>
            </div>
            <div className="flex flex-col border border-border-disabled bg-surface-basic-default rounded-md">
              <TextInput
                prefixIcon={Search}
                placeholder="Cluster(s)"
                className="bg-surface-basic-subdued rounded-none rounded-t-md border-0 border-b border-border-disabled"
              />
              <RadioGroup
                className="flex flex-col pr-2xl !gap-y-0"
                labelPlacement="left"
                items={clusters.map((child, index) => {
                  return {
                    label: (
                      <div
                        className="p-2xl pl-lg flex flex-row gap-lg items-center"
                        key={index}
                      >
                        <CircleDashed size={24} />
                        <div className="flex flex-row flex-1 items-center gap-lg">
                          <span className="headingMd text-text-default">
                            Plaxonic
                          </span>
                          <span className="bodyMd text-text-default ">
                            . AWS, India
                          </span>
                        </div>
                      </div>
                    ),
                    value: `${index}`,
                    className: 'w-full justify-between',
                  };
                })}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default NewProject;
