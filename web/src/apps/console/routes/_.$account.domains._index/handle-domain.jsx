import * as Popup from '~/components/molecule/popup';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { toast } from '~/components/molecule/toast';
import { useEffect, useState } from 'react';
import { TextInput } from '~/components/atoms/input';
import { Search } from '@jengaicons/react';
import { dummyData } from '~/console/dummy/data';
import { ResourceProjectItem } from './resources';
import ResourceList from './resource-list-projects';

const HandleDomain = ({ show, setShow }) => {
  const [data, setData] = useState(dummyData.projectList);

  const [selectedProject, setSelectedProject] = useState(null);

  const [validationSchema, setValidationSchema] = useState(Yup.object({}));

  const {
    values,
    errors,
    handleSubmit,
    handleChange,
    isLoading,
    resetValues,
    setValues,
  } = useForm({
    initialValues: {},
    validationSchema,

    onSubmit: async (val) => {
      try {
        if (show.type === 'add') {
          //
        } else {
          //
        }
      } catch (err) {
        toast.error(err.message);
      }
    },
  });

  useEffect(() => {
    if (show?.type === 'edit') {
      setValues({});
      setValidationSchema(Yup.object({}));
    }
  }, [show]);

  return (
    <Popup.PopupRoot
      show={show}
      onOpenChange={(e) => {
        if (!e) {
          resetValues();
        }
        setShow(e);
      }}
    >
      <Popup.Header>
        {show.type === 'add' ? 'Add new domain' : 'Edit domain'}
      </Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-3xl">
            <div className="bodyMd text-text-default">
              Select a project to add your domain to:
            </div>
            <div className="rounded-lg flex flex-col border border-border-default">
              <TextInput
                prefixIcon={Search}
                placeholder="Search"
                className="rounded-none rounded-t-lg border-0 border-b border-border-default z-10"
              />
              <div className="overflow-hidden rounded-b-lg">
                <div className="h-14xl overflow-y-scroll py-sm pb-[6px]">
                  <ResourceList>
                    {data.map((p) => (
                      <ResourceList.ResourceItem key={p.id} textValue={p.id}>
                        <ResourceProjectItem
                          item={p}
                          // selected={}
                          onSelect={() => {
                            setSelectedProject(p);
                          }}
                        />
                      </ResourceList.ResourceItem>
                    ))}
                  </ResourceList>
                </div>
              </div>
            </div>
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button content="Cancel" variant="basic" closable />
          <Popup.Button
            loading={isLoading}
            type="submit"
            disabled={!data.find((d) => d.selected)}
            content={show.type === 'add' ? 'Continue' : 'Update'}
            variant="primary"
          />
        </Popup.Footer>
      </form>
    </Popup.PopupRoot>
  );
};

export default HandleDomain;
