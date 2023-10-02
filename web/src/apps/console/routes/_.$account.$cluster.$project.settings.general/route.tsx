import { CopySimple } from '@jengaicons/react';
import { useOutletContext } from '@remix-run/react';
import { TextInput } from '~/components/atoms/input';
import { toast } from '~/components/molecule/toast';
import {
  Box,
  DeleteContainer,
} from '~/console/components/common-console-components';
import { parseName } from '~/console/server/r-utils/common';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import { consoleBaseUrl } from '~/root/lib/configs/base-url.cjs';
import { IProjectContext } from '../_.$account.$cluster.$project';

const ProjectSettingGeneral = () => {
  const { project, account, cluster } = useOutletContext<IProjectContext>();

  const { copy } = useClipboard({
    onSuccess() {
      toast.success('Text copied to clipboard.');
    },
  });

  return (
    <>
      <Box title="General">
        <TextInput value={project.displayName} label="Project name" />
        <div className="flex flex-row items-center gap-3xl">
          <div className="flex-1">
            <TextInput
              label="Project URL"
              value={`${consoleBaseUrl}/${parseName(account)}/${parseName(
                cluster
              )}/${parseName(project)}`}
              message="This is your URL namespace within Kloudlite"
              disabled
              suffix={
                <div className="flex justify-center items-center" title="Copy">
                  <button
                    onClick={() =>
                      copy(
                        `${consoleBaseUrl}/${parseName(account)}/${parseName(
                          cluster
                        )}/${parseName(project)}`
                      )
                    }
                    className="outline-none hover:bg-surface-basic-hovered active:bg-surface-basic-active rounded text-text-default"
                    tabIndex={-1}
                  >
                    <CopySimple size={16} />
                  </button>
                </div>
              }
            />
          </div>
          <div className="flex-1">
            <TextInput
              value={parseName(project)}
              label="Project ID"
              message="Used when interacting with the Kloudlite API"
              suffix={
                <div className="flex justify-center items-center" title="Copy">
                  <button
                    onClick={() => copy(parseName(project))}
                    className="outline-none hover:bg-surface-basic-hovered active:bg-surface-basic-active rounded text-text-default"
                    tabIndex={-1}
                  >
                    <CopySimple size={16} />
                  </button>
                </div>
              }
              disabled
            />
          </div>
        </div>
      </Box>

      <DeleteContainer title="Delete Project" action={() => {}}>
        Permanently remove your Project and all of its contents from the
        Kloudlite platform. This action is not reversible — please continue with
        caution.
      </DeleteContainer>
    </>
  );
};
export default ProjectSettingGeneral;
