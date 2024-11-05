import OverlaySideDialog from './overlay-side-dialog';
import YamlEditor from '../page-components/yaml-editor/yaml-editor';

const YamlEditorOverlay = ({
  item,
  showDialog,
  setShowDialog,
  onCommit,
}: {
  item: any;
  showDialog: boolean;
  setShowDialog: React.Dispatch<React.SetStateAction<boolean>>;
  onCommit: ({ spec }: { spec: any }) => Promise<boolean>;
}) => {
  return (
    <OverlaySideDialog show={showDialog} onOpenChange={() => {}}>
      <YamlEditor
        item={item}
        onCloseButtonClick={() => {
          setShowDialog(false);
        }}
        onCommit={onCommit}
      />
    </OverlaySideDialog>
  );
};

export default YamlEditorOverlay;
