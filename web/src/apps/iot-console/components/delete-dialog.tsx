import { ReactNode, useEffect, useState } from 'react';
import { TextInput } from '~/components/atoms/input';
import AlertModal from './alert-modal';

interface IDeleteDialog {
  setShow: (show: any) => void;
  onSubmit?: (val: any) => any;
  show: any;
  resourceType: string;
  resourceName?: string;
  customMessages?: {
    warning?: ReactNode;
    prompt?: ReactNode;
    action?: 'Delete' | string;
  };
}
const DeleteDialog = ({
  resourceName,
  resourceType,
  show,
  customMessages,
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
      title={`${customMessages?.action || 'Delete'} ${resourceType}`}
      variant="critical"
      message={
        <div className="flex flex-col gap-3xl">
          <div className="text-text-default bodyMd">
            {customMessages?.warning || (
              <span>
                Are you sure you want to delete &ldquo;{resourceName}&rdquo;?
              </span>
            )}
          </div>
          <div className="flex flex-col gap-md">
            <div className="flex flex-row items-center text-text-default">
              {customMessages?.prompt || (
                <>
                  <div className="bodyMd">Enter the {resourceType} name</div>
                  <div className="bodyMd-semibold">
                    &nbsp;{resourceName}&nbsp;
                  </div>
                  <div className="bodyMd">to continue:</div>
                </>
              )}
            </div>
            <TextInput
              value={inputName}
              onChange={({ target }) => {
                setInputName(target.value);
              }}
              disabled={loading}
              autoComplete="off"
            />
          </div>
        </div>
      }
      show={show}
      setShow={setShow}
      okText={customMessages?.action || 'Delete'}
      cancelText="Cancel"
      footer
      onSubmit={onSubmit}
      okDisabled={inputName !== resourceName}
      setLoading={setLoading}
    />
  );
};

export default DeleteDialog;
