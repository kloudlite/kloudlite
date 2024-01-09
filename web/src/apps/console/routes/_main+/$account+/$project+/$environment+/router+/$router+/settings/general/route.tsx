import { CopySimple } from '@jengaicons/react';
import { useNavigate, useOutletContext, useParams } from '@remix-run/react';
import { TextInput } from '~/components/atoms/input';
import { toast } from '~/components/molecule/toast';
import {
  Box,
  DeleteContainer,
} from '~/console/components/common-console-components';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { parseName } from '~/console/server/r-utils/common';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { IRouterConfig } from 'websocket';
import Wrapper from '~/console/components/wrapper';
import DeleteDialog from '~/console/components/delete-dialog';
import { useState } from 'react';
import { handleError } from '~/root/lib/utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { IProjectContext } from '../../../../../_layout';
import { IRouterContext } from '../../_layout';

// export const updateProject = async ({
//   api,
//   data,
// }: {
//   api: ConsoleApiType;
//   data: ExtractNodeType<IProject>;
// }) => {
//   try {
//     const { errors: e } = await api.updateProject({
//       project: {
//         displayName: data.displayName,
//         metadata: {
//           name: parseName(data),
//         },
//         spec: {
//           targetNamespace: data.spec.targetNamespace,
//         },
//       },
//     });
//     if (e) {
//       throw e[0];
//     }
//   } catch (err) {
//     handleError(err);
//   }
// };

const ProjectSettingGeneral = () => {
  const { router, environment, project } = useOutletContext<IRouterContext>();

  const { setHasChanges, resetAndReload } = useUnsavedChanges();
  const api = useConsoleApi();
  const reload = useReload();
  const navigate = useNavigate();
  const { account } = useParams();

  const [deleteRouter, setDeleteRouter] = useState(false);

  const { copy } = useClipboard({
    onSuccess() {
      toast.success('Text copied to clipboard.');
    },
  });

  // const { values, handleChange, submit, isLoading, resetValues } = useForm({
  //   initialValues: {
  //     displayName: project.displayName,
  //   },
  //   validationSchema: Yup.object({
  //     displayName: Yup.string().required('Name is required.'),
  //   }),
  //   onSubmit: async (val) => {
  //     await updateProject({
  //       api,
  //       data: { ...project, displayName: val.displayName },
  //     });
  //     resetAndReload();
  //   },
  // });

  // useEffect(() => {
  //   setHasChanges(values.displayName !== project.displayName);
  // }, [values]);

  // useEffect(() => {
  //   resetValues();
  // }, [project]);

  return (
    <div>
      <Wrapper secondaryHeader={{ title: 'General' }}>
        <Box title="Router detail">
          <TextInput label="Router name" value={router.displayName} />
          <div className="flex flex-row items-center gap-3xl">
            <div className="flex-1">
              <TextInput
                label="Router ID"
                value={parseName(router)}
                message="Used when interacting with the Kloudlite API"
                suffix={
                  <div
                    className="flex justify-center items-center"
                    title="Copy"
                  >
                    <button
                      aria-label="copy"
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

        <DeleteContainer
          title="Delete router"
          action={async () => {
            setDeleteRouter(true);
          }}
        >
          Permanently remove your router and all of its contents from the “
          {router.displayName}” router. This action is not reversible, so please
          continue with caution.
        </DeleteContainer>
        <DeleteDialog
          resourceName={parseName(router)}
          resourceType="router"
          show={deleteRouter}
          setShow={setDeleteRouter}
          onSubmit={async () => {
            try {
              const { errors } = await api.deleteRouter({
                envName: parseName(environment),
                projectName: parseName(project),
                routerName: parseName(router),
              });

              if (errors) {
                throw errors[0];
              }
              reload();
              toast.success(`Router deleted successfully`);
              setDeleteRouter(false);
              navigate(
                `/${account}/${parseName(project)}/${parseName(
                  environment
                )}/routers/`
              );
            } catch (err) {
              handleError(err);
            }
          }}
        />
      </Wrapper>
    </div>
  );
};
export default ProjectSettingGeneral;
