import { useState } from 'react';
import { ContextualSaveBar } from '~/components/organisms/contextual-save-bar.jsx';

const NewProject = () => {
  const [_clusters, _setClusters] = useState([
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
      {/* <ContextualSaveBar fullwidth={true} message={"Unsaved changes"} fixed />
            <div className="flex flex-row justify-between gap-[91px] pt-16">
                <div className="flex flex-col gap-5 items-start">
                    <Button content="Back" prefix={ArrowLeftFill} variant="plain" href={"/projects"} LinkComponent={Link} />
                    <span className="heading2xl text-text-default">
                        Let’s create new project.
                    </span>
                    <ProgressTracker items={[
                        {
                            label: "Configure projects",
                            active: true,
                            key: "configureprojects"
                        },
                        {
                            label: "Publish",
                            active: false,
                            key: "publish"
                        }
                    ]} />
                </div>
                <div className="flex flex-col border border-border-default bg-surface-basic-default shadow-card rounded-md flex-1">
                    <div className="bg-surface-basic-subdued p-5 text-text-default headingXl rounded-t-md">
                        Configure Projects
                    </div>
                    <div className="flex flex-col gap-8 px-5 pt-5 pb-8">
                        <div className="flex flex-row gap-5">
                            <TextInput label={"Project Name"} className={"flex-1"} placeholder={""} />
                            <TextInput label={"Project ID"} suffix={Info} className={"flex-1"} placeholder={""} />
                        </div>
                        <div className="flex flex-col border border-border-disabled bg-surface-basic-default rounded-md">
                            <div className="bg-surface-basic-subdued py-2 px-4 text-text-default headingMd rounded-t-md">
                                Cluster(s)
                            </div>
                            <div className="flex flex-col">
                                {clusters.map((child, index) => {
                                    return (
                                        <div className="p-4 flex flex-row gap-2.5 items-center" key={index}>
                                            <CircleDashed />
                                            <div className="flex flex-row flex-1 items-center gap-2">
                                                <span className="headingMd text-text-default">Plaxonic</span>
                                                <span className="bodyMd text-text-default">. 197d ago</span>
                                            </div>
                                            <Checkbox label="" />
                                        </div>
                                    )
                                })}
                            </div>
                        </div>
                    </div>
                </div>
            </div> */}
      <ContextualSaveBar fullwidth message="Unsaved changes" fixed />
      new project
    </div>
  );
};

export default NewProject;
