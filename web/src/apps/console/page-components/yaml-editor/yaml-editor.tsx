import CodeEditorClient from '~/root/lib/client/components/editor-client';
import { Button, IconButton } from '~/components/atoms/button';
import { useCallback, useEffect, useState } from 'react';
import { yamlDump, yamlParse } from '~/console/components/diff-viewer';
import { Box } from '~/console/components/common-console-components';
import { validateType } from '~/root/src/generated/gql/validator';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import Popup from '~/components/molecule/popup';
import { X } from '~/console/components/icons';

const YamlEditor = ({
  item,
  onCloseButtonClick,
  onCommit = () => {},
}: {
  item: any;
  onCloseButtonClick?: () => void;
  onCommit?: () => Promise<null>;
}) => {
  const { setHasChanges, hasChanges } = useUnsavedChanges();
  const ymlSpec = useCallback(() => {
    if (!item || !item.spec) {
      return '';
    }

    return yamlDump(item.spec);
  }, [item])();

  const [initialState, setState] = useState(ymlSpec);
  const [loading] = useState(false);

  const [closeEditor, setcloseEditor] = useState(false);

  const [errors, setErrors] = useState<string[]>([]);

  useEffect(() => {
    const { data: spec, error } = yamlParse(initialState);
    if (error) {
      setErrors([error]);

      return;
    }

    try {
      // console.log(spec);

      const res = validateType(
        {
          ...item,
          spec,
        },
        'AppIn'
      );

      setErrors(res);
    } catch (e) {
      const er = e as Error;
      setErrors([er.message]);
    }
  }, [initialState]);

  useEffect(() => {
    setHasChanges(initialState !== ymlSpec);
  }, [initialState]);

  return (
    <Box
      className="!shadow-none h-full !border-none"
      title={
        <div className="flex justify-between">
          <span>Edit App As Yaml</span>
          <IconButton
            variant="plain"
            onClick={() => {
              if (hasChanges) {
                setcloseEditor(true);
              } else {
                onCloseButtonClick?.();
              }
            }}
            icon={<X />}
          />
        </div>
      }
    >
      <div className="flex gap-lg justify-end">
        <Button
          disabled={!hasChanges}
          content="Discard Changes"
          variant="outline"
          onClick={() => {
            setState(ymlSpec);
          }}
        />
        <Button
          disabled={!hasChanges}
          content={loading ? 'Committing changes.' : 'Commit Changes'}
          // loading={loading}
          // onClick={() => setPerformAction('view-changes')}
        />
      </div>
      <div className="h-full">
        <div className="mb-[2px]">
          <CodeEditorClient
            height="calc(70vh - 132px - 2px - 20px)" ///
            value={initialState}
            onChange={(v) => {
              if (v) {
                setState(v);
              }
            }}
            lang="yaml"
          />
        </div>
        <div className="px-[54px] py-[54px] overflow-y-auto h-[30vh] border border-border-critical transition-all">
          <pre className=" text-text-critical">
            {errors.map((r) => {
              return <pre key={r}>{r}</pre>;
            })}
          </pre>
        </div>
      </div>
      {/* <CodeEditorClient
        height="50vh"
        value={initialState}
        onChange={(v) => {
          if (v) {
            setState(v);
          }
        }}
        lang="yaml"
      />
      <div className="overflow-y-auto h-[30vh] border border-border-critical transition-all">
        <pre className=" text-text-critical">
          {errors.map((r) => {
            return <pre key={r}>{r}</pre>;
          })}
        </pre>
      </div> */}

      <Popup.Root
        show={hasChanges && closeEditor}
        onOpenChange={() => {
          setcloseEditor(false);
        }}
        backdrop={false}
      >
        <Popup.Header>Unsaved changes</Popup.Header>
        <Popup.Content>
          Are you sure you want to discard the changes?
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button content="Cancel" variant="basic" closable />
          <Popup.Button
            content="Discard"
            variant="warning"
            onClick={() => {
              //
              setState(ymlSpec);
              onCloseButtonClick?.();
            }}
          />
        </Popup.Footer>
      </Popup.Root>
    </Box>
  );
};

export default YamlEditor;
