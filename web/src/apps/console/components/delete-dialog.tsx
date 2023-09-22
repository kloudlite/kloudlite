import { useEffect, useState } from 'react';
import { TextInput } from '~/components/atoms/input';
import AlertModal from './alert-modal';

interface IDeleteDialog {
  setShow: (show: any) => void;
  onSubmit?: (val: any) => any;
  show: any;
  resourceType: string;
  resourceName?: string;
}
const DeleteDialog = ({
  resourceName,
  resourceType,
  show,
  setShow,
  onSubmit,
}: IDeleteDialog) => {
  const [inputName, setInputName] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setInputName('');
  }, [show]);
  return (
    <AlertModal
      title={`Delete ${resourceType}`}
      variant="critical"
      message={
        <div className="flex flex-col gap-3xl">
          <div className="text-text-default bodyMd">
            Are you sure you want to delete &ldquo;{resourceName}&rdquo;?
          </div>
          <div className="flex flex-col gap-md">
            <div className="flex flex-row items-center text-text-default">
              <div className="bodyMd">Enter the {resourceType} name</div>
              <div className="bodyMd-semibold">&nbsp;{resourceName}&nbsp;</div>
              <div className="bodyMd">to continue:</div>
            </div>
            <TextInput
              value={inputName}
              onChange={({ target }) => {
                setInputName(target.value);
              }}
              disabled={loading}
            />
          </div>
        </div>
      }
      show={show}
      setShow={setShow}
      okText="Delete"
      cancelText="Cancel"
      footer
      onSubmit={onSubmit}
      okDisabled={inputName !== resourceName}
      submitType="submit"
      setLoading={setLoading}
    />
  );
};

export default DeleteDialog;
