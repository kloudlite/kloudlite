import CodeEditorClient from '~/root/lib/client/components/editor-client';
import { Button } from '~/components/atoms/button';
import { useCallback, useEffect, useState } from 'react';
import { yamlDump, yamlParse } from '~/console/components/diff-viewer';
import { Box } from '~/console/components/common-console-components';
import { validateType } from '~/root/src/generated/gql/validator';

const YamlEditor = ({
  item,
  onCommit = () => {},
}: {
  item: any;
  onCommit?: () => Promise<null>;
}) => {
  const ymlSpec = useCallback(() => {
    if (!item || !item.spec) {
      return '';
    }

    return yamlDump(item.spec);
  }, [item])();

  const [initialState, setState] = useState(ymlSpec);
  const [updated, setUpdated] = useState(false);

  const [loading] = useState(false);

  useEffect(() => {
    setUpdated(initialState !== ymlSpec);
  }, [initialState]);

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

  return (
    <Box title="Edit App As Yaml">
      <CodeEditorClient
        height="50rem"
        value={ymlSpec}
        onChange={(v) => {
          if (v) {
            setState(v);
          }
        }}
        lang="yaml"
      />
      <pre className="border border-border-critical transition-all min-h-4xl text-text-critical">
        {errors.map((r) => {
          return <pre key={r}>{r}</pre>;
        })}
      </pre>

      <pre>{initialState}</pre>
      <div className="flex justify-end">
        <Button
          disabled={!updated}
          content={loading ? 'Committing changes.' : 'Commit Changes'}
          // loading={loading}
          // onClick={() => setPerformAction('view-changes')}
        />
      </div>
    </Box>
  );
};

export default YamlEditor;
