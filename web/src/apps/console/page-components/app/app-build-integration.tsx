import { useOutletContext } from '@remix-run/react';
import { Checkbox } from '@kloudlite/design-system/atoms/checkbox';
import { TextArea, TextInput } from '@kloudlite/design-system/atoms/input';
import Git from '~/console/components/git';
import KeyValuePair from '~/console/components/key-value-pair';
import { IGIT_PROVIDERS } from '~/console/hooks/use-git';
import { IAppContext } from '~/console/routes/_main+/$account+/env+/$environment+/app+/$app+/_layout';
import { dummyEvent } from '~/root/lib/client/hooks/use-form';

const AppBuildIntegration = ({
  values,
  errors,
  handleChange,
}: {
  values: Record<string, any>;
  errors: Record<string, any>;
  handleChange: (key: string) => (e: { target: { value: any } }) => void;
}) => {
  const { logins, loginUrls } = useOutletContext<IAppContext>();

  return (
    <div className="flex flex-col gap-3xl">
      <Git
        logins={logins}
        loginUrls={loginUrls}
        error={errors?.source}
        onChange={(git) => {
          handleChange('source')(
            dummyEvent({
              branch: git.branch,
              repository: git.repo,
              provider: git.provider,
            })
          );
        }}
        value={{
          branch: values.source.branch,
          repo: values.source.repository,
          provider: (values.source.provider as IGIT_PROVIDERS) || 'github',
        }}
      />

      <Checkbox
        label="Advance options"
        checked={values.advanceOptions}
        onChange={(check) => {
          handleChange('advanceOptions')(dummyEvent(!!check));
        }}
      />
      {values.advanceOptions && (
        <div className="flex flex-col gap-3xl">
          <KeyValuePair
            size="lg"
            label="Build args"
            value={Object.entries(values.buildArgs || {}).map(
              ([key, value]) => ({ key, value })
            )}
            onChange={(_, items) => {
              handleChange('buildArgs')(dummyEvent(items));
            }}
            error={!!errors.buildArgs}
            message={errors.buildArgs}
          />
          <KeyValuePair
            size="lg"
            label="Build contexts"
            value={Object.entries(values.buildContexts || {}).map(
              ([key, value]) => ({ key, value })
            )}
            onChange={(_, items) => {
              handleChange('buildContexts')(dummyEvent(items));
            }}
            error={!!errors.buildContexts}
            message={errors.buildContexts}
          />
          <TextInput
            size="lg"
            placeholder="Enter context dir"
            label="Context dir"
            value={values.contextDir}
            onChange={handleChange('contextDir')}
          />
          <TextInput
            size="lg"
            placeholder="Enter docker file path"
            label="Docker file path"
            value={values.dockerfilePath}
            onChange={handleChange('dockerfilePath')}
          />
          <TextArea
            placeholder="Enter docker file content"
            label="Docker file content"
            value={values.dockerfileContent}
            onChange={handleChange('dockerfileContent')}
            resize={false}
            rows="6"
          />
        </div>
      )}
    </div>
  );
};

export default AppBuildIntegration;
