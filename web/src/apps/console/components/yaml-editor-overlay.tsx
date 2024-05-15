import OverlaySideDialog from './overlay-side-dialog';
import YamlEditor from '../page-components/yaml-editor/yaml-editor';

const YamlEditorOverlay = ({
  item,
  showDialog,
  setShowDialog,
}: {
  item: any;
  showDialog: boolean;
  setShowDialog: React.Dispatch<React.SetStateAction<boolean>>;
}) => {
  return (
    <OverlaySideDialog show={showDialog} onOpenChange={() => {}}>
      <YamlEditor
        item={item}
        onCloseButtonClick={() => {
          setShowDialog(false);
        }}
      />
    </OverlaySideDialog>
  );
};

export default YamlEditorOverlay;
