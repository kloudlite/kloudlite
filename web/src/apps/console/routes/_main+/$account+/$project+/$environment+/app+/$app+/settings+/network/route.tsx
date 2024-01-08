import Wrapper from '~/console/components/wrapper';
import { Button } from '~/components/atoms/button';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { ExposedPorts } from '../../../../new-app/app-network';

const AppNetwork = () => {
  const { setPerformAction, hasChanges, loading } = useUnsavedChanges();

  return (
    <div>
      <Wrapper
        secondaryHeader={{
          title: 'Network',
          action: hasChanges && (
            <div className="flex flex-row items-center gap-lg">
              <Button
                disabled={loading}
                variant="basic"
                content="Discard changes"
                onClick={() => setPerformAction('discard-changes')}
              />
              <Button
                disabled={loading}
                content={loading ? 'Committing changes.' : 'View changes'}
                loading={loading}
                onClick={() => setPerformAction('view-changes')}
              />
            </div>
          ),
        }}
      >
        <ExposedPorts />
      </Wrapper>
    </div>
  );
};

export default AppNetwork;
